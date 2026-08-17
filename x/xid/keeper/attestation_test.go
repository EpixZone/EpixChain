package keeper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/x/xid/keeper"
	"github.com/cosmos/evm/x/xid/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AttestationTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
}

func TestAttestationTestSuite(t *testing.T) {
	suite.Run(t, new(AttestationTestSuite))
}

func (s *AttestationTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx
	s.cdc = encCfg.Codec
	// The attestation-store methods exercised here only touch the KV store, so the
	// bank/account/staking keepers can be nil.
	s.keeper = keeper.NewKeeper(s.cdc, key, authtypes.NewModuleAddress("gov").String(), nil, nil, nil)
}

// TestClearAttestationsForDigest_RemovesEverything locks in the invariant that a
// superseded digest is fully pruned: attestations, the count, AND the canonical
// block_time. The block_time deletion is the regression guard — it used to be
// orphaned, leaving one entry per historical digest forever.
func (s *AttestationTestSuite) TestClearAttestationsForDigest_RemovesEverything() {
	digest := strings.Repeat("ab", 32)

	s.keeper.SetAttestation(s.ctx, types.Attestation{Digest: digest, ValidatorAddr: "val1", Signature: "sig1", VotingPower: 100})
	s.keeper.IncrementAttestationCount(s.ctx, digest)
	s.keeper.SetAttestation(s.ctx, types.Attestation{Digest: digest, ValidatorAddr: "val2", Signature: "sig2", VotingPower: 100})
	s.keeper.IncrementAttestationCount(s.ctx, digest)
	s.keeper.SetDigestBlockTime(s.ctx, digest, 1786934336)

	// Preconditions: everything is present.
	s.Require().Len(s.keeper.GetAttestations(s.ctx, digest), 2)
	s.Require().Equal(uint64(2), s.keeper.GetAttestationCount(s.ctx, digest))
	bt, ok := s.keeper.GetDigestBlockTime(s.ctx, digest)
	s.Require().True(ok)
	s.Require().Equal(int64(1786934336), bt)

	s.keeper.ClearAttestationsForDigest(s.ctx, digest)

	// Postconditions: attestations, count, and block_time are all gone.
	s.Require().Empty(s.keeper.GetAttestations(s.ctx, digest))
	s.Require().Equal(uint64(0), s.keeper.GetAttestationCount(s.ctx, digest))
	_, ok = s.keeper.GetDigestBlockTime(s.ctx, digest)
	s.Require().False(ok, "DigestBlockTime must be pruned when a digest is cleared")
}

// TestClearAttestationsForDigest_LeavesOtherDigests ensures the clear is scoped
// to one digest and does not disturb attestations recorded for a different one.
func (s *AttestationTestSuite) TestClearAttestationsForDigest_LeavesOtherDigests() {
	stale := strings.Repeat("ab", 32)
	current := strings.Repeat("cd", 32)

	s.keeper.SetAttestation(s.ctx, types.Attestation{Digest: stale, ValidatorAddr: "val1", Signature: "s"})
	s.keeper.SetDigestBlockTime(s.ctx, stale, 111)
	s.keeper.SetAttestation(s.ctx, types.Attestation{Digest: current, ValidatorAddr: "val1", Signature: "s"})
	s.keeper.SetDigestBlockTime(s.ctx, current, 222)

	s.keeper.ClearAttestationsForDigest(s.ctx, stale)

	s.Require().Empty(s.keeper.GetAttestations(s.ctx, stale))
	_, ok := s.keeper.GetDigestBlockTime(s.ctx, stale)
	s.Require().False(ok)

	s.Require().Len(s.keeper.GetAttestations(s.ctx, current), 1)
	bt, ok := s.keeper.GetDigestBlockTime(s.ctx, current)
	s.Require().True(ok)
	s.Require().Equal(int64(222), bt)
}
