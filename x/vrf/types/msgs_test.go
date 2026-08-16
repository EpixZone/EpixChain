package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/x/vrf/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestMsgUpdateParamsValidateBasic(t *testing.T) {
	validAuthority := authtypes.NewModuleAddress("gov").String()

	tests := []struct {
		name    string
		msg     types.MsgUpdateParams
		expPass bool
	}{
		{
			name: "valid",
			msg: types.MsgUpdateParams{
				Authority: validAuthority,
				Params:    types.DefaultParams(),
			},
			expPass: true,
		},
		{
			name: "invalid - empty authority",
			msg: types.MsgUpdateParams{
				Authority: "",
				Params:    types.DefaultParams(),
			},
			expPass: false,
		},
		{
			name: "invalid - malformed authority",
			msg: types.MsgUpdateParams{
				Authority: "not-a-bech32-address",
				Params:    types.DefaultParams(),
			},
			expPass: false,
		},
		{
			name: "invalid - bad params",
			msg: types.MsgUpdateParams{
				Authority: validAuthority,
				Params:    types.Params{Enabled: true, LookbackBlocks: 0},
			},
			expPass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgUpdateParamsGetSigners(t *testing.T) {
	authority := authtypes.NewModuleAddress("gov")
	msg := types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    types.DefaultParams(),
	}
	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, authority, signers[0])
}
