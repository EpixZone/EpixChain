package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/x/vrf/types"
)

func TestRandomBeaconMarshalUnmarshal(t *testing.T) {
	beacon := types.RandomBeacon{
		Height:    100,
		Beacon:    strings.Repeat("ab", 32),
		Proposer:  "cosmosvalcons1abcdef",
		Timestamp: 1700000000,
	}

	bz, err := beacon.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	var decoded types.RandomBeacon
	err = decoded.Unmarshal(bz)
	require.NoError(t, err)

	require.Equal(t, beacon.Height, decoded.Height)
	require.Equal(t, beacon.Beacon, decoded.Beacon)
	require.Equal(t, beacon.Proposer, decoded.Proposer)
	require.Equal(t, beacon.Timestamp, decoded.Timestamp)
}

func TestRandomBeaconMarshalZeroValues(t *testing.T) {
	beacon := types.RandomBeacon{}

	bz, err := beacon.Marshal()
	require.NoError(t, err)

	var decoded types.RandomBeacon
	err = decoded.Unmarshal(bz)
	require.NoError(t, err)

	require.Equal(t, uint64(0), decoded.Height)
	require.Equal(t, "", decoded.Beacon)
	require.Equal(t, "", decoded.Proposer)
	require.Equal(t, int64(0), decoded.Timestamp)
}

func TestRandomBeaconSize(t *testing.T) {
	beacon := types.RandomBeacon{
		Height:    42,
		Beacon:    strings.Repeat("cd", 32),
		Proposer:  "cosmos1proposer",
		Timestamp: 999999,
	}

	bz, err := beacon.Marshal()
	require.NoError(t, err)
	require.Equal(t, beacon.Size(), len(bz))
}

func TestParamsMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name   string
		params types.Params
	}{
		{
			name:   "enabled with default lookback",
			params: types.Params{Enabled: true, LookbackBlocks: 256},
		},
		{
			name:   "disabled",
			params: types.Params{Enabled: false, LookbackBlocks: 512},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bz, err := tc.params.Marshal()
			require.NoError(t, err)

			var decoded types.Params
			err = decoded.Unmarshal(bz)
			require.NoError(t, err)

			require.Equal(t, tc.params.Enabled, decoded.Enabled)
			require.Equal(t, tc.params.LookbackBlocks, decoded.LookbackBlocks)
		})
	}
}

func TestGenesisStateMarshalUnmarshal(t *testing.T) {
	gs := types.GenesisState{
		Params: types.DefaultParams(),
		Beacons: []types.RandomBeacon{
			{Height: 10, Beacon: strings.Repeat("aa", 32), Proposer: "val1", Timestamp: 100},
			{Height: 20, Beacon: strings.Repeat("bb", 32), Proposer: "val2", Timestamp: 200},
		},
	}

	bz, err := gs.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	var decoded types.GenesisState
	err = decoded.Unmarshal(bz)
	require.NoError(t, err)

	require.Equal(t, gs.Params.Enabled, decoded.Params.Enabled)
	require.Equal(t, gs.Params.LookbackBlocks, decoded.Params.LookbackBlocks)
	require.Len(t, decoded.Beacons, 2)
	require.Equal(t, uint64(10), decoded.Beacons[0].Height)
	require.Equal(t, uint64(20), decoded.Beacons[1].Height)
}
