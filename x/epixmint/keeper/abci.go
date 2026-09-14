package keeper

import (
	"context"

	"github.com/cosmos/evm/x/epixmint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker mints new tokens for the previous block.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't crash the chain
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Error("panic in epixmint BeginBlocker", "error", r)
		}
	}()

	return k.MintCoins(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx context.Context) error {
	params := k.GetParams(ctx)

	// Check if we've reached the maximum supply
	currentSupply := k.bankKeeper.GetSupply(ctx, params.MintDenom)
	if currentSupply.Amount.GTE(params.MaxSupply) {
		// Maximum supply reached, stop minting
		return nil
	}

	// Calculate tokens to mint this block using dynamic emission rate
	tokensPerBlock := calculateCurrentEmissionRate(ctx, params)

	// Ensure we don't exceed max supply
	if currentSupply.Amount.Add(tokensPerBlock).GT(params.MaxSupply) {
		// Only mint what's needed to reach max supply
		tokensPerBlock = params.MaxSupply.Sub(currentSupply.Amount)
	}

	// Skip minting if amount is zero or negative
	if tokensPerBlock.IsZero() || tokensPerBlock.IsNegative() {
		return nil
	}

	mintedCoin := sdk.NewCoin(params.MintDenom, tokensPerBlock)
	mintedCoins := sdk.NewCoins(mintedCoin)

	// Mint coins to the epixmint module account
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintedCoins)
	if err != nil {
		return err
	}

	// Distribute minted tokens according to configured rates
	err = k.distributeMintedTokens(ctx, mintedCoins, params)
	if err != nil {
		return err
	}

	// Emit mint event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyAmount, tokensPerBlock.String()),
		),
	)

	return nil
}

// distributeMintedTokens distributes the minted tokens according to the configured rates
func (k Keeper) distributeMintedTokens(ctx context.Context, mintedCoins sdk.Coins, params types.Params) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Calculate distribution amounts using both configured rates
	totalAmount := mintedCoins[0].Amount
	communityPoolAmount := params.CommunityPoolRate.MulInt(totalAmount).TruncateInt()
	stakingRewardsAmount := params.StakingRewardsRate.MulInt(totalAmount).TruncateInt()

	// Assign any truncation remainder to staking rewards to avoid rounding loss
	distributedAmount := communityPoolAmount.Add(stakingRewardsAmount)
	if distributedAmount.LT(totalAmount) {
		stakingRewardsAmount = stakingRewardsAmount.Add(totalAmount.Sub(distributedAmount))
	} else if distributedAmount.GT(totalAmount) {
		// Adjust staking rewards down if rounding caused over-distribution
		stakingRewardsAmount = totalAmount.Sub(communityPoolAmount)
	}

	// 1. Send to community pool
	if communityPoolAmount.IsPositive() {
		communityPoolCoins := sdk.NewCoins(sdk.NewCoin(params.MintDenom, communityPoolAmount))

		// Get the epixmint module address to send from
		epixmintModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

		err := k.distributionKeeper.FundCommunityPool(ctx, communityPoolCoins, epixmintModuleAddr)
		if err != nil {
			return err
		}

		// Emit community pool funding event
		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute("community_pool_amount", communityPoolAmount.String()),
			),
		)
	}

	// 2. Distribute staking rewards to validators
	if stakingRewardsAmount.IsPositive() {
		stakingRewardsCoins := sdk.NewCoins(sdk.NewCoin(params.MintDenom, stakingRewardsAmount))

		bondedValidators, totalVotingPower, err := k.getBondedValidators(ctx)
		if err != nil {
			return err
		}

		// If there are no bonded validators, route the staking share to the
		// community pool. This must happen before the transfer to the
		// distribution module below, since the epixmint module account is the
		// one still holding the coins at this point.
		if len(bondedValidators) == 0 || totalVotingPower.IsZero() {
			epixmintModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
			if err := k.distributionKeeper.FundCommunityPool(ctx, stakingRewardsCoins, epixmintModuleAddr); err != nil {
				return err
			}

			sdkCtx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeMint,
					sdk.NewAttribute("staking_rewards_to_community_pool", stakingRewardsAmount.String()),
				),
			)
			return nil
		}

		// Send to distribution module for validator rewards
		err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, "distribution", stakingRewardsCoins)
		if err != nil {
			return err
		}

		// Allocate tokens to all bonded validators
		err = k.allocateTokensToValidators(ctx, bondedValidators, totalVotingPower, sdk.NewDecCoinsFromCoins(stakingRewardsCoins...))
		if err != nil {
			return err
		}

		// Emit staking rewards event
		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute("staking_rewards_amount", stakingRewardsAmount.String()),
			),
		)
	}

	return nil
}

// getBondedValidators returns the bonded validators and their total voting power.
func (k Keeper) getBondedValidators(ctx context.Context) ([]stakingtypes.Validator, math.Int, error) {
	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, math.ZeroInt(), err
	}

	bondedValidators := make([]stakingtypes.Validator, 0, len(validators))
	totalVotingPower := math.ZeroInt()

	for _, validator := range validators {
		if validator.IsBonded() {
			bondedValidators = append(bondedValidators, validator)
			totalVotingPower = totalVotingPower.Add(validator.GetTokens())
		}
	}

	return bondedValidators, totalVotingPower, nil
}

// allocateTokensToValidators allocates tokens to the given bonded validators
// proportionally to their voting power. The caller must ensure the tokens have
// already been transferred to the distribution module and that
// totalVotingPower is non-zero.
func (k Keeper) allocateTokensToValidators(ctx context.Context, bondedValidators []stakingtypes.Validator, totalVotingPower math.Int, tokens sdk.DecCoins) error {
	totalPower := math.LegacyNewDecFromInt(totalVotingPower)

	for _, validator := range bondedValidators {
		// Calculate validator's share based on voting power
		validatorPower := math.LegacyNewDecFromInt(validator.GetTokens())
		validatorShare := validatorPower.Quo(totalPower)

		// Calculate validator's allocation
		validatorTokens := make(sdk.DecCoins, len(tokens))
		for i, token := range tokens {
			validatorTokens[i] = sdk.NewDecCoinFromDec(token.Denom, token.Amount.Mul(validatorShare))
		}

		// Allocate tokens to validator
		err := k.distributionKeeper.AllocateTokensToValidator(ctx, validator, validatorTokens)
		if err != nil {
			return err
		}
	}

	return nil
}
