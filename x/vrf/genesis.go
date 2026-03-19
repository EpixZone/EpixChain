package vrf

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/vrf/keeper"
	"github.com/cosmos/evm/x/vrf/types"
)

// InitGenesis initializes the vrf module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Restore beacons from genesis
	for _, beacon := range genState.Beacons {
		k.SetBeacon(ctx, beacon)
	}

	// Set latest height to the highest beacon
	if len(genState.Beacons) > 0 {
		var maxHeight uint64
		for _, beacon := range genState.Beacons {
			if beacon.Height > maxHeight {
				maxHeight = beacon.Height
			}
		}
		k.SetLatestHeight(ctx, maxHeight)
	}
}

// ExportGenesis returns the vrf module's exported genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)

	var beacons []types.RandomBeacon
	k.IterateBeacons(ctx, func(beacon types.RandomBeacon) bool {
		beacons = append(beacons, beacon)
		return false
	})

	return &types.GenesisState{
		Params:  params,
		Beacons: beacons,
	}
}
