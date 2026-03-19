package types_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/x/vrf/types"
)

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()
	require.True(t, params.Enabled)
	require.Equal(t, uint64(256), params.LookbackBlocks)
}

func TestParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  types.Params
		expPass bool
	}{
		{
			name:    "valid - default",
			params:  types.DefaultParams(),
			expPass: true,
		},
		{
			name:    "valid - disabled",
			params:  types.Params{Enabled: false, LookbackBlocks: 100},
			expPass: true,
		},
		{
			name:    "valid - minimum lookback",
			params:  types.Params{Enabled: true, LookbackBlocks: 1},
			expPass: true,
		},
		{
			name:    "valid - large lookback",
			params:  types.Params{Enabled: true, LookbackBlocks: math.MaxUint64},
			expPass: true,
		},
		{
			name:    "invalid - zero lookback",
			params:  types.Params{Enabled: true, LookbackBlocks: 0},
			expPass: false,
		},
		{
			name:    "invalid - zero lookback even if disabled",
			params:  types.Params{Enabled: false, LookbackBlocks: 0},
			expPass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), "lookback_blocks must be greater than 0")
			}
		})
	}
}
