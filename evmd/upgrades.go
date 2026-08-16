package evmd

import (
	"context"
	"fmt"

	"cosmossdk.io/math"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"

	"github.com/cosmos/evm/config"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	vrftypes "github.com/cosmos/evm/x/vrf/types"
	xidtypes "github.com/cosmos/evm/x/xid/types"
)

// Upgrade names
const UpgradeName_v0_5_1 = "v0.5.1"
const UpgradeName_v0_5_2 = "v0.5.2"
const UpgradeName_v0_5_3 = "v0.5.3"
const UpgradeName_v0_5_4 = "v0.5.4"
const UpgradeName_v0_5_5 = "v0.5.5"
const UpgradeName_v0_7_0 = "v0.7.0"
const UpgradeName_v0_7_1 = "v0.7.1"

// UpgradeName_v0_7_2 activates xID chain-attested finality.
//
// The vote-extension machinery (ExtendVote / VerifyVoteExtension /
// PrepareProposal injection / PreBlocker enabling vote extensions changes what
// CometBFT collects during voting, so all validators have to switch on the same
// height or consensus would fork. No store keys change (the xid store key has
// existed since v0.5.5) and no params gained fields — the only state change is
// the consensus-params enable height.
const UpgradeName_v0_7_2 = "v0.7.2"

// UpgradeName is the current upgrade (for store upgrades)
const UpgradeName = UpgradeName_v0_7_2

// RegisterUpgradeHandlers registers upgrade handlers for v0.5.1 and v0.5.2
func (app EVMD) RegisterUpgradeHandlers() {
	// Register v0.5.1 upgrade handler (the one that's currently stuck)
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_5_1,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.5.1 upgrade with recovery fix...")

			// Apply the recovery fix
			if err := app.applyRecoveryFix(ctx); err != nil {
				return nil, err
			}

			// Run module migrations
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	// Register v0.5.2 upgrade handler (for future use)
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_5_2,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.5.2 upgrade...")

			// Apply the recovery fix (in case v0.5.1 didn't run it)
			if err := app.applyRecoveryFix(ctx); err != nil {
				return nil, err
			}

			// Run module migrations
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	// Register v0.5.3 upgrade handler - Deterministic Emission Fix + Min Validator Self-Delegation
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_5_3,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.5.3 upgrade - Deterministic Emission Fix + Min Validator Self-Delegation...")

			// Log the upgrade details for transparency
			sdkCtx.Logger().Info("This upgrade fixes a consensus bug in token emission calculations")
			sdkCtx.Logger().Info("This upgrade also sets minimum validator self-delegation to 1,000,000 EPIX")

			// Set minimum validator self-delegation to 1M EPIX
			// 1M EPIX in aepix (18 decimals): 1 * 10^6 * 10^18 = 10^24
			minValidatorSelfDelegation, ok := math.NewIntFromString("1000000000000000000000000")
			if !ok {
				return nil, fmt.Errorf("failed to parse min validator self delegation")
			}

			epixmintParams := app.EpixMintKeeper.GetParams(sdkCtx)
			epixmintParams.MinValidatorSelfDelegation = minValidatorSelfDelegation
			if err := app.EpixMintKeeper.SetParams(sdkCtx, epixmintParams); err != nil {
				return nil, fmt.Errorf("failed to set epixmint params: %w", err)
			}
			sdkCtx.Logger().Info("Minimum validator self-delegation set to 1,000,000 EPIX")

			// Run module migrations
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	// Register v0.5.4 upgrade handler - IBC Wallet Compatibility Fix + Min Validator Self-Delegation
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_5_4,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.5.4 upgrade - IBC Wallet Compatibility Fix + Min Validator Self-Delegation...")
			sdkCtx.Logger().Info("This upgrade fixes IBC transfer signature verification for Keplr/Leap wallets")
			sdkCtx.Logger().Info("This upgrade also sets minimum validator self-delegation to 1,000,000 EPIX")

			// Set minimum validator self-delegation to 1M EPIX
			// 1M EPIX in aepix (18 decimals): 1 * 10^6 * 10^18 = 10^24
			minValidatorSelfDelegation, ok := math.NewIntFromString("1000000000000000000000000")
			if !ok {
				return nil, fmt.Errorf("failed to parse min validator self delegation")
			}

			epixmintParams := app.EpixMintKeeper.GetParams(sdkCtx)
			epixmintParams.MinValidatorSelfDelegation = minValidatorSelfDelegation
			if err := app.EpixMintKeeper.SetParams(sdkCtx, epixmintParams); err != nil {
				return nil, fmt.Errorf("failed to set epixmint params: %w", err)
			}
			sdkCtx.Logger().Info("Minimum validator self-delegation set to 1,000,000 EPIX")

			// IBC signature fix requires no state migration - the fix is in signature
			// verification logic (crypto/ethsecp256k1/ethsecp256k1.go) which strips
			// extra amino fields added by IBC v10 (use_aliasing, encoding) that
			// wallets don't include.

			// Run module migrations
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	// Register v0.5.5 upgrade handler - xID Identity System + VRF Randomness Beacon
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_5_5,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.5.5 upgrade - xID Identity System + VRF Randomness Beacon...")
			sdkCtx.Logger().Info("This upgrade introduces the xID on-chain identity and DNS module")
			sdkCtx.Logger().Info("This upgrade introduces the VRF on-chain verifiable randomness module")

			// Add the xID and VRF precompiles to the active static precompiles list
			evmParams := app.EVMKeeper.GetParams(sdkCtx)
			evmParams.ActiveStaticPrecompiles = append(
				evmParams.ActiveStaticPrecompiles,
				evmtypes.XIDPrecompileAddress,
				evmtypes.VRFPrecompileAddress,
			)
			if err := app.EVMKeeper.SetParams(sdkCtx, evmParams); err != nil {
				return nil, fmt.Errorf("failed to set EVM params with xID and VRF precompiles: %w", err)
			}
			sdkCtx.Logger().Info("Enabled xID precompile at " + evmtypes.XIDPrecompileAddress)
			sdkCtx.Logger().Info("Enabled VRF precompile at " + evmtypes.VRFPrecompileAddress)

			// RunMigrations will call InitGenesis for both new modules
			// (since they have no prior version in the version map):
			// - xid: creates the default .epix TLD with length-based pricing
			// - vrf: sets default params (Enabled=true, LookbackBlocks=256)
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	// Register v0.6.0 -> v0.7.0 upgrade handler.
	// State changes are wiring-only (drop x/precisebank, drop x/ibc/transfer override,
	// adopt Krakatoa app-side mempool, BlockSTM with virtual fees, optimistic execution,
	// ibc-go v10 -> v11). x/precisebank store key is removed; existing precisebank state
	// is dropped on the upgrade boundary via storeUpgrades.Deleted below.
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_7_0,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.6 -> v0.7.0 upgrade (cosmos/evm v0.7)...")
			sdkCtx.Logger().Info("- drop x/precisebank (chain is 18-decimal; remainder is empty)")
			sdkCtx.Logger().Info("- drop x/ibc/transfer override")
			sdkCtx.Logger().Info("- ibc-go v10 -> v11, cosmos-sdk v0.53 -> v0.54, cometbft v0.38 -> v0.39")
			sdkCtx.Logger().Info("- Krakatoa app-side mempool, BlockSTM + virtual fees, optimistic execution")

			// Run module ConsensusVersion migrations (auth/bank/staking/gov from
			// sdk v0.54, ibc client/connection/channel from ibc-go v11, x/vm and
			// x/erc20 from cosmos/evm v0.7). Stock SDK handles the heavy lifting;
			// the steps below cover gaps specific to a v0.6-shaped chain.
			vm, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
			if err != nil {
				return nil, err
			}

			// v0.7 added `Params.HistoryServeWindow` (EIP-2935 history buffer).
			// A v0.6 params row deserialises with HistoryServeWindow = 0, which
			// disables history serving. Backfill the default so eth_getBlockByHash
			// and friends keep working as expected post-upgrade.
			//
			// (`Params.AccessControl` does not need a backfill — its zero value
			//  is `AccessTypePermissionless` for both Create and Call, which
			//  matches the upstream default `DefaultAccessControl`.)
			evmParams := app.EVMKeeper.GetParams(sdkCtx)
			if evmParams.HistoryServeWindow == 0 {
				evmParams.HistoryServeWindow = evmtypes.DefaultHistoryServeWindow
				sdkCtx.Logger().Info(fmt.Sprintf("Backfilling Params.HistoryServeWindow to %d (default for EIP-2935)", evmtypes.DefaultHistoryServeWindow))
			}
			if err := app.EVMKeeper.SetParams(sdkCtx, evmParams); err != nil {
				return nil, fmt.Errorf("failed to backfill v0.7 EVM params: %w", err)
			}

			// Refresh the EVM coin info store from bank denom metadata. After
			// the binary swap the in-memory `evmCoinInfo` global is rebuilt by
			// `vmModule.HydrateGlobals` in app.go, but the on-disk key/value
			// row is what HydrateGlobals reads. Re-running InitEvmCoinInfo
			// guarantees that row matches the live bank metadata (display
			// denom + 18-decimal exponent) before virtual-fee collection
			// reads it on the first post-upgrade tx — without this a stale
			// row could panic `x/vm/keeper.DeductFees`.
			if err := app.EVMKeeper.InitEvmCoinInfo(sdkCtx); err != nil {
				return nil, fmt.Errorf("failed to re-init EVM coin info: %w", err)
			}
			sdkCtx.Logger().Info("EVM coin info refreshed from bank metadata")

			sdkCtx.Logger().Info("EpixChain v0.7.0 upgrade complete")
			return vm, nil
		},
	)

	// Register v0.7.0 -> v0.7.1 upgrade handler (cosmos/evm v0.7.1 sync).
	// Migrations-only: no store keys change and no params gained fields, so
	// there is nothing to backfill. The boundary exists purely so every
	// validator switches transaction-processing behaviour on the same block.
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_7_1,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.7.0 -> v0.7.1 upgrade (cosmos/evm v0.7.1)...")
			sdkCtx.Logger().Info("- EVM signature verification respects ctx.IsSigverifyTx()")
			sdkCtx.Logger().Info("- mempool rejects EVM txs below base fee at admission")
			sdkCtx.Logger().Info("- statedb locked-balance snapshot + balance/event amount hardening")
			sdkCtx.Logger().Info("- erc20 v2 IBC middleware ack validation aligned with ibc-go")
			sdkCtx.Logger().Info("(the module-account statedb fix shipped earlier as a binary hotfix)")

			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	// Register v0.7.1 -> v0.7.2 upgrade handler - activate xID finality attestation.
	// Migrations run first, then vote extensions are switched on by setting the
	// consensus-params enable height to the next block. baseapp returns the current
	// consensus params as ConsensusParamUpdates at the end of every FinalizeBlock, so
	// writing the enable height here is what tells CometBFT to start collecting the
	// per-validator digest attestations. No store keys change.
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName_v0_7_2,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Info("Starting EpixChain v0.7.1 -> v0.7.2 upgrade - xID finality attestation...")

			vm, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
			if err != nil {
				return nil, err
			}

			// Enable ABCI++ vote extensions from the next block. CometBFT requires the
			// enable height to be strictly greater than the height that produces the
			// param update (this block), so +1 is the earliest valid value and the
			// feature turns on the block immediately after the upgrade.
			cp := app.GetConsensusParams(sdkCtx)
			enableHeight := sdkCtx.BlockHeight() + 1
			if cp.Abci == nil {
				cp.Abci = &cmtproto.ABCIParams{}
			}
			if cp.Abci.VoteExtensionsEnableHeight == 0 {
				cp.Abci.VoteExtensionsEnableHeight = enableHeight
				if err := app.StoreConsensusParams(sdkCtx, cp); err != nil {
					return nil, fmt.Errorf("failed to enable vote extensions: %w", err)
				}
				sdkCtx.Logger().Info(fmt.Sprintf("xID finality: vote extensions enabled from height %d", enableHeight))
			} else {
				sdkCtx.Logger().Info(fmt.Sprintf("xID finality: vote extensions already enabled at height %d (no change)", cp.Abci.VoteExtensionsEnableHeight))
			}

			sdkCtx.Logger().Info("EpixChain v0.7.2 upgrade complete")
			return vm, nil
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// v0.5.1 through v0.5.4 had no store-key changes. Earlier versions of
	// this file passed `Added: []string{}` which is equivalent to omitting
	// the store-loader call entirely — drop the no-op for clarity.

	// Handle v0.5.5 upgrade - adds xID and VRF module store keys
	if upgradeInfo.Name == UpgradeName_v0_5_5 &&
		!app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{xidtypes.StoreKey, vrftypes.StoreKey},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	// Handle v0.6 -> v0.7.0 upgrade - deletes the precisebank store key.
	// Mainnet currently runs `epixd v0.5.5`, which is built on cosmos/evm
	// v0.6.x and *does* have the precisebank store registered (confirmed
	// by querying cosmos.evm.precisebank.v1.Query/Remainder on-chain — the
	// remainder is zero but the store key exists). Removing the key here
	// prunes the orphan from the IAVL tree at the upgrade height so the
	// post-v0.7 binary (which no longer registers the module) doesn't see
	// a dangling sub-store.
	if upgradeInfo.Name == UpgradeName_v0_7_0 &&
		!app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Deleted: []string{"precisebank"},
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

}

// applyRecoveryFix contains the actual recovery logic for v0.5.1/v0.5.2 upgrades
// This fixes the IBC channel state that was not properly migrated:
// - Sets denom metadata for aepix/epix
// - Updates EVM params and coin info
// - Initializes missing IBC channel sequence counters
func (app EVMD) applyRecoveryFix(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Set denom metadata for EpixChain's native token (aepix/epix)
	app.BankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
		Description: "The native staking and governance token of the EpixChain",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    config.EpixChainDenom,
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    config.EpixDisplayDenom,
				Exponent: 18,
				Aliases:  nil,
			},
		},
		Base:    config.EpixChainDenom,
		Display: config.EpixDisplayDenom,
		Name:    "EpixChain",
		Symbol:  "EPIX",
		URI:     "https://epix.zone/",
	})
	sdkCtx.Logger().Info("EpixChain denom metadata set successfully")

	// Update EVM params to set the EvmDenom and ExtendedDenomOptions
	evmParams := app.EVMKeeper.GetParams(sdkCtx)
	evmParams.EvmDenom = config.EpixChainDenom
	evmParams.ExtendedDenomOptions = &evmtypes.ExtendedDenomOptions{
		ExtendedDenom: config.EpixChainDenom,
	}
	if err := app.EVMKeeper.SetParams(sdkCtx, evmParams); err != nil {
		return fmt.Errorf("failed to set EVM params: %w", err)
	}
	sdkCtx.Logger().Info("EVM params updated successfully")

	// Initialize EVM coin info from the bank metadata
	if err := app.EVMKeeper.InitEvmCoinInfo(sdkCtx); err != nil {
		return fmt.Errorf("failed to initialize EVM coin info: %w", err)
	}
	sdkCtx.Logger().Info("EVM coin info initialized successfully")

	// Fix IBC channel sequence counters
	channels := app.IBCKeeper.ChannelKeeper.GetAllChannels(sdkCtx)
	sdkCtx.Logger().Info(fmt.Sprintf("Found %d IBC channels to check", len(channels)))

	for _, channel := range channels {
		portID := channel.PortId
		channelID := channel.ChannelId

		// Check if NextSequenceSend exists
		_, found := app.IBCKeeper.ChannelKeeper.GetNextSequenceSend(sdkCtx, portID, channelID)
		if !found {
			app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(sdkCtx, portID, channelID, 1)
			sdkCtx.Logger().Info(fmt.Sprintf("Initialized NextSequenceSend for port %s, channel %s to 1", portID, channelID))
		}

		// Check if NextSequenceRecv exists
		_, found = app.IBCKeeper.ChannelKeeper.GetNextSequenceRecv(sdkCtx, portID, channelID)
		if !found {
			app.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(sdkCtx, portID, channelID, 1)
			sdkCtx.Logger().Info(fmt.Sprintf("Initialized NextSequenceRecv for port %s, channel %s to 1", portID, channelID))
		}

		// For unordered channels, check NextSequenceAck
		if channel.Ordering == channeltypes.UNORDERED {
			_, found = app.IBCKeeper.ChannelKeeper.GetNextSequenceAck(sdkCtx, portID, channelID)
			if !found {
				app.IBCKeeper.ChannelKeeper.SetNextSequenceAck(sdkCtx, portID, channelID, 1)
				sdkCtx.Logger().Info(fmt.Sprintf("Initialized NextSequenceAck for port %s, channel %s to 1", portID, channelID))
			}
		}
	}

	sdkCtx.Logger().Info("IBC channel sequence counters recovery complete")
	return nil
}
