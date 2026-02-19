package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/evm/x/xid/types"
)

const maxPageSize uint64 = 10

var _ types.QueryServer = Keeper{}

// ResolveName resolves a name.tld to its name record
func (k Keeper) ResolveName(goCtx context.Context, req *types.QueryResolveNameRequest) (*types.QueryResolveNameResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	record, found := k.GetNameRecord(ctx, req.Tld, req.Name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", req.Name, req.Tld)
	}

	return &types.QueryResolveNameResponse{Record: &record}, nil
}

// ReverseResolve finds the primary name for an address
func (k Keeper) ReverseResolve(goCtx context.Context, req *types.QueryReverseResolveRequest) (*types.QueryReverseResolveResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	names := k.GetNamesByOwnerAddr(ctx, addr)
	if len(names) == 0 {
		return &types.QueryReverseResolveResponse{}, nil
	}

	// Return the first name as the primary
	return &types.QueryReverseResolveResponse{PrimaryName: &names[0]}, nil
}

// GetProfile returns the profile for a name
func (k Keeper) GetProfile(goCtx context.Context, req *types.QueryGetProfileRequest) (*types.QueryGetProfileResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists
	if !k.HasNameRecord(ctx, req.Tld, req.Name) {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", req.Name, req.Tld)
	}

	profile, found := k.GetProfileRecord(ctx, req.Tld, req.Name)
	if !found {
		return &types.QueryGetProfileResponse{}, nil
	}

	return &types.QueryGetProfileResponse{Profile: &profile}, nil
}

// GetDNSRecords returns all DNS records for a name
func (k Keeper) GetDNSRecords(goCtx context.Context, req *types.QueryGetDNSRecordsRequest) (*types.QueryGetDNSRecordsResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists
	if !k.HasNameRecord(ctx, req.Tld, req.Name) {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", req.Name, req.Tld)
	}

	records := k.GetAllDNSRecords(ctx, req.Tld, req.Name)
	return &types.QueryGetDNSRecordsResponse{Records: records}, nil
}

// GetTLD returns the configuration for a TLD
func (k Keeper) GetTLD(goCtx context.Context, req *types.QueryGetTLDRequest) (*types.QueryGetTLDResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidTLD, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config, found := k.GetTLDConfig(ctx, req.Tld)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrTLDNotFound, "TLD %q not found", req.Tld)
	}

	return &types.QueryGetTLDResponse{TldConfig: &config}, nil
}

// ListTLDs returns all registered TLDs
func (k Keeper) ListTLDs(goCtx context.Context, _ *types.QueryListTLDsRequest) (*types.QueryListTLDsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	tlds := k.GetAllTLDs(ctx)
	return &types.QueryListTLDsResponse{Tlds: tlds}, nil
}

// GetNamesByOwner returns names owned by an address with pagination
func (k Keeper) GetNamesByOwner(goCtx context.Context, req *types.QueryGetNamesByOwnerRequest) (*types.QueryGetNamesByOwnerResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: maxPageSize}
	} else if pageReq.Limit == 0 || pageReq.Limit > maxPageSize {
		pageReq.Limit = maxPageSize
	}

	// Disable CountTotal in Paginate — we use a stored counter instead
	wantTotal := pageReq.CountTotal
	pageReq.CountTotal = false

	names, pageRes, err := k.GetNamesByOwnerPaginated(ctx, addr, pageReq)
	if err != nil {
		return nil, err
	}

	// Supply the total from the O(1) stored counter
	if wantTotal {
		pageRes.Total = k.GetOwnerCount(ctx, addr)
	}

	return &types.QueryGetNamesByOwnerResponse{Names: names, Pagination: pageRes}, nil
}

// Params returns the module parameters
func (k Keeper) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// GetRegistrationFee returns the fee for a given name and TLD
func (k Keeper) GetRegistrationFee(goCtx context.Context, req *types.QueryGetRegistrationFeeRequest) (*types.QueryGetRegistrationFeeResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	fee, err := k.CalculateRegistrationFee(ctx, req.Tld, req.Name)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetRegistrationFeeResponse{Fee: fee.Amount}, nil
}
