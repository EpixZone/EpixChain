package keeper_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/x/xid/keeper"
	"github.com/cosmos/evm/x/xid/types"

	"cosmossdk.io/math"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// noopBank accepts every fee transfer and burn so RegisterName can run without
// a real bank module.
type noopBank struct{}

func (noopBank) SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	return nil
}
func (noopBank) BurnCoins(context.Context, string, sdk.Coins) error { return nil }
func (noopBank) GetBalance(context.Context, sdk.AccAddress, string) sdk.Coin {
	return sdk.Coin{}
}
func (noopBank) SpendableCoin(context.Context, sdk.AccAddress, string) sdk.Coin {
	return sdk.Coin{}
}
func (noopBank) SendCoins(context.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
	return nil
}

type MsgServerTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper
	alice  sdk.AccAddress
	bob    sdk.AccAddress
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx
	s.keeper = keeper.NewKeeper(encCfg.Codec, key, authtypes.NewModuleAddress("gov").String(), noopBank{}, nil, nil)
	s.alice = sdk.AccAddress([]byte("alice_______________"))
	s.bob = sdk.AccAddress([]byte("bob_________________"))

	s.keeper.SetTLDConfig(s.ctx, types.TLDConfig{
		Tld:        "epix",
		Enabled:    true,
		PriceTiers: []types.PriceTier{{MaxLength: 64, Price: math.NewInt(1)}},
	})

	// The first registration initializes the Merkle tree with a full rebuild,
	// so register a lowercase name first: every later write goes through the
	// incremental UpdateDomainInTree path that the case bug affected.
	_, err := s.keeper.RegisterName(s.ctx, &types.MsgRegisterName{Owner: s.alice.String(), Name: "first", Tld: "epix"})
	s.Require().NoError(err)
}

// requireProvable asserts that name.tld has a Merkle leaf whose hash matches the
// current record data, and that the proof root is the stored state digest.
func (s *MsgServerTestSuite) requireProvable(name, tld string) *types.QueryResolveWithProofResponse {
	res, err := s.keeper.ResolveWithProof(s.ctx, &types.QueryResolveWithProofRequest{Name: name, Tld: tld})
	s.Require().NoError(err)
	pre, err := hex.DecodeString(res.LeafPreimage)
	s.Require().NoError(err)
	sum := sha256.Sum256(pre)
	s.Require().Equal(hex.EncodeToString(sum[:]), res.Proof.LeafHash, "leaf hash must match the current record data")
	s.Require().Equal(res.Root, res.Proof.Root)
	return res
}

func (s *MsgServerTestSuite) TestRegisterName_MixedCaseUpdatesCanonicalLeaf() {
	_, err := s.keeper.RegisterName(s.ctx, &types.MsgRegisterName{Owner: s.alice.String(), Name: "Alice", Tld: "EPIX"})
	s.Require().NoError(err)

	_, found := s.keeper.GetLeafIndex(s.ctx, "epix", "alice")
	s.Require().True(found, "the stored lowercase name must get a Merkle leaf")
	_, found = s.keeper.GetLeafIndex(s.ctx, "EPIX", "Alice")
	s.Require().False(found, "no leaf may be created for the non-canonical spelling")

	s.requireProvable("alice", "epix")
	// Mixed-case queries resolve to the same record.
	s.requireProvable("ALICE", "Epix")
}

func (s *MsgServerTestSuite) TestTransferName_MixedCaseUpdatesCanonicalLeaf() {
	_, err := s.keeper.RegisterName(s.ctx, &types.MsgRegisterName{Owner: s.alice.String(), Name: "alice", Tld: "epix"})
	s.Require().NoError(err)

	_, err = s.keeper.TransferName(s.ctx, &types.MsgTransferName{Owner: s.alice.String(), NewOwner: s.bob.String(), Name: "Alice", Tld: "EPIX"})
	s.Require().NoError(err)

	res := s.requireProvable("alice", "epix")
	s.Require().Equal(s.bob.String(), res.Domain.Record.Owner)
}

func (s *MsgServerTestSuite) TestOwnerWrites_AcceptMixedCaseName() {
	_, err := s.keeper.RegisterName(s.ctx, &types.MsgRegisterName{Owner: s.alice.String(), Name: "alice", Tld: "epix"})
	s.Require().NoError(err)

	_, err = s.keeper.UpdateProfile(s.ctx, &types.MsgUpdateProfile{Owner: s.alice.String(), Name: "Alice", Tld: "EPIX", Profile: types.Profile{Bio: "hi"}})
	s.Require().NoError(err)
	profile, found := s.keeper.GetProfileRecord(s.ctx, "epix", "alice")
	s.Require().True(found)
	s.Require().Equal("hi", profile.Bio)

	record := types.DNSRecord{RecordType: types.DNSRecordTypeA, Value: "1.2.3.4", Ttl: 300}
	_, err = s.keeper.SetDNSRecord(s.ctx, &types.MsgSetDNSRecord{Owner: s.alice.String(), Name: "ALICE", Tld: "epix", Record: record})
	s.Require().NoError(err)
	_, found = s.keeper.GetDNSRecordEntry(s.ctx, "epix", "alice", types.DNSRecordTypeA)
	s.Require().True(found)

	_, err = s.keeper.DeleteDNSRecord(s.ctx, &types.MsgDeleteDNSRecord{Owner: s.alice.String(), Name: "Alice", Tld: "epix", RecordType: types.DNSRecordTypeA})
	s.Require().NoError(err)

	identity := types.LinkedIdentity{Address: "epix1identity", Label: "l"}
	_, err = s.keeper.LinkIdentity(s.ctx, &types.MsgLinkIdentity{Owner: s.alice.String(), Name: "Alice", Tld: "epix", Identity: identity})
	s.Require().NoError(err)
	_, err = s.keeper.UnlinkIdentity(s.ctx, &types.MsgUnlinkIdentity{Owner: s.alice.String(), Name: "Alice", Tld: "epix", Address: "epix1identity"})
	s.Require().NoError(err)

	_, err = s.keeper.SetPrimaryName(s.ctx, &types.MsgSetPrimaryName{Owner: s.alice.String(), Name: "Alice", Tld: "EPIX"})
	s.Require().NoError(err)
	tld, name, found := s.keeper.GetPrimaryNameEntry(s.ctx, s.alice)
	s.Require().True(found)
	s.Require().Equal("epix", tld)
	s.Require().Equal("alice", name)

	s.requireProvable("alice", "epix")
}

func (s *MsgServerTestSuite) TestCreateTLD_StoresLowercase() {
	_, err := s.keeper.CreateTLD(s.ctx, &types.MsgCreateTLD{
		Authority: s.keeper.GetAuthority(),
		TldConfig: types.TLDConfig{Tld: "XYZ", Enabled: true, PriceTiers: []types.PriceTier{{MaxLength: 64, Price: math.NewInt(1)}}},
	})
	s.Require().NoError(err)
	s.Require().True(s.keeper.HasTLDConfig(s.ctx, "xyz"))
	s.Require().False(s.keeper.HasTLDConfig(s.ctx, "XYZ"))

	// The new TLD is usable for registration, which always lowercases the TLD.
	_, err = s.keeper.RegisterName(s.ctx, &types.MsgRegisterName{Owner: s.alice.String(), Name: "alice", Tld: "XYZ"})
	s.Require().NoError(err)
}
