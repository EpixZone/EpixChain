package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/x/vrf/types"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := types.DefaultGenesisState()
	require.NotNil(t, gs)
	require.Equal(t, types.DefaultParams(), gs.Params)
	require.Empty(t, gs.Beacons)
	require.NotNil(t, gs.Beacons) // should be empty slice, not nil
}

func TestGenesisStateValidate(t *testing.T) {
	validBeacon64 := strings.Repeat("ab", 32) // 64 hex chars

	tests := []struct {
		name      string
		genesis   types.GenesisState
		expPass   bool
		expErrMsg string
	}{
		{
			name:    "valid - default",
			genesis: *types.DefaultGenesisState(),
			expPass: true,
		},
		{
			name: "valid - with beacons",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.RandomBeacon{
					{Height: 10, Beacon: validBeacon64, Proposer: "cosmos1abc", Timestamp: 1000},
					{Height: 20, Beacon: validBeacon64, Proposer: "cosmos1def", Timestamp: 2000},
				},
			},
			expPass: true,
		},
		{
			name: "valid - empty beacons",
			genesis: types.GenesisState{
				Params:  types.DefaultParams(),
				Beacons: []types.RandomBeacon{},
			},
			expPass: true,
		},
		{
			name: "invalid - bad params",
			genesis: types.GenesisState{
				Params:  types.Params{Enabled: true, LookbackBlocks: 0},
				Beacons: []types.RandomBeacon{},
			},
			expPass:   false,
			expErrMsg: "invalid params",
		},
		{
			name: "invalid - duplicate heights",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.RandomBeacon{
					{Height: 10, Beacon: validBeacon64},
					{Height: 10, Beacon: validBeacon64},
				},
			},
			expPass:   false,
			expErrMsg: "duplicate beacon at height 10",
		},
		{
			name: "invalid - beacon too short (63 chars)",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.RandomBeacon{
					{Height: 10, Beacon: strings.Repeat("a", 63)},
				},
			},
			expPass:   false,
			expErrMsg: "invalid length 63",
		},
		{
			name: "invalid - beacon too long (65 chars)",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.RandomBeacon{
					{Height: 10, Beacon: strings.Repeat("a", 65)},
				},
			},
			expPass:   false,
			expErrMsg: "invalid length 65",
		},
		{
			name: "invalid - empty beacon",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.RandomBeacon{
					{Height: 10, Beacon: ""},
				},
			},
			expPass:   false,
			expErrMsg: "invalid length 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			}
		})
	}
}
