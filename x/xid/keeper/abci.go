package keeper

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/evm/x/xid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlock auto-attests the current state digest for all bonded validators
// who committed the previous block. Consensus itself is the attestation —
// no cryptographic signature is needed because the block proves participation.
func (k Keeper) BeginBlock(ctx sdk.Context) error {
	// Stamp the current digest with this block's height+time every block — so a
	// STABLE digest (no name changes) still carries a fresh block_time for the
	// signed attestations and the client's freshness check. Runs unconditionally
	// (even before attestation is enabled) so the digest is ready when it turns on.
	if sd, found := k.GetStateDigest(ctx); found {
		sd.Height = uint64(ctx.BlockHeight())
		sd.BlockTime = ctx.BlockTime().Unix()
		k.SetStateDigest(ctx, sd)
	}

	config := k.GetAttestationConfig(ctx)
	if !config.Enabled {
		return nil
	}

	currentDigest, found := k.GetStateDigest(ctx)
	if !found {
		return nil
	}

	for _, voteInfo := range ctx.VoteInfos() {
		if voteInfo.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}

		consAddr := sdk.ConsAddress(voteInfo.Validator.Address)
		validator, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
		if err != nil {
			continue
		}
		if validator.GetStatus() != stakingtypes.Bonded {
			continue
		}
		// auto:consensus is the legacy count-based path, kept only so existing
		// clients keep working BEFORE vote extensions are enabled. Once enabled,
		// every validator's signed (power-bearing) attestation is written by
		// PreBlocker and the power path in IsDigestFinalized takes precedence; the
		// zero-power auto:consensus entry (keyed by account addr, distinct from the
		// valcons-keyed signed entry) coexists harmlessly and the client ignores it.

		valOperAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			continue
		}
		accAddr := sdk.AccAddress(valOperAddr.Bytes())
		addrStr := accAddr.String()

		if k.HasAttestation(ctx, currentDigest.Digest, addrStr) {
			continue
		}

		att := types.Attestation{
			ValidatorAddr: addrStr,
			Digest:        currentDigest.Digest,
			Signature:     "auto:consensus",
			Height:        uint64(ctx.BlockHeight()),
		}
		k.SetAttestation(ctx, att)
		k.IncrementAttestationCount(ctx, currentDigest.Digest)
	}

	return nil
}
