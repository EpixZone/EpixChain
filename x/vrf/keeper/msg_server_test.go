package keeper_test

import (
	"github.com/cosmos/evm/x/vrf/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (s *KeeperTestSuite) TestUpdateParams_Success() {
	authority := authtypes.NewModuleAddress("gov").String()
	newParams := types.Params{Enabled: false, LookbackBlocks: 128}

	msg := &types.MsgUpdateParams{
		Authority: authority,
		Params:    newParams,
	}

	resp, err := s.keeper.UpdateParams(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	got := s.keeper.GetParams(s.ctx)
	s.Require().Equal(false, got.Enabled)
	s.Require().Equal(uint64(128), got.LookbackBlocks)
}

func (s *KeeperTestSuite) TestUpdateParams_Unauthorized() {
	msg := &types.MsgUpdateParams{
		Authority: "cosmos1wrongauthority",
		Params:    types.DefaultParams(),
	}

	_, err := s.keeper.UpdateParams(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

func (s *KeeperTestSuite) TestUpdateParams_InvalidParams() {
	authority := authtypes.NewModuleAddress("gov").String()
	msg := &types.MsgUpdateParams{
		Authority: authority,
		Params:    types.Params{Enabled: true, LookbackBlocks: 0},
	}

	_, err := s.keeper.UpdateParams(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "lookback_blocks must be greater than 0")
}
