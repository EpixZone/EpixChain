package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/evm/x/xid/types"
)

var _ types.MsgServer = &Keeper{}

// RegisterName handles MsgRegisterName
func (k Keeper) RegisterName(goCtx context.Context, msg *types.MsgRegisterName) (*types.MsgRegisterNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if err := k.RegisterNameRecord(ctx, ownerAddr, msg.Tld, msg.Name); err != nil {
		return nil, err
	}

	return &types.MsgRegisterNameResponse{}, nil
}

// TransferName handles MsgTransferName
func (k Keeper) TransferName(goCtx context.Context, msg *types.MsgTransferName) (*types.MsgTransferNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	currentOwner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	newOwner, err := sdk.AccAddressFromBech32(msg.NewOwner)
	if err != nil {
		return nil, err
	}

	if err := k.TransferNameRecord(ctx, currentOwner, newOwner, msg.Tld, msg.Name); err != nil {
		return nil, err
	}

	return &types.MsgTransferNameResponse{}, nil
}

// UpdateProfile handles MsgUpdateProfile
func (k Keeper) UpdateProfile(goCtx context.Context, msg *types.MsgUpdateProfile) (*types.MsgUpdateProfileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, msg.Tld, msg.Name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", msg.Name, msg.Tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	k.SetProfileRecord(ctx, msg.Tld, msg.Name, msg.Profile)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_profile_updated",
			sdk.NewAttribute("name", msg.Name),
			sdk.NewAttribute("tld", msg.Tld),
			sdk.NewAttribute("owner", msg.Owner),
		),
	})

	return &types.MsgUpdateProfileResponse{}, nil
}

// SetDNSRecord handles MsgSetDNSRecord
func (k Keeper) SetDNSRecord(goCtx context.Context, msg *types.MsgSetDNSRecord) (*types.MsgSetDNSRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, msg.Tld, msg.Name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", msg.Name, msg.Tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	k.SetDNSRecordEntry(ctx, msg.Tld, msg.Name, msg.Record)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_dns_record_set",
			sdk.NewAttribute("name", msg.Name+"."+msg.Tld),
			sdk.NewAttribute("value", msg.Record.Value),
		),
	})

	return &types.MsgSetDNSRecordResponse{}, nil
}

// DeleteDNSRecord handles MsgDeleteDNSRecord
func (k Keeper) DeleteDNSRecord(goCtx context.Context, msg *types.MsgDeleteDNSRecord) (*types.MsgDeleteDNSRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, msg.Tld, msg.Name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", msg.Name, msg.Tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	// Verify the DNS record exists
	if _, found := k.GetDNSRecordEntry(ctx, msg.Tld, msg.Name, msg.RecordType); !found {
		return nil, errorsmod.Wrapf(types.ErrDNSRecordNotFound, "record type %d not found for %s.%s", msg.RecordType, msg.Name, msg.Tld)
	}

	k.DeleteDNSRecordEntry(ctx, msg.Tld, msg.Name, msg.RecordType)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_dns_record_deleted",
			sdk.NewAttribute("name", msg.Name+"."+msg.Tld),
		),
	})

	return &types.MsgDeleteDNSRecordResponse{}, nil
}

// CreateTLD handles MsgCreateTLD (governance only)
func (k Keeper) CreateTLD(goCtx context.Context, msg *types.MsgCreateTLD) (*types.MsgCreateTLDResponse, error) {
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check TLD doesn't already exist
	if k.HasTLDConfig(ctx, msg.TldConfig.Tld) {
		return nil, errorsmod.Wrapf(types.ErrTLDAlreadyExists, "TLD %q already exists", msg.TldConfig.Tld)
	}

	k.SetTLDConfig(ctx, msg.TldConfig)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_tld_created",
			sdk.NewAttribute("tld", msg.TldConfig.Tld),
		),
	})

	return &types.MsgCreateTLDResponse{}, nil
}

// UpdateTLDConfig handles MsgUpdateTLDConfig (governance only)
func (k Keeper) UpdateTLDConfig(goCtx context.Context, msg *types.MsgUpdateTLDConfig) (*types.MsgUpdateTLDConfigResponse, error) {
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check TLD exists
	if !k.HasTLDConfig(ctx, msg.TldConfig.Tld) {
		return nil, errorsmod.Wrapf(types.ErrTLDNotFound, "TLD %q not found", msg.TldConfig.Tld)
	}

	k.SetTLDConfig(ctx, msg.TldConfig)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_tld_updated",
			sdk.NewAttribute("tld", msg.TldConfig.Tld),
		),
	})

	return &types.MsgUpdateTLDConfigResponse{}, nil
}

// UpdateParams handles MsgUpdateParams (governance only)
func (k Keeper) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// SetEpixNetPeer handles MsgSetEpixNetPeer
func (k Keeper) SetEpixNetPeer(goCtx context.Context, msg *types.MsgSetEpixNetPeer) (*types.MsgSetEpixNetPeerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, msg.Tld, msg.Name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", msg.Name, msg.Tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	if err := k.SetEpixNetPeerEntry(ctx, msg.Tld, msg.Name, msg.Peer); err != nil {
		return nil, err
	}

	newRoot := k.RecomputeAndStoreContentRoot(ctx, msg.Tld, msg.Name)

	events := sdk.Events{
		sdk.NewEvent(
			"xid_epixnet_peer_set",
			sdk.NewAttribute("name", msg.Name+"."+msg.Tld),
			sdk.NewAttribute("address", msg.Peer.Address),
			sdk.NewAttribute("label", msg.Peer.Label),
		),
		sdk.NewEvent(
			"xid_content_root_updated",
			sdk.NewAttribute("name", msg.Name+"."+msg.Tld),
			sdk.NewAttribute("root", newRoot),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return &types.MsgSetEpixNetPeerResponse{}, nil
}

// UpdateContentRoot is deprecated — content root is now auto-computed from active peers.
func (k Keeper) UpdateContentRoot(_ context.Context, _ *types.MsgUpdateContentRoot) (*types.MsgUpdateContentRootResponse, error) {
	return nil, errorsmod.Wrap(types.ErrInvalidContentRoot, "content root is auto-computed from active peers; manual updates are no longer supported")
}

// RevokeEpixNetPeer handles MsgRevokeEpixNetPeer
func (k Keeper) RevokeEpixNetPeer(goCtx context.Context, msg *types.MsgRevokeEpixNetPeer) (*types.MsgRevokeEpixNetPeerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, msg.Tld, msg.Name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", msg.Name, msg.Tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	// Verify the peer exists and is currently active
	peer, found := k.GetEpixNetPeerEntry(ctx, msg.Tld, msg.Name, msg.Address)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpixNetPeerNotFound, "peer %s not found for %s.%s", msg.Address, msg.Name, msg.Tld)
	}
	if !peer.Active {
		return nil, errorsmod.Wrapf(types.ErrEpixNetPeerNotFound, "peer %s is already revoked for %s.%s", msg.Address, msg.Name, msg.Tld)
	}

	if err := k.RevokeEpixNetPeerEntry(ctx, msg.Tld, msg.Name, msg.Address); err != nil {
		return nil, err
	}

	newRoot := k.RecomputeAndStoreContentRoot(ctx, msg.Tld, msg.Name)

	events := sdk.Events{
		sdk.NewEvent(
			"xid_epixnet_peer_revoked",
			sdk.NewAttribute("name", msg.Name+"."+msg.Tld),
			sdk.NewAttribute("address", msg.Address),
		),
		sdk.NewEvent(
			"xid_content_root_updated",
			sdk.NewAttribute("name", msg.Name+"."+msg.Tld),
			sdk.NewAttribute("root", newRoot),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return &types.MsgRevokeEpixNetPeerResponse{}, nil
}
