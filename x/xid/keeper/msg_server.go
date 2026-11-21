package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/evm/x/xid/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ types.MsgServer = &Keeper{}

// RegisterName handles MsgRegisterName
func (k Keeper) RegisterName(goCtx context.Context, msg *types.MsgRegisterName) (*types.MsgRegisterNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if err := k.RegisterNameRecord(ctx, ownerAddr, tld, name); err != nil {
		return nil, err
	}

	k.UpdateDomainInTree(ctx, tld, name)

	return &types.MsgRegisterNameResponse{}, nil
}

// TransferName handles MsgTransferName
func (k Keeper) TransferName(goCtx context.Context, msg *types.MsgTransferName) (*types.MsgTransferNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	currentOwner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	newOwner, err := sdk.AccAddressFromBech32(msg.NewOwner)
	if err != nil {
		return nil, err
	}

	if err := k.TransferNameRecord(ctx, currentOwner, newOwner, tld, name); err != nil {
		return nil, err
	}

	k.UpdateDomainInTree(ctx, tld, name)

	return &types.MsgTransferNameResponse{}, nil
}

// UpdateProfile handles MsgUpdateProfile
func (k Keeper) UpdateProfile(goCtx context.Context, msg *types.MsgUpdateProfile) (*types.MsgUpdateProfileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", name, tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	k.SetProfileRecord(ctx, tld, name, msg.Profile)
	k.UpdateDomainInTree(ctx, tld, name)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_profile_updated",
			sdk.NewAttribute("name", name),
			sdk.NewAttribute("tld", tld),
			sdk.NewAttribute("owner", msg.Owner),
		),
	})

	return &types.MsgUpdateProfileResponse{}, nil
}

// SetDNSRecord handles MsgSetDNSRecord
func (k Keeper) SetDNSRecord(goCtx context.Context, msg *types.MsgSetDNSRecord) (*types.MsgSetDNSRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", name, tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	k.SetDNSRecordEntry(ctx, tld, name, msg.Record)
	k.UpdateDomainInTree(ctx, tld, name)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_dns_record_set",
			sdk.NewAttribute("name", name+"."+tld),
			sdk.NewAttribute("value", msg.Record.Value),
		),
	})

	return &types.MsgSetDNSRecordResponse{}, nil
}

// DeleteDNSRecord handles MsgDeleteDNSRecord
func (k Keeper) DeleteDNSRecord(goCtx context.Context, msg *types.MsgDeleteDNSRecord) (*types.MsgDeleteDNSRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", name, tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	// Verify the DNS record exists
	if _, found := k.GetDNSRecordEntry(ctx, tld, name, msg.RecordType); !found {
		return nil, errorsmod.Wrapf(types.ErrDNSRecordNotFound, "record type %d not found for %s.%s", msg.RecordType, name, tld)
	}

	k.DeleteDNSRecordEntry(ctx, tld, name, msg.RecordType)
	k.UpdateDomainInTree(ctx, tld, name)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_dns_record_deleted",
			sdk.NewAttribute("name", name+"."+tld),
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
	tldConfig := msg.TldConfig
	tldConfig.Tld = types.NormalizeTLD(tldConfig.Tld)

	// Check TLD doesn't already exist
	if k.HasTLDConfig(ctx, tldConfig.Tld) {
		return nil, errorsmod.Wrapf(types.ErrTLDAlreadyExists, "TLD %q already exists", tldConfig.Tld)
	}

	k.SetTLDConfig(ctx, tldConfig)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_tld_created",
			sdk.NewAttribute("tld", tldConfig.Tld),
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
	tldConfig := msg.TldConfig
	tldConfig.Tld = types.NormalizeTLD(tldConfig.Tld)

	// Check TLD exists
	if !k.HasTLDConfig(ctx, tldConfig.Tld) {
		return nil, errorsmod.Wrapf(types.ErrTLDNotFound, "TLD %q not found", tldConfig.Tld)
	}

	k.SetTLDConfig(ctx, tldConfig)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_tld_updated",
			sdk.NewAttribute("tld", tldConfig.Tld),
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

// LinkIdentity handles MsgLinkIdentity
func (k Keeper) LinkIdentity(goCtx context.Context, msg *types.MsgLinkIdentity) (*types.MsgLinkIdentityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", name, tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	if err := k.SetLinkedIdentityEntry(ctx, tld, name, msg.Identity); err != nil {
		return nil, err
	}

	newRoot := k.RecomputeAndStoreContentRoot(ctx, tld, name)
	k.UpdateDomainInTree(ctx, tld, name)

	events := sdk.Events{
		sdk.NewEvent(
			"xid_identity_linked",
			sdk.NewAttribute("name", name+"."+tld),
			sdk.NewAttribute("address", msg.Identity.Address),
			sdk.NewAttribute("label", msg.Identity.Label),
		),
		sdk.NewEvent(
			"xid_content_root_updated",
			sdk.NewAttribute("name", name+"."+tld),
			sdk.NewAttribute("root", newRoot),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return &types.MsgLinkIdentityResponse{}, nil
}

// UpdateContentRoot is deprecated — content root is now auto-computed from active identities.
func (k Keeper) UpdateContentRoot(_ context.Context, _ *types.MsgUpdateContentRoot) (*types.MsgUpdateContentRootResponse, error) {
	return nil, errorsmod.Wrap(types.ErrInvalidContentRoot, "content root is auto-computed from active identities; manual updates are no longer supported")
}

// UnlinkIdentity handles MsgUnlinkIdentity
func (k Keeper) UnlinkIdentity(goCtx context.Context, msg *types.MsgUnlinkIdentity) (*types.MsgUnlinkIdentityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", name, tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	// Verify the linked identity exists and is currently active
	identity, found := k.GetLinkedIdentityEntry(ctx, tld, name, msg.Address)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrIdentityNotFound, "identity %s not found for %s.%s", msg.Address, name, tld)
	}
	if !identity.Active {
		return nil, errorsmod.Wrapf(types.ErrIdentityNotFound, "identity %s is already unlinked for %s.%s", msg.Address, name, tld)
	}

	if err := k.RevokeLinkedIdentityEntry(ctx, tld, name, msg.Address); err != nil {
		return nil, err
	}

	newRoot := k.RecomputeAndStoreContentRoot(ctx, tld, name)
	k.UpdateDomainInTree(ctx, tld, name)

	events := sdk.Events{
		sdk.NewEvent(
			"xid_identity_unlinked",
			sdk.NewAttribute("name", name+"."+tld),
			sdk.NewAttribute("address", msg.Address),
		),
		sdk.NewEvent(
			"xid_content_root_updated",
			sdk.NewAttribute("name", name+"."+tld),
			sdk.NewAttribute("root", newRoot),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return &types.MsgUnlinkIdentityResponse{}, nil
}

// SetPrimaryName handles MsgSetPrimaryName
func (k Keeper) SetPrimaryName(goCtx context.Context, msg *types.MsgSetPrimaryName) (*types.MsgSetPrimaryNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	name, tld := types.NormalizeName(msg.Name), types.NormalizeTLD(msg.Tld)

	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	// Verify name exists and caller is owner
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", name, tld)
	}
	if record.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner", msg.Owner)
	}

	k.SetPrimaryNameEntry(ctx, ownerAddr, tld, name)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_primary_name_set",
			sdk.NewAttribute("name", name),
			sdk.NewAttribute("tld", tld),
			sdk.NewAttribute("owner", msg.Owner),
		),
	})

	return &types.MsgSetPrimaryNameResponse{}, nil
}

// AttestStateDigest handles MsgAttestStateDigest
func (k Keeper) AttestStateDigest(goCtx context.Context, msg *types.MsgAttestStateDigest) (*types.MsgAttestStateDigestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.SubmitAttestation(ctx, msg); err != nil {
		return nil, err
	}

	count := k.GetAttestationCount(ctx, msg.Digest)
	finalized := k.IsDigestFinalized(ctx, msg.Digest)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_attestation_submitted",
			sdk.NewAttribute("validator", msg.Signer),
			sdk.NewAttribute("digest", msg.Digest),
			sdk.NewAttribute("count", fmt.Sprintf("%d", count)),
			sdk.NewAttribute("finalized", fmt.Sprintf("%t", finalized)),
		),
	})

	return &types.MsgAttestStateDigestResponse{}, nil
}
