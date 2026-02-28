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

	// Check if the owner has an explicit primary name set
	if pTld, pName, found := k.GetPrimaryNameEntry(ctx, addr); found {
		// Verify the name still exists and is still owned by this address
		if record, exists := k.GetNameRecord(ctx, pTld, pName); exists && record.Owner == addr.String() {
			return &types.QueryReverseResolveResponse{PrimaryName: &record}, nil
		}
		// Stale primary — clean it up
		k.DeletePrimaryNameEntry(ctx, addr)
	}

	// Fall back to first name from owner index
	names := k.GetNamesByOwnerAddr(ctx, addr)
	if len(names) == 0 {
		return &types.QueryReverseResolveResponse{}, nil
	}

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

// ListAllNames returns a paginated list of all registered names
func (k Keeper) ListAllNames(goCtx context.Context, req *types.QueryListAllNamesRequest) (*types.QueryListAllNamesResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: maxPageSize}
	} else if pageReq.Limit == 0 || pageReq.Limit > maxPageSize {
		pageReq.Limit = maxPageSize
	}

	// Disable CountTotal in Paginate — we use a stored counter instead
	wantTotal := pageReq.CountTotal
	pageReq.CountTotal = false

	names, pageRes, err := k.GetAllNamesPaginated(ctx, pageReq)
	if err != nil {
		return nil, err
	}

	// Supply the total from the O(1) stored counter
	if wantTotal {
		pageRes.Total = k.GetGlobalNameCount(ctx)
	}

	return &types.QueryListAllNamesResponse{Names: names, Pagination: pageRes}, nil
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

// GetEpixNetPeers returns all EpixNet peers for a name
func (k Keeper) GetEpixNetPeers(goCtx context.Context, req *types.QueryGetEpixNetPeersRequest) (*types.QueryGetEpixNetPeersResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists
	if !k.HasNameRecord(ctx, req.Tld, req.Name) {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", req.Name, req.Tld)
	}

	peers := k.GetAllEpixNetPeers(ctx, req.Tld, req.Name)
	return &types.QueryGetEpixNetPeersResponse{Peers: peers}, nil
}

// QueryStateDigest returns the current xID state digest
func (k Keeper) QueryStateDigest(goCtx context.Context, _ *types.QueryStateDigestRequest) (*types.QueryStateDigestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	digest, found := k.GetStateDigest(ctx)
	if !found {
		return &types.QueryStateDigestResponse{}, nil
	}

	return &types.QueryStateDigestResponse{
		Digest:   digest.Digest,
		Height:   digest.Height,
		NumNames: digest.NumNames,
	}, nil
}

// QueryAttestations returns attestations for a given digest
func (k Keeper) QueryAttestations(goCtx context.Context, req *types.QueryAttestationsRequest) (*types.QueryAttestationsResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAttestation, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// If no digest specified, use the current one
	digest := req.Digest
	if digest == "" {
		sd, found := k.GetStateDigest(ctx)
		if !found {
			return &types.QueryAttestationsResponse{}, nil
		}
		digest = sd.Digest
	}

	attestations := k.GetAttestations(ctx, digest)
	finalized := k.IsDigestFinalized(ctx, digest)

	return &types.QueryAttestationsResponse{
		Attestations: attestations,
		Finalized:    finalized,
	}, nil
}

// QueryStateSnapshot returns a paginated snapshot of all domain data
func (k Keeper) QueryStateSnapshot(goCtx context.Context, req *types.QueryStateSnapshotRequest) (*types.QueryStateSnapshotResponse, error) {
	if req == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: maxPageSize}
	} else if pageReq.Limit == 0 || pageReq.Limit > maxPageSize {
		pageReq.Limit = maxPageSize
	}

	names, pageRes, err := k.GetAllNamesPaginated(ctx, pageReq)
	if err != nil {
		return nil, err
	}

	var domains []types.DomainSnapshot
	for _, record := range names {
		snap := types.DomainSnapshot{
			Record: record,
		}

		if profile, found := k.GetProfileRecord(ctx, record.Tld, record.Name); found {
			snap.Profile = &profile
		}

		dns := k.GetAllDNSRecords(ctx, record.Tld, record.Name)
		if len(dns) > 0 {
			snap.DnsRecords = dns
		}

		peers := k.GetAllEpixNetPeers(ctx, record.Tld, record.Name)
		if len(peers) > 0 {
			snap.Peers = peers
		}

		if cr, found := k.GetContentRoot(ctx, record.Tld, record.Name); found && cr.Root != "" {
			snap.ContentRoot = cr.Root
		}

		domains = append(domains, snap)
	}

	digest, _ := k.GetStateDigest(ctx)

	return &types.QueryStateSnapshotResponse{
		Domains:    domains,
		Digest:     digest.Digest,
		Height:     digest.Height,
		Pagination: pageRes,
	}, nil
}

// GetStats returns xID module statistics
func (k Keeper) GetStats(goCtx context.Context, _ *types.QueryGetStatsRequest) (*types.QueryGetStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tlds := k.GetAllTLDs(ctx)
	var tldStats []types.TLDStats
	for _, tld := range tlds {
		tldStats = append(tldStats, types.TLDStats{
			Tld:        tld.Tld,
			NameCount:  k.GetTLDNameCount(ctx, tld.Tld),
			FeesBurned: k.GetTLDFeesBurned(ctx, tld.Tld).String(),
			Enabled:    tld.Enabled,
		})
	}

	return &types.QueryGetStatsResponse{
		TotalNames:      k.GetGlobalNameCount(ctx),
		TotalFeesBurned: k.GetGlobalFeesBurned(ctx).String(),
		TldStats:        tldStats,
	}, nil
}
