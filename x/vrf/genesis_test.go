package vrf_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/evm/x/vrf"
	"github.com/cosmos/evm/x/vrf/keeper"
	"github.com/cosmos/evm/x/vrf/types"
)

func setupKeeper(t *testing.T) (keeper.Keeper, storetypes.StoreKey, testutil.TestContext) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		authtypes.NewModuleAddress("gov").String(),
	)
	return k, key, testCtx
}

func TestInitGenesis_Default(t *testing.T) {
	k, _, testCtx := setupKeeper(t)
	ctx := testCtx.Ctx

	vrf.InitGenesis(ctx, k, types.DefaultGenesisState())

	params := k.GetParams(ctx)
	require.Equal(t, types.DefaultParams(), params)

	_, found := k.GetLatestBeacon(ctx)
	require.False(t, found)
}

func TestInitGenesis_WithBeacons(t *testing.T) {
	k, _, testCtx := setupKeeper(t)
	ctx := testCtx.Ctx

	beacon64 := strings.Repeat("ab", 32)
	gs := &types.GenesisState{
		Params: types.DefaultParams(),
		Beacons: []types.RandomBeacon{
			{Height: 10, Beacon: beacon64, Proposer: "val1", Timestamp: 100},
			{Height: 20, Beacon: beacon64, Proposer: "val2", Timestamp: 200},
			{Height: 30, Beacon: beacon64, Proposer: "val3", Timestamp: 300},
		},
	}
	vrf.InitGenesis(ctx, k, gs)

	// All beacons should be retrievable
	for _, h := range []uint64{10, 20, 30} {
		_, found := k.GetBeaconByHeight(ctx, h)
		require.True(t, found, "beacon at height %d should exist", h)
	}

	// Latest height should be the max
	height, found := k.GetLatestHeight(ctx)
	require.True(t, found)
	require.Equal(t, uint64(30), height)
}

func TestInitGenesis_UnsortedBeacons(t *testing.T) {
	k, _, testCtx := setupKeeper(t)
	ctx := testCtx.Ctx

	beacon64 := strings.Repeat("ab", 32)
	gs := &types.GenesisState{
		Params: types.DefaultParams(),
		Beacons: []types.RandomBeacon{
			{Height: 30, Beacon: beacon64, Proposer: "val3", Timestamp: 300},
			{Height: 10, Beacon: beacon64, Proposer: "val1", Timestamp: 100},
			{Height: 20, Beacon: beacon64, Proposer: "val2", Timestamp: 200},
		},
	}
	vrf.InitGenesis(ctx, k, gs)

	// Latest height should still be 30 (the max, not last in slice)
	height, found := k.GetLatestHeight(ctx)
	require.True(t, found)
	require.Equal(t, uint64(30), height)
}

func TestInitGenesis_InvalidPanics(t *testing.T) {
	k, _, testCtx := setupKeeper(t)
	ctx := testCtx.Ctx

	gs := &types.GenesisState{
		Params:  types.Params{Enabled: true, LookbackBlocks: 0},
		Beacons: []types.RandomBeacon{},
	}

	require.Panics(t, func() {
		vrf.InitGenesis(ctx, k, gs)
	})
}

func TestExportGenesis(t *testing.T) {
	k, _, testCtx := setupKeeper(t)
	ctx := testCtx.Ctx

	beacon64 := strings.Repeat("cd", 32)
	gs := &types.GenesisState{
		Params: types.Params{Enabled: false, LookbackBlocks: 128},
		Beacons: []types.RandomBeacon{
			{Height: 5, Beacon: beacon64, Proposer: "val1", Timestamp: 50},
			{Height: 15, Beacon: beacon64, Proposer: "val2", Timestamp: 150},
		},
	}
	vrf.InitGenesis(ctx, k, gs)

	exported := vrf.ExportGenesis(ctx, k)
	require.Equal(t, false, exported.Params.Enabled)
	require.Equal(t, uint64(128), exported.Params.LookbackBlocks)
	require.Len(t, exported.Beacons, 2)
	// Should be ascending order from IterateBeacons
	require.Equal(t, uint64(5), exported.Beacons[0].Height)
	require.Equal(t, uint64(15), exported.Beacons[1].Height)
}

func TestExportGenesis_Empty(t *testing.T) {
	k, _, testCtx := setupKeeper(t)
	ctx := testCtx.Ctx

	vrf.InitGenesis(ctx, k, types.DefaultGenesisState())

	exported := vrf.ExportGenesis(ctx, k)
	require.Equal(t, types.DefaultParams().Enabled, exported.Params.Enabled)
	require.Equal(t, types.DefaultParams().LookbackBlocks, exported.Params.LookbackBlocks)
	require.Empty(t, exported.Beacons)
}

func TestInitExportRoundTrip(t *testing.T) {
	k, _, testCtx := setupKeeper(t)
	ctx := testCtx.Ctx

	beacon64 := strings.Repeat("ef", 32)
	original := &types.GenesisState{
		Params: types.Params{Enabled: true, LookbackBlocks: 64},
		Beacons: []types.RandomBeacon{
			{Height: 100, Beacon: beacon64, Proposer: "valA", Timestamp: 1000},
			{Height: 200, Beacon: beacon64, Proposer: "valB", Timestamp: 2000},
		},
	}
	vrf.InitGenesis(ctx, k, original)

	exported := vrf.ExportGenesis(ctx, k)

	require.Equal(t, original.Params.Enabled, exported.Params.Enabled)
	require.Equal(t, original.Params.LookbackBlocks, exported.Params.LookbackBlocks)
	require.Len(t, exported.Beacons, len(original.Beacons))
	require.Equal(t, uint64(100), exported.Beacons[0].Height)
	require.Equal(t, uint64(200), exported.Beacons[1].Height)
}
