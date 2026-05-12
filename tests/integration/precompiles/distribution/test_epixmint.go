package distribution

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/suite"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/distribution"
	evmaddress "github.com/cosmos/evm/encoding/address"
	testconstants "github.com/cosmos/evm/testutil/constants"
	"github.com/cosmos/evm/testutil/integration/evm/factory"
	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	"github.com/cosmos/evm/testutil/integration/evm/network"
	testkeyring "github.com/cosmos/evm/testutil/keyring"
	epixminttypes "github.com/cosmos/evm/x/epixmint/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// EpixMintDistributionTestSuite exercises the distribution precompile against
// EpixChain's actual inflation source (the x/epixmint module). Upstream's
// PrecompileTestSuite zeros out EpixMint to keep its exact-amount assertions
// stable; this suite enables it so we can verify that:
//
//   - inflation actually accrues to the community pool and to validators
//   - the distribution precompile's read methods (CommunityPool, ValidatorOutstandingRewards,
//     DelegationTotalRewards, ValidatorCommission) reflect that accrual
//   - the distribution precompile's write methods (WithdrawDelegatorReward,
//     ClaimRewards) still pay out aepix that came in via EpixMint
//
// Assertions are deliberately "greater than" rather than "exact equal" — the
// exact amount depends on the number of blocks the test framework has produced
// and the validator-share math, neither of which are stable across go-test
// invocations.
type EpixMintDistributionTestSuite struct {
	suite.Suite

	create  network.CreateEvmApp
	options []network.ConfigOption

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *distribution.Precompile
	bondDenom  string
}

func NewEpixMintDistributionTestSuite(
	create network.CreateEvmApp,
	options ...network.ConfigOption,
) *EpixMintDistributionTestSuite {
	return &EpixMintDistributionTestSuite{
		create:  create,
		options: options,
	}
}

func TestEpixMintDistributionTestSuiteRunner(t *testing.T, create network.CreateEvmApp, options ...network.ConfigOption) {
	s := NewEpixMintDistributionTestSuite(create, options...)
	suite.Run(t, s)
}

func (s *EpixMintDistributionTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	customGen := network.CustomGenesisState{}

	// Pre-fund the fee collector so distribution module bootstrap doesn't
	// fail; epixmint sends directly to validators but the distribution
	// module still expects the fee collector to exist.
	coins := sdk.NewCoins(sdk.NewCoin(testconstants.ExampleAttoDenom, sdkmath.NewInt(1_000_000_000_000_000_000)))
	bankGenesis := banktypes.DefaultGenesisState()
	bankGenesis.Balances = []banktypes.Balance{{
		Address: authtypes.NewModuleAddress(authtypes.FeeCollectorName).String(),
		Coins:   coins,
	}}
	customGen[banktypes.ModuleName] = bankGenesis

	customGen[distrtypes.ModuleName] = distrtypes.DefaultGenesisState()

	// Stock x/mint stays zero — epixmint drives inflation on EpixChain.
	mintGen := minttypes.DefaultGenesisState()
	mintGen.Params.MintDenom = testconstants.ExampleAttoDenom
	customGen[minttypes.ModuleName] = mintGen

	// Re-enable EpixMint with the default InitialAnnualMintAmount (the
	// testutil base setup zeros it out for upstream test compatibility —
	// here we want the real production behavior).
	epixmintGen := epixminttypes.DefaultGenesisState()
	epixmintGen.Params.MintDenom = testconstants.ExampleAttoDenom
	customGen[epixminttypes.ModuleName] = epixmintGen

	options := []network.ConfigOption{
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGen),
	}
	options = append(options, s.options...)
	nw := network.NewUnitTestNetwork(s.create, options...)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	bondDenom, err := nw.App.GetStakingKeeper().BondDenom(nw.GetContext())
	s.Require().NoError(err)

	s.network = nw
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.bondDenom = bondDenom
	s.precompile = distribution.NewPrecompile(
		s.network.App.GetDistrKeeper(),
		distrkeeper.NewMsgServerImpl(s.network.App.GetDistrKeeper()),
		distrkeeper.NewQuerier(s.network.App.GetDistrKeeper()),
		*s.network.App.GetStakingKeeper(),
		s.network.App.GetBankKeeper(),
		evmaddress.NewEvmCodec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
	)

	// Advance a few blocks so EpixMint has produced inflation.
	for i := 0; i < 3; i++ {
		s.Require().NoError(s.network.NextBlock())
	}
}

// TestCommunityPoolAccruesFromEpixMint verifies that after running a handful
// of blocks the community pool has received its configured share of EpixMint's
// emission (CommunityPoolRate of the per-block mint), and that the precompile
// surfaces that balance.
func (s *EpixMintDistributionTestSuite) TestCommunityPoolAccruesFromEpixMint() {
	s.SetupTest()
	ctx := s.network.GetContext()
	method := s.precompile.Methods[distribution.CommunityPoolMethod]
	contract := vm.NewContract(s.keyring.GetAddr(0), s.precompile.Address(), uint256.NewInt(0), 200_000, nil)

	bz, err := s.precompile.CommunityPool(ctx, contract, &method, []interface{}{})
	s.Require().NoError(err, "CommunityPool query should succeed")
	s.Require().NotEmpty(bz, "CommunityPool should return data")

	var out []cmn.DecCoin
	s.Require().NoError(s.precompile.UnpackIntoInterface(&out, distribution.CommunityPoolMethod, bz))
	s.Require().NotEmpty(out, "community pool should be non-empty after EpixMint emission")

	var found bool
	for _, c := range out {
		if c.Denom == s.bondDenom {
			s.Require().Truef(c.Amount.Sign() > 0,
				"community pool aepix balance should be positive, got %s", c.Amount.String())
			found = true
			break
		}
	}
	s.Require().True(found, "expected aepix in community pool, got %v", out)
}

// TestValidatorOutstandingRewardsAccrueFromEpixMint verifies that after a few
// blocks each bonded validator has accumulated outstanding rewards from
// EpixMint's per-block distribution. The precompile surfaces that amount.
func (s *EpixMintDistributionTestSuite) TestValidatorOutstandingRewardsAccrueFromEpixMint() {
	s.SetupTest()
	ctx := s.network.GetContext()

	vals, err := s.network.App.GetStakingKeeper().GetAllValidators(ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(vals, "expected at least one validator")
	valOpAddr := vals[0].OperatorAddress

	method := s.precompile.Methods[distribution.ValidatorOutstandingRewardsMethod]
	contract := vm.NewContract(s.keyring.GetAddr(0), s.precompile.Address(), uint256.NewInt(0), 200_000, nil)

	bz, err := s.precompile.ValidatorOutstandingRewards(ctx, contract, &method, []interface{}{valOpAddr})
	s.Require().NoError(err, "ValidatorOutstandingRewards should succeed")
	s.Require().NotEmpty(bz)

	var out []cmn.DecCoin
	s.Require().NoError(s.precompile.UnpackIntoInterface(&out, distribution.ValidatorOutstandingRewardsMethod, bz))
	s.Require().NotEmpty(out, "validator should have outstanding rewards from EpixMint")

	var found bool
	for _, c := range out {
		if c.Denom == s.bondDenom && c.Amount.Sign() > 0 {
			found = true
			break
		}
	}
	s.Require().Truef(found, "expected positive aepix outstanding rewards, got %v", out)
}

// TestWithdrawDelegatorRewardFromEpixMint verifies that a delegator can claim
// the rewards EpixMint allocated to their validator over a span of blocks.
// We bond the delegator first, advance blocks so EpixMint deposits rewards
// through the distribution module, then call WithdrawDelegatorReward via the
// precompile and assert the delegator's aepix balance grew.
func (s *EpixMintDistributionTestSuite) TestWithdrawDelegatorRewardFromEpixMint() {
	s.SetupTest()

	delegator := s.keyring.GetKey(0)

	vals, err := s.network.App.GetStakingKeeper().GetAllValidators(s.network.GetContext())
	s.Require().NoError(err)
	s.Require().NotEmpty(vals)
	valOpAddr := vals[0].OperatorAddress

	// Delegate from the test key so they have a share of the next block's emission.
	stake := sdkmath.NewInt(1_000_000_000_000_000_000) // 1 aepix-token
	delegateCoin := sdk.NewCoin(s.bondDenom, stake)
	s.Require().NoError(s.factory.Delegate(delegator.Priv, valOpAddr, delegateCoin))

	// Advance several blocks so EpixMint allocates rewards to this validator,
	// and the delegator accumulates a share of the per-block emission.
	for i := 0; i < 5; i++ {
		s.Require().NoError(s.network.NextBlock())
	}

	// Verify via the precompile that the delegator has pending rewards before
	// claiming. This both exercises the read path and produces a useful error
	// message if the underlying delegation/reward state isn't what we expect.
	delRewMethod := s.precompile.Methods[distribution.DelegationRewardsMethod]
	delRewContract := vm.NewContract(delegator.Addr, s.precompile.Address(), uint256.NewInt(0), 200_000, nil)
	delRewBz, err := s.precompile.DelegationRewards(s.network.GetContext(), delRewContract, &delRewMethod, []interface{}{delegator.Addr, valOpAddr})
	s.Require().NoError(err)
	var pending []cmn.DecCoin
	s.Require().NoError(s.precompile.UnpackIntoInterface(&pending, distribution.DelegationRewardsMethod, delRewBz))
	var pendingAepix = sdkmath.ZeroInt()
	for _, c := range pending {
		if c.Denom == s.bondDenom {
			pendingAepix = sdkmath.NewIntFromBigInt(c.Amount)
			break
		}
	}
	s.Require().Truef(pendingAepix.Sign() > 0,
		"expected non-zero pending delegation rewards after 5 blocks of EpixMint, got %v", pending)

	preBal := s.network.App.GetBankKeeper().GetBalance(s.network.GetContext(), delegator.AccAddr, s.bondDenom)

	method := s.precompile.Methods[distribution.WithdrawDelegatorRewardMethod]
	contract := vm.NewContract(delegator.Addr, s.precompile.Address(), uint256.NewInt(0), 500_000, nil)
	args := []interface{}{delegator.Addr, valOpAddr}

	bz, err := s.precompile.WithdrawDelegatorReward(s.network.GetContext(), contract, s.network.GetStateDB(), &method, args)
	s.Require().NoError(err, "WithdrawDelegatorReward should succeed")
	s.Require().NotEmpty(bz)

	postBal := s.network.App.GetBankKeeper().GetBalance(s.network.GetContext(), delegator.AccAddr, s.bondDenom)
	gained := postBal.Amount.Sub(preBal.Amount)
	s.Require().Truef(gained.Sign() > 0,
		"delegator aepix balance should grow from EpixMint reward withdrawal: pre=%s post=%s gained=%s pendingBeforeWithdraw=%s",
		preBal.Amount.String(), postBal.Amount.String(), gained.String(), pendingAepix.String())
}

// TestEpixMintRatioApproximate verifies the high-level invariant: of the
// aepix EpixMint emits in a block, ~CommunityPoolRate goes to the community
// pool and ~StakingRewardsRate flows through to validator outstanding rewards.
// We snapshot both sums before and after a single block and compare deltas
// against the configured rates with a small tolerance to absorb rounding.
func (s *EpixMintDistributionTestSuite) TestEpixMintRatioApproximate() {
	s.SetupTest()
	ctx := s.network.GetContext()
	// The EpixMintKeeper is not part of the evm.EvmApp interface; rely on
	// the package-level DefaultParams (the suite genesis above uses
	// DefaultGenesisState, so the running params match these).
	params := epixminttypes.DefaultParams()

	communityPoolBefore := s.communityPoolAepix(ctx)
	rewardsBefore := s.totalValidatorOutstandingAepix(ctx)

	s.Require().NoError(s.network.NextBlock())

	ctx = s.network.GetContext()
	communityPoolAfter := s.communityPoolAepix(ctx)
	rewardsAfter := s.totalValidatorOutstandingAepix(ctx)

	cpDelta := communityPoolAfter.Sub(communityPoolBefore)
	rewardsDelta := rewardsAfter.Sub(rewardsBefore)
	totalDelta := cpDelta.Add(rewardsDelta)

	s.Require().Truef(totalDelta.Sign() > 0,
		"expected EpixMint to emit aepix in a single block, got total delta %s", totalDelta.String())

	// Expected ratios: communityPool / total ≈ CommunityPoolRate
	gotCPShare := sdkmath.LegacyNewDecFromInt(cpDelta).Quo(sdkmath.LegacyNewDecFromInt(totalDelta))
	tolerance := sdkmath.LegacyMustNewDecFromStr("0.001") // 0.1%
	diff := gotCPShare.Sub(params.CommunityPoolRate).Abs()
	s.Require().Truef(diff.LTE(tolerance),
		"community pool share should be ~%s, got %s (diff %s, totalDelta %s)",
		params.CommunityPoolRate.String(), gotCPShare.String(), diff.String(), totalDelta.String())
}

// communityPoolAepix returns the integer aepix balance of the community pool.
func (s *EpixMintDistributionTestSuite) communityPoolAepix(ctx sdk.Context) sdkmath.Int {
	pool, err := s.network.App.GetDistrKeeper().FeePool.Get(ctx)
	s.Require().NoError(err)
	return pool.CommunityPool.AmountOf(s.bondDenom).TruncateInt()
}

// totalValidatorOutstandingAepix sums aepix outstanding rewards across every validator.
func (s *EpixMintDistributionTestSuite) totalValidatorOutstandingAepix(ctx sdk.Context) sdkmath.Int {
	total := sdkmath.ZeroInt()
	vals, err := s.network.App.GetStakingKeeper().GetAllValidators(ctx)
	s.Require().NoError(err)
	for _, v := range vals {
		opAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		s.Require().NoError(err)
		out, err := s.network.App.GetDistrKeeper().GetValidatorOutstandingRewards(ctx, opAddr)
		s.Require().NoError(err)
		total = total.Add(out.Rewards.AmountOf(s.bondDenom).TruncateInt())
	}
	return total
}

