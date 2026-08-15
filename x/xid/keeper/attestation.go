package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/x/xid/types"
)

// SubmitAttestation validates and stores an attestation from a validator.
func (k Keeper) SubmitAttestation(ctx sdk.Context, msg *types.MsgAttestStateDigest) error {
	// Check attestation is enabled
	config := k.GetAttestationConfig(ctx)
	if !config.Enabled {
		return types.ErrAttestationDisabled
	}

	// Verify the digest matches current state digest
	currentDigest, found := k.GetStateDigest(ctx)
	if !found {
		return errorsmod.Wrap(types.ErrDigestMismatch, "no state digest computed yet")
	}
	if msg.Digest != currentDigest.Digest {
		return errorsmod.Wrapf(types.ErrDigestMismatch, "expected %s, got %s", currentDigest.Digest, msg.Digest)
	}

	// Verify signer is an active bonded validator
	if !k.isActiveBondedValidator(ctx, msg.Signer) {
		return errorsmod.Wrapf(types.ErrNotValidator, "%s is not an active bonded validator", msg.Signer)
	}

	// Check for duplicate attestation
	if k.HasAttestation(ctx, msg.Digest, msg.Signer) {
		return errorsmod.Wrapf(types.ErrAlreadyAttested, "validator %s already attested to digest %s", msg.Signer, msg.Digest)
	}

	// Store the attestation
	att := types.Attestation{
		ValidatorAddr: msg.Signer,
		Digest:        msg.Digest,
		Signature:     msg.Signature,
		Height:        uint64(ctx.BlockHeight()),
	}
	k.SetAttestation(ctx, att)
	k.IncrementAttestationCount(ctx, msg.Digest)

	return nil
}

// isActiveBondedValidator checks if the given account address corresponds to
// an active bonded validator operator.
// addr is a bech32 account address (epix1...) while val.OperatorAddress uses
// the valoper prefix (epixvaloper1...). They share the same 20-byte key, so
// we convert the account address bytes to a ValAddress for comparison.
func (k Keeper) isActiveBondedValidator(ctx sdk.Context, addr string) bool {
	accAddr, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return false
	}
	// Convert account address bytes to validator operator address string
	valAddr := sdk.ValAddress(accAddr.Bytes()).String()

	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return false
	}

	for _, val := range validators {
		if val.OperatorAddress == valAddr && val.GetStatus() == stakingtypes.Bonded {
			return true
		}
	}
	return false
}

// SetAttestation stores an attestation.
func (k Keeper) SetAttestation(ctx sdk.Context, att types.Attestation) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(att)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal attestation: %v", err))
	}
	store.Set(types.AttestationKey(att.Digest, att.ValidatorAddr), bz)
}

// HasAttestation checks if a validator has already attested to a digest.
func (k Keeper) HasAttestation(ctx sdk.Context, digest, validatorAddr string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.AttestationKey(digest, validatorAddr))
}

// GetAttestations returns all attestations for a given digest.
func (k Keeper) GetAttestations(ctx sdk.Context, digest string) []types.Attestation {
	store := ctx.KVStore(k.storeKey)
	pfx := types.AttestationPrefix(digest)
	iterator := storetypes.KVStorePrefixIterator(store, pfx)
	defer iterator.Close()

	var attestations []types.Attestation
	for ; iterator.Valid(); iterator.Next() {
		var att types.Attestation
		if err := json.Unmarshal(iterator.Value(), &att); err != nil {
			continue
		}
		attestations = append(attestations, att)
	}
	return attestations
}

// GetAttestationCount returns the number of attestations for a digest.
func (k Keeper) GetAttestationCount(ctx sdk.Context, digest string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AttestationCountKey(digest))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// IncrementAttestationCount increments the attestation count for a digest.
func (k Keeper) IncrementAttestationCount(ctx sdk.Context, digest string) {
	count := k.GetAttestationCount(ctx, digest) + 1
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(types.AttestationCountKey(digest), bz)
}

// IsDigestFinalized checks if a digest has been attested by >= 2/3 of bonded validators.
// If config.Threshold is set (non-zero), it is used as an override instead.
func (k Keeper) IsDigestFinalized(ctx sdk.Context, digest string) bool {
	config := k.GetAttestationConfig(ctx)
	if !config.Enabled {
		return false
	}

	count := k.GetAttestationCount(ctx, digest)

	// Use explicit threshold override if set
	if config.Threshold > 0 {
		return count >= config.Threshold
	}

	// Default: require 2/3+ of bonded validators
	bondedCount := k.countBondedValidators(ctx)
	if bondedCount == 0 {
		return false
	}
	// 2/3 threshold: count * 3 > bondedCount * 2 (avoids floating point)
	return count*3 > bondedCount*2
}

// countBondedValidators returns the number of bonded validators.
func (k Keeper) countBondedValidators(ctx sdk.Context) uint64 {
	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return 0
	}
	var bonded uint64
	for _, val := range validators {
		if val.GetStatus() == stakingtypes.Bonded {
			bonded++
		}
	}
	return bonded
}

// ClearAttestationsForDigest removes all attestations and the count for a digest.
func (k Keeper) ClearAttestationsForDigest(ctx sdk.Context, digest string) {
	store := ctx.KVStore(k.storeKey)

	// Delete all attestation entries
	pfx := types.AttestationPrefix(digest)
	iterator := storetypes.KVStorePrefixIterator(store, pfx)
	defer iterator.Close()

	var keysToDelete [][]byte
	for ; iterator.Valid(); iterator.Next() {
		keysToDelete = append(keysToDelete, iterator.Key())
	}
	for _, key := range keysToDelete {
		store.Delete(key)
	}

	// Delete the count
	store.Delete(types.AttestationCountKey(digest))
}

// GetAttestationConfig retrieves the attestation configuration.
func (k Keeper) GetAttestationConfig(ctx sdk.Context) types.AttestationConfig {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AttestationConfigKey())
	if bz == nil {
		return types.DefaultAttestationConfig()
	}
	var config types.AttestationConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return types.DefaultAttestationConfig()
	}
	return config
}

// SetAttestationConfig stores the attestation configuration.
func (k Keeper) SetAttestationConfig(ctx sdk.Context, config types.AttestationConfig) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal attestation config: %v", err))
	}
	store.Set(types.AttestationConfigKey(), bz)
}

// ---------------------------------------------------------------------------
// Attestation keys (valcons -> ed25519 attestation pubkey)
// ---------------------------------------------------------------------------

// SubmitAttestKey validates a MsgRegisterAttestKey and stores the binding. The
// signer's operator/account key must own the consensus address, and the pubkey
// must be a 32-byte hex ed25519 key. This is the key the chain signs digests with
// (via vote extensions) and that light clients pin — the consensus key is never
// reused for app-level attestations.
func (k Keeper) SubmitAttestKey(ctx sdk.Context, msg *types.MsgRegisterAttestKey) error {
	pk, err := hex.DecodeString(msg.Ed25519Pubkey)
	if err != nil || len(pk) != 32 {
		return errorsmod.Wrap(types.ErrInvalidAttestKey, "ed25519_pubkey must be 32-byte hex")
	}
	consAddr, err := sdk.ConsAddressFromBech32(msg.ValidatorConsAddr)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidAttestKey, "invalid validator_cons_addr")
	}
	validator, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return errorsmod.Wrapf(types.ErrNotValidator, "no validator for consensus address %s", msg.ValidatorConsAddr)
	}
	valOper, err := sdk.ValAddressFromBech32(validator.GetOperator())
	if err != nil {
		return err
	}
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return err
	}
	// Operator address and account address share the same 20 key bytes.
	if !sdk.AccAddress(valOper.Bytes()).Equals(signer) {
		return errorsmod.Wrap(types.ErrInvalidAttestKey, "signer is not the validator operator for that consensus address")
	}
	k.SetAttestKey(ctx, msg.ValidatorConsAddr, msg.Ed25519Pubkey)
	return nil
}

// SetAttestKey stores a validator's registered ed25519 attestation pubkey (hex).
func (k Keeper) SetAttestKey(ctx sdk.Context, valcons, pubkeyHex string) {
	ctx.KVStore(k.storeKey).Set(types.AttestKeyKey(valcons), []byte(pubkeyHex))
}

// GetAttestKey returns a validator's registered attestation pubkey (hex), if any.
func (k Keeper) GetAttestKey(ctx sdk.Context, valcons string) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.AttestKeyKey(valcons))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// totalBondedPower returns the total consensus voting power of all bonded
// validators — the denominator light clients verify signed power against.
func (k Keeper) totalBondedPower(ctx sdk.Context) uint64 {
	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return 0
	}
	var total int64
	for _, val := range validators {
		if val.GetStatus() == stakingtypes.Bonded {
			total += val.GetConsensusPower(sdk.DefaultPowerReduction)
		}
	}
	if total < 0 {
		return 0
	}
	return uint64(total)
}
