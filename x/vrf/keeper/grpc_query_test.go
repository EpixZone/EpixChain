package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/evm/x/vrf/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *KeeperTestSuite) TestGRPCGetBeacon_Found() {
	beacon := makeBeacon(50)
	s.keeper.SetBeacon(s.ctx, beacon)

	resp, err := s.keeper.GetBeacon(sdk.WrapSDKContext(s.ctx), &types.QueryGetBeaconRequest{Height: 50})
	s.Require().NoError(err)
	s.Require().NotNil(resp.Beacon)
	s.Require().Equal(uint64(50), resp.Beacon.Height)
	s.Require().Equal(beacon.Beacon, resp.Beacon.Beacon)
	s.Require().Equal(beacon.Proposer, resp.Beacon.Proposer)
	s.Require().Equal(beacon.Timestamp, resp.Beacon.Timestamp)
}

func (s *KeeperTestSuite) TestGRPCGetBeacon_NotFound() {
	_, err := s.keeper.GetBeacon(sdk.WrapSDKContext(s.ctx), &types.QueryGetBeaconRequest{Height: 999})
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.NotFound, st.Code())
}

func (s *KeeperTestSuite) TestGRPCGetBeacon_NilRequest() {
	_, err := s.keeper.GetBeacon(sdk.WrapSDKContext(s.ctx), nil)
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
}

func (s *KeeperTestSuite) TestGRPCLatestBeacon_Found() {
	s.keeper.SetBeacon(s.ctx, makeBeacon(10))
	s.keeper.SetBeacon(s.ctx, makeBeacon(20))
	s.keeper.SetBeacon(s.ctx, makeBeacon(30))
	s.keeper.SetLatestHeight(s.ctx, 30)

	resp, err := s.keeper.LatestBeacon(sdk.WrapSDKContext(s.ctx), &types.QueryLatestBeaconRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp.Beacon)
	s.Require().Equal(uint64(30), resp.Beacon.Height)
}

func (s *KeeperTestSuite) TestGRPCLatestBeacon_Empty() {
	_, err := s.keeper.LatestBeacon(sdk.WrapSDKContext(s.ctx), &types.QueryLatestBeaconRequest{})
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.NotFound, st.Code())
}

func (s *KeeperTestSuite) TestGRPCParams() {
	custom := types.Params{Enabled: false, LookbackBlocks: 512}
	err := s.keeper.SetParams(s.ctx, custom)
	s.Require().NoError(err)

	resp, err := s.keeper.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(false, resp.Params.Enabled)
	s.Require().Equal(uint64(512), resp.Params.LookbackBlocks)
}

func (s *KeeperTestSuite) TestGRPCParams_Default() {
	resp, err := s.keeper.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultParams().Enabled, resp.Params.Enabled)
	s.Require().Equal(types.DefaultParams().LookbackBlocks, resp.Params.LookbackBlocks)
}
