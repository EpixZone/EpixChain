package keeper

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/xid/types"
)

// CalculateRegistrationFee calculates the fee for registering a name under a TLD
func (k Keeper) CalculateRegistrationFee(ctx sdk.Context, tld, name string) (sdk.Coin, error) {
	tldConfig, found := k.GetTLDConfig(ctx, tld)
	if !found {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrTLDNotFound, "TLD %q not found", tld)
	}

	params := k.GetParams(ctx)
	nameLen := uint32(len(name))

	// Find the matching price tier (tiers are sorted by max_length ascending)
	for _, tier := range tldConfig.PriceTiers {
		if nameLen <= tier.MaxLength {
			return sdk.NewCoin(params.FeeDenom, tier.Price), nil
		}
	}

	// If no tier matches, use the last tier (catch-all for longest names)
	if len(tldConfig.PriceTiers) > 0 {
		lastTier := tldConfig.PriceTiers[len(tldConfig.PriceTiers)-1]
		return sdk.NewCoin(params.FeeDenom, lastTier.Price), nil
	}

	return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidPriceTier, "no price tiers configured for TLD %q", tld)
}

// RegisterNameRecord registers a new name under a TLD, charges the fee, and burns it
func (k Keeper) RegisterNameRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, tld, name string) error {
	// Normalize
	name = strings.ToLower(name)
	tld = strings.ToLower(tld)

	// Validate name
	if err := types.ValidateName(name); err != nil {
		return err
	}

	// Validate name length against params
	params := k.GetParams(ctx)
	if uint32(len(name)) < params.MinNameLength {
		return errorsmod.Wrapf(types.ErrInvalidName, "name must be at least %d characters", params.MinNameLength)
	}
	if uint32(len(name)) > params.MaxNameLength {
		return errorsmod.Wrapf(types.ErrInvalidName, "name must be at most %d characters", params.MaxNameLength)
	}

	// Check TLD exists and is enabled
	tldConfig, found := k.GetTLDConfig(ctx, tld)
	if !found {
		return errorsmod.Wrapf(types.ErrTLDNotFound, "TLD %q not found", tld)
	}
	if !tldConfig.Enabled {
		return errorsmod.Wrapf(types.ErrTLDDisabled, "TLD %q is disabled", tld)
	}

	// Check name is not already taken
	if k.HasNameRecord(ctx, tld, name) {
		return errorsmod.Wrapf(types.ErrNameTaken, "%s.%s is already registered", name, tld)
	}

	// Calculate and charge the registration fee
	fee, err := k.CalculateRegistrationFee(ctx, tld, name)
	if err != nil {
		return err
	}

	feeCoins := sdk.NewCoins(fee)

	// Send fee from user to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAddr, types.ModuleName, feeCoins); err != nil {
		return errorsmod.Wrapf(types.ErrInsufficientFee, "failed to charge registration fee: %s", err)
	}

	// Burn the fee
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, feeCoins); err != nil {
		return errorsmod.Wrap(err, "failed to burn registration fee")
	}

	// Store the name record
	record := types.NameRecord{
		Name:         name,
		Tld:          tld,
		Owner:        ownerAddr.String(),
		RegisteredAt: ctx.BlockHeight(),
	}
	k.SetNameRecord(ctx, record)

	// Set owner index for reverse lookups
	k.SetOwnerIndex(ctx, ownerAddr, tld, name)

	// Increment owner's name count
	k.IncrementOwnerCount(ctx, ownerAddr)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_name_registered",
			sdk.NewAttribute("name", name),
			sdk.NewAttribute("tld", tld),
			sdk.NewAttribute("owner", ownerAddr.String()),
			sdk.NewAttribute("fee", fee.String()),
		),
	})

	k.Logger(ctx).Info("name registered",
		"name", name+"."+tld,
		"owner", ownerAddr.String(),
		"fee", fee.String(),
	)

	return nil
}

// TransferNameRecord transfers ownership of a name to a new address
func (k Keeper) TransferNameRecord(ctx sdk.Context, currentOwner, newOwner sdk.AccAddress, tld, name string) error {
	// Normalize
	name = strings.ToLower(name)
	tld = strings.ToLower(tld)

	// Get the name record
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return errorsmod.Wrapf(types.ErrNameNotFound, "%s.%s not found", name, tld)
	}

	// Verify ownership
	if record.Owner != currentOwner.String() {
		return errorsmod.Wrapf(types.ErrNotOwner, "sender %s is not the owner of %s.%s", currentOwner, name, tld)
	}

	// Update owner index
	k.DeleteOwnerIndex(ctx, currentOwner, tld, name)
	k.SetOwnerIndex(ctx, newOwner, tld, name)

	// Update owner counts
	k.DecrementOwnerCount(ctx, currentOwner)
	k.IncrementOwnerCount(ctx, newOwner)

	// Update the name record
	record.Owner = newOwner.String()
	k.SetNameRecord(ctx, record)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"xid_name_transferred",
			sdk.NewAttribute("name", name),
			sdk.NewAttribute("tld", tld),
			sdk.NewAttribute("from", currentOwner.String()),
			sdk.NewAttribute("to", newOwner.String()),
		),
	})

	k.Logger(ctx).Info("name transferred",
		"name", name+"."+tld,
		"from", currentOwner.String(),
		"to", newOwner.String(),
	)

	return nil
}
