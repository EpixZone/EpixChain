package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/x/epixmint"
	"github.com/cosmos/evm/x/epixmint/keeper"
	"github.com/cosmos/evm/x/epixmint/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx                sdk.Context
	keeper             keeper.Keeper
	bankKeeper         *MockBankKeeper
	accountKeeper      *MockAccountKeeper
	distributionKeeper *MockDistributionKeeper
	stakingKeeper      *MockStakingKeeper
	cdc                codec.BinaryCodec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(epixmint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	s.cdc = encCfg.Codec
	s.bankKeeper = &MockBankKeeper{}
	s.accountKeeper = &MockAccountKeeper{}
	s.distributionKeeper = &MockDistributionKeeper{}
	// Default to a single bonded validator so the normal staking-reward path is exercised
	s.stakingKeeper = &MockStakingKeeper{validators: []stakingtypes.Validator{bondedValidator("val1", 100)}}

	s.keeper = keeper.NewKeeper(
		s.cdc,
		key,
		s.bankKeeper,
		s.accountKeeper,
		s.distributionKeeper,
		s.stakingKeeper,
		authtypes.NewModuleAddress("gov").String(),
	)
}

// bondedValidator builds a bonded validator with the given operator name and token amount.
func bondedValidator(name string, tokens int64) stakingtypes.Validator {
	return stakingtypes.Validator{
		OperatorAddress: sdk.ValAddress([]byte(name)).String(),
		Status:          stakingtypes.Bonded,
		Tokens:          math.NewInt(tokens),
	}
}

func (s *KeeperTestSuite) TestGetSetParams() {
	params := types.DefaultParams()

	// Test setting params
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Test getting params
	retrievedParams := s.keeper.GetParams(s.ctx)
	s.Require().Equal(params, retrievedParams)
}

func (s *KeeperTestSuite) TestMintCoins() {
	// Set up default params
	params := types.DefaultParams()
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Mock current supply (below max)
	currentSupply, _ := math.NewIntFromString("1000000000000000000000000000") // 1B EPIX in aepix
	s.bankKeeper.SetSupply(params.MintDenom, currentSupply)

	// Test minting
	err = s.keeper.MintCoins(s.ctx)
	s.Require().NoError(err)

	// Verify mint was called
	s.Require().True(s.bankKeeper.MintCalled)
	s.Require().True(s.bankKeeper.SendCalled)
}

func (s *KeeperTestSuite) TestMintCoins_MaxSupplyReached() {
	// Set up default params
	params := types.DefaultParams()
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Mock current supply at max
	s.bankKeeper.SetSupply(params.MintDenom, params.MaxSupply)

	// Test minting when max supply is reached
	err = s.keeper.MintCoins(s.ctx)
	s.Require().NoError(err)

	// Verify no mint was called
	s.Require().False(s.bankKeeper.MintCalled)
	s.Require().False(s.bankKeeper.SendCalled)
}

func (s *KeeperTestSuite) TestMintCoins_NearMaxSupply() {
	// Set up default params
	params := types.DefaultParams()
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Calculate tokens per block using dynamic emission (at genesis, it's the initial amount)
	secondsPerYear := uint64(365 * 24 * 60 * 60)
	blocksPerYear := secondsPerYear / params.BlockTimeSeconds
	tokensPerBlock := params.InitialAnnualMintAmount.Quo(math.NewIntFromUint64(blocksPerYear))

	// Mock current supply very close to max (less than one block's worth)
	nearMaxSupply := params.MaxSupply.Sub(tokensPerBlock.QuoRaw(2)) // Half a block's worth below max
	s.bankKeeper.SetSupply(params.MintDenom, nearMaxSupply)

	// Test minting
	err = s.keeper.MintCoins(s.ctx)
	s.Require().NoError(err)

	// Verify mint was called with reduced amount
	s.Require().True(s.bankKeeper.MintCalled)
	s.Require().True(s.bankKeeper.SendCalled)

	// Verify the minted amount was capped
	expectedMintAmount := params.MaxSupply.Sub(nearMaxSupply)
	s.Require().Equal(expectedMintAmount, s.bankKeeper.LastMintedAmount)
}

func (s *KeeperTestSuite) TestMintCoins_NoBondedValidators() {
	params := types.DefaultParams()
	s.Require().NoError(s.keeper.SetParams(s.ctx, params))

	currentSupply, _ := math.NewIntFromString("1000000000000000000000000000") // 1B EPIX in aepix
	s.bankKeeper.SetSupply(params.MintDenom, currentSupply)

	// No bonded validators at all
	s.stakingKeeper.validators = nil

	s.Require().NoError(s.keeper.MintCoins(s.ctx))
	s.Require().True(s.bankKeeper.MintCalled)

	// Nothing must be moved to the distribution module: the epixmint module
	// account has to keep the coins so it can fund the community pool itself.
	s.Require().False(s.bankKeeper.SendCalled)
	s.Require().Empty(s.distributionKeeper.AllocateCalls)

	// Both the community share and the redirected staking share are funded
	// from the epixmint module account and add up to the full minted amount.
	epixmintAddr := authtypes.NewModuleAddress(types.ModuleName)
	s.Require().Len(s.distributionKeeper.FundCalls, 2)
	funded := math.ZeroInt()
	for _, call := range s.distributionKeeper.FundCalls {
		s.Require().Equal(epixmintAddr, call.Sender)
		funded = funded.Add(call.Amount.AmountOf(params.MintDenom))
	}
	s.Require().Equal(s.bankKeeper.LastMintedAmount, funded)
}

func (s *KeeperTestSuite) TestMintCoins_BondedValidatorsWithZeroPower() {
	params := types.DefaultParams()
	s.Require().NoError(s.keeper.SetParams(s.ctx, params))

	currentSupply, _ := math.NewIntFromString("1000000000000000000000000000")
	s.bankKeeper.SetSupply(params.MintDenom, currentSupply)

	// Bonded but with zero tokens: total voting power is zero, so the staking
	// share must also be redirected to the community pool.
	s.stakingKeeper.validators = []stakingtypes.Validator{bondedValidator("val1", 0)}

	s.Require().NoError(s.keeper.MintCoins(s.ctx))
	s.Require().False(s.bankKeeper.SendCalled)
	s.Require().Empty(s.distributionKeeper.AllocateCalls)
	s.Require().Len(s.distributionKeeper.FundCalls, 2)
}

func (s *KeeperTestSuite) TestMintCoins_ProportionalAllocation() {
	params := types.DefaultParams()
	s.Require().NoError(s.keeper.SetParams(s.ctx, params))

	currentSupply, _ := math.NewIntFromString("1000000000000000000000000000")
	s.bankKeeper.SetSupply(params.MintDenom, currentSupply)

	s.stakingKeeper.validators = []stakingtypes.Validator{
		bondedValidator("val1", 300),
		bondedValidator("val2", 100),
		{OperatorAddress: sdk.ValAddress([]byte("unbonded")).String(), Status: stakingtypes.Unbonded, Tokens: math.NewInt(1000)},
	}

	s.Require().NoError(s.keeper.MintCoins(s.ctx))
	s.Require().True(s.bankKeeper.MintCalled)
	s.Require().True(s.bankKeeper.SendCalled)

	// Only the community share is funded directly
	s.Require().Len(s.distributionKeeper.FundCalls, 1)
	expectedCommunity := params.CommunityPoolRate.MulInt(s.bankKeeper.LastMintedAmount).TruncateInt()
	s.Require().Equal(expectedCommunity, s.distributionKeeper.FundCalls[0].Amount.AmountOf(params.MintDenom))

	// Only bonded validators receive allocations, proportional to voting power
	stakingShare := s.bankKeeper.LastSentAmount
	s.Require().Equal(s.bankKeeper.LastMintedAmount.Sub(expectedCommunity), stakingShare)
	s.Require().Len(s.distributionKeeper.AllocateCalls, 2)
	stakingShareDec := math.LegacyNewDecFromInt(stakingShare)
	s.Require().Equal(stakingShareDec.MulInt64(3).QuoInt64(4), s.distributionKeeper.AllocateCalls[0].Tokens.AmountOf(params.MintDenom))
	s.Require().Equal(stakingShareDec.QuoInt64(4), s.distributionKeeper.AllocateCalls[1].Tokens.AmountOf(params.MintDenom))
}

// Mock implementations
type MockBankKeeper struct {
	supply           map[string]math.Int
	MintCalled       bool
	SendCalled       bool
	LastMintedAmount math.Int
	LastSentAmount   math.Int
}

func (m *MockBankKeeper) GetSupply(ctx context.Context, denom string) sdk.Coin {
	if amount, exists := m.supply[denom]; exists {
		return sdk.NewCoin(denom, amount)
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

func (m *MockBankKeeper) SetSupply(denom string, amount math.Int) {
	if m.supply == nil {
		m.supply = make(map[string]math.Int)
	}
	m.supply[denom] = amount
}

func (m *MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	m.MintCalled = true
	if len(amt) > 0 {
		m.LastMintedAmount = amt[0].Amount
	}
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	m.SendCalled = true
	if len(amt) > 0 {
		m.LastSentAmount = amt[0].Amount
	}
	return nil
}

type MockAccountKeeper struct{}

func (m *MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}

func (m *MockAccountKeeper) GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI {
	return nil
}

// FundCall records a FundCommunityPool invocation.
type FundCall struct {
	Amount sdk.Coins
	Sender sdk.AccAddress
}

// AllocateCall records an AllocateTokensToValidator invocation.
type AllocateCall struct {
	Validator stakingtypes.ValidatorI
	Tokens    sdk.DecCoins
}

// MockDistributionKeeper implements the DistributionKeeper interface for testing
type MockDistributionKeeper struct {
	FundCalls     []FundCall
	AllocateCalls []AllocateCall
}

func (m *MockDistributionKeeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	m.FundCalls = append(m.FundCalls, FundCall{Amount: amount, Sender: sender})
	return nil
}

func (m *MockDistributionKeeper) AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error {
	m.AllocateCalls = append(m.AllocateCalls, AllocateCall{Validator: val, Tokens: tokens})
	return nil
}

// MockStakingKeeper implements the StakingKeeper interface for testing
type MockStakingKeeper struct {
	validators []stakingtypes.Validator
}

func (m *MockStakingKeeper) GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error) {
	return m.validators, nil
}

func (m *MockStakingKeeper) BondedRatio(ctx context.Context) (ratio math.LegacyDec, err error) {
	return math.LegacyNewDec(1), nil
}
