package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/vrf/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// GetBeacon returns the random beacon for a specific block height.
func (k Keeper) GetBeacon(goCtx context.Context, req *types.QueryGetBeaconRequest) (*types.QueryGetBeaconResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	beacon, found := k.GetBeaconByHeight(ctx, req.Height)
	if !found {
		return nil, status.Errorf(codes.NotFound, "beacon not found for height %d", req.Height)
	}

	return &types.QueryGetBeaconResponse{Beacon: &beacon}, nil
}

// LatestBeacon returns the most recent random beacon.
func (k Keeper) LatestBeacon(goCtx context.Context, req *types.QueryLatestBeaconRequest) (*types.QueryLatestBeaconResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	beacon, found := k.GetLatestBeacon(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "no beacons available yet")
	}

	return &types.QueryLatestBeaconResponse{Beacon: &beacon}, nil
}

// Params returns the current VRF module parameters.
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}
