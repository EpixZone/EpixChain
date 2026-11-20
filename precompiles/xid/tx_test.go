package xid_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/precompiles/xid"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/cosmos/evm/x/xid/keeper"
	"github.com/cosmos/evm/x/xid/types"

	"cosmossdk.io/math"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// noopBank accepts every fee transfer and burn so Register can run without a
// real bank module.
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

// logStateDB records EVM logs; no other StateDB method is used by the xID
// precompile's write methods.
type logStateDB struct {
	vm.StateDB
	logs []*ethtypes.Log
}

func (s *logStateDB) AddLog(l *ethtypes.Log) { s.logs = append(s.logs, l) }

type TxTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	keeper  keeper.Keeper
	p       *xid.Precompile
	stateDB *logStateDB
	caller  common.Address
	owner   sdk.AccAddress
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(TxTestSuite))
}

func (s *TxTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx
	s.keeper = keeper.NewKeeper(encCfg.Codec, key, authtypes.NewModuleAddress("gov").String(), noopBank{}, nil, nil)
	s.p = xid.NewPrecompile(s.keeper, nil)
	s.stateDB = &logStateDB{}
	s.caller = common.HexToAddress("0x1111111111111111111111111111111111111111")
	s.owner = sdk.AccAddress(s.caller.Bytes())

	s.keeper.SetTLDConfig(s.ctx, types.TLDConfig{
		Tld:        "epix",
		Enabled:    true,
		PriceTiers: []types.PriceTier{{MaxLength: 64, Price: math.NewInt(1)}},
	})
	// "alice.epix" is registered through the precompile with a mixed-case name,
	// which RegisterNameRecord stores as "alice".
	_, err := s.call(xid.RegisterMethod, "Alice", "EPIX")
	s.Require().NoError(err)
	_, found := s.keeper.GetNameRecord(s.ctx, "epix", "alice")
	s.Require().True(found, "register must store the lowercase name")
}

func (s *TxTestSuite) contract() *vm.Contract {
	return vm.NewContract(s.caller, common.HexToAddress(evmtypes.XIDPrecompileAddress), uint256.NewInt(0), 10_000_000, nil)
}

// call runs a precompile write method as s.caller.
func (s *TxTestSuite) call(methodName string, args ...interface{}) ([]byte, error) {
	method := s.p.Methods[methodName]
	contract := s.contract()
	switch methodName {
	case xid.RegisterMethod:
		return s.p.Register(s.ctx, contract, s.stateDB, &method, args)
	case xid.TransferNameMethod:
		return s.p.TransferName(s.ctx, contract, s.stateDB, &method, args)
	case xid.UpdateProfileMethod:
		return s.p.UpdateProfile(s.ctx, contract, s.stateDB, &method, args)
	case xid.SetDNSRecordMethod:
		return s.p.SetDNSRecord(s.ctx, contract, s.stateDB, &method, args)
	case xid.DeleteDNSRecordMethod:
		return s.p.DeleteDNSRecord(s.ctx, contract, s.stateDB, &method, args)
	case xid.LinkIdentityMethod:
		return s.p.LinkIdentity(s.ctx, contract, s.stateDB, &method, args)
	case xid.UnlinkIdentityMethod:
		return s.p.UnlinkIdentity(s.ctx, contract, s.stateDB, &method, args)
	case xid.AttestStateDigestMethod:
		return s.p.AttestStateDigest(s.ctx, contract, s.stateDB, &method, args)
	case xid.SetPrimaryNameMethod:
		return s.p.SetPrimaryName(s.ctx, contract, s.stateDB, &method, args)
	}
	s.FailNow("unknown method " + methodName)
	return nil, nil
}

// --- Validation parity with the Cosmos Msg ValidateBasic -------------------

func (s *TxTestSuite) TestUpdateProfile_RejectsOversizeFields() {
	_, err := s.call(xid.UpdateProfileMethod, "alice", "epix", strings.Repeat("a", types.MaxAvatarLength+1), "bio")
	s.Require().ErrorIs(err, types.ErrInvalidProfile)

	_, err = s.call(xid.UpdateProfileMethod, "alice", "epix", "avatar", strings.Repeat("b", types.MaxBioLength+1))
	s.Require().ErrorIs(err, types.ErrInvalidProfile)

	_, found := s.keeper.GetProfileRecord(s.ctx, "epix", "alice")
	s.Require().False(found, "a rejected profile must not be stored")

	// At the limits it is accepted.
	_, err = s.call(xid.UpdateProfileMethod, "alice", "epix", strings.Repeat("a", types.MaxAvatarLength), strings.Repeat("b", types.MaxBioLength))
	s.Require().NoError(err)
}

func (s *TxTestSuite) TestSetDNSRecord_RejectsInvalidRecords() {
	cases := []struct {
		name       string
		recordType uint16
		value      string
		ttl        uint32
	}{
		{"unsupported type", 99, "1.2.3.4", 300},
		{"type zero", 0, "1.2.3.4", 300},
		{"empty value", uint16(types.DNSRecordTypeA), "", 300},
		{"oversize value", uint16(types.DNSRecordTypeTXT), strings.Repeat("x", types.MaxDNSValueLength+1), 300},
		{"ttl below minimum", uint16(types.DNSRecordTypeA), "1.2.3.4", types.MinDNSTTL - 1},
		{"ttl above maximum", uint16(types.DNSRecordTypeA), "1.2.3.4", types.MaxDNSTTL + 1},
	}
	for _, tc := range cases {
		_, err := s.call(xid.SetDNSRecordMethod, "alice", "epix", tc.recordType, tc.value, tc.ttl)
		s.Require().ErrorIs(err, types.ErrInvalidDNSRecord, tc.name)
	}
	s.Require().Empty(s.keeper.GetAllDNSRecords(s.ctx, "epix", "alice"), "rejected records must not be stored")

	_, err := s.call(xid.SetDNSRecordMethod, "alice", "epix", uint16(types.DNSRecordTypeA), "1.2.3.4", uint32(0))
	s.Require().NoError(err)
	s.Require().Len(s.keeper.GetAllDNSRecords(s.ctx, "epix", "alice"), 1)
}

func (s *TxTestSuite) TestDeleteDNSRecord_RequiresExistingRecord() {
	_, err := s.call(xid.DeleteDNSRecordMethod, "alice", "epix", uint16(types.DNSRecordTypeA))
	s.Require().ErrorContains(err, "not found")
	s.Require().Empty(s.stateDB.logs[1:], "no DNSRecordDeleted log for a missing record")

	_, err = s.call(xid.DeleteDNSRecordMethod, "alice", "epix", uint16(0))
	s.Require().ErrorIs(err, types.ErrInvalidDNSRecord)
}

func (s *TxTestSuite) TestTransferName_RejectsZeroAddressAndSelf() {
	_, err := s.call(xid.TransferNameMethod, "alice", "epix", common.Address{})
	s.Require().ErrorContains(err, "zero address")

	_, err = s.call(xid.TransferNameMethod, "alice", "epix", s.caller)
	s.Require().ErrorContains(err, "yourself")

	record, _ := s.keeper.GetNameRecord(s.ctx, "epix", "alice")
	s.Require().Equal(s.owner.String(), record.Owner)
}

func (s *TxTestSuite) TestLinkIdentity_RejectsEmptyAddress() {
	_, err := s.call(xid.LinkIdentityMethod, "alice", "epix", "", "label")
	s.Require().ErrorIs(err, types.ErrInvalidIdentity)
	s.Require().Empty(s.keeper.GetAllLinkedIdentities(s.ctx, "epix", "alice"))
}

func (s *TxTestSuite) TestUnlinkIdentity_RejectsAlreadyUnlinked() {
	_, err := s.call(xid.LinkIdentityMethod, "alice", "epix", "epix1identity", "label")
	s.Require().NoError(err)
	_, err = s.call(xid.UnlinkIdentityMethod, "alice", "epix", "epix1identity")
	s.Require().NoError(err)
	first, _ := s.keeper.GetLinkedIdentityEntry(s.ctx, "epix", "alice", "epix1identity")

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 10)
	_, err = s.call(xid.UnlinkIdentityMethod, "alice", "epix", "epix1identity")
	s.Require().ErrorContains(err, "already unlinked")

	again, _ := s.keeper.GetLinkedIdentityEntry(s.ctx, "epix", "alice", "epix1identity")
	s.Require().Equal(first.RevokedAt, again.RevokedAt, "a second unlink must not rewrite RevokedAt")

	_, err = s.call(xid.UnlinkIdentityMethod, "alice", "epix", "")
	s.Require().ErrorIs(err, types.ErrInvalidIdentity)
}

func (s *TxTestSuite) TestAttestStateDigest_RejectsMalformedInput() {
	_, err := s.call(xid.AttestStateDigestMethod, strings.Repeat("ab", 32), "")
	s.Require().ErrorIs(err, types.ErrInvalidAttestation)

	_, err = s.call(xid.AttestStateDigestMethod, "not-hex", "sig")
	s.Require().ErrorIs(err, types.ErrInvalidAttestation)
}

func (s *TxTestSuite) TestRegister_RejectsInvalidName() {
	_, err := s.call(xid.RegisterMethod, "-bad-", "epix")
	s.Require().ErrorIs(err, types.ErrInvalidName)
}

// --- Name case normalization ----------------------------------------------

func (s *TxTestSuite) TestWriteMethods_NormalizeNameCase() {
	_, err := s.call(xid.UpdateProfileMethod, "Alice", "EPIX", "avatar", "bio")
	s.Require().NoError(err)
	profile, found := s.keeper.GetProfileRecord(s.ctx, "epix", "alice")
	s.Require().True(found)
	s.Require().Equal("bio", profile.Bio)

	_, err = s.call(xid.SetDNSRecordMethod, "ALICE", "Epix", uint16(types.DNSRecordTypeA), "1.2.3.4", uint32(300))
	s.Require().NoError(err)
	_, found = s.keeper.GetDNSRecordEntry(s.ctx, "epix", "alice", types.DNSRecordTypeA)
	s.Require().True(found)

	_, err = s.call(xid.DeleteDNSRecordMethod, "aLiCe", "epix", uint16(types.DNSRecordTypeA))
	s.Require().NoError(err)
	_, found = s.keeper.GetDNSRecordEntry(s.ctx, "epix", "alice", types.DNSRecordTypeA)
	s.Require().False(found)

	_, err = s.call(xid.LinkIdentityMethod, "Alice", "epix", "epix1identity", "label")
	s.Require().NoError(err)
	_, err = s.call(xid.UnlinkIdentityMethod, "Alice", "epix", "epix1identity")
	s.Require().NoError(err)

	_, err = s.call(xid.SetPrimaryNameMethod, "Alice", "EPIX")
	s.Require().NoError(err)
	tld, name, found := s.keeper.GetPrimaryNameEntry(s.ctx, s.owner)
	s.Require().True(found)
	s.Require().Equal("epix", tld)
	s.Require().Equal("alice", name)

	newOwner := common.HexToAddress("0x2222222222222222222222222222222222222222")
	_, err = s.call(xid.TransferNameMethod, "Alice", "EPIX", newOwner)
	s.Require().NoError(err)
	record, _ := s.keeper.GetNameRecord(s.ctx, "epix", "alice")
	s.Require().Equal(sdk.AccAddress(newOwner.Bytes()).String(), record.Owner)
}

func (s *TxTestSuite) TestQueries_NormalizeNameCase() {
	method := s.p.Methods[xid.ResolveMethod]
	bz, err := s.p.Resolve(s.ctx, &method, []interface{}{"Alice", "EPIX"})
	s.Require().NoError(err)
	out, err := method.Outputs.Unpack(bz)
	s.Require().NoError(err)
	s.Require().Equal(s.caller, out[0].(common.Address))
}

func (s *TxTestSuite) TestRegister_EmitsNormalizedName() {
	// SetupTest registered "Alice"."EPIX"; the NameRegistered log must carry
	// the stored (lowercase) name and TLD.
	s.Require().NotEmpty(s.stateDB.logs)
	event := s.p.Events[xid.EventNameRegistered]
	vals, err := event.Inputs.NonIndexed().Unpack(s.stateDB.logs[0].Data)
	s.Require().NoError(err)
	s.Require().Equal("alice", vals[0])
	s.Require().Equal("epix", vals[1])
}
