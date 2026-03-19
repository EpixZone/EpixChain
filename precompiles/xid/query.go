package xid

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Resolve handles the resolve(name, tld) view function.
// Returns the owner address for a registered name.
func (p Precompile) Resolve(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}

	record, found := p.xidKeeper.GetNameRecord(ctx, tld, name)
	if !found {
		// Return zero address if not found
		return method.Outputs.Pack(common.Address{})
	}

	ownerAddr, err := sdk.AccAddressFromBech32(record.Owner)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(common.BytesToAddress(ownerAddr.Bytes()))
}

// ReverseResolve handles the reverseResolve(addr) view function.
// Returns the primary name and TLD for an address.
func (p Precompile) ReverseResolve(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for addr: %T", args[0])
	}

	ownerAddr := sdk.AccAddress(addr.Bytes())

	// Check explicit primary name first
	if pTld, pName, found := p.xidKeeper.GetPrimaryNameEntry(ctx, ownerAddr); found {
		if record, exists := p.xidKeeper.GetNameRecord(ctx, pTld, pName); exists && record.Owner == ownerAddr.String() {
			return method.Outputs.Pack(pName, pTld)
		}
	}

	// Fall back to first name from owner index
	names := p.xidKeeper.GetNamesByOwnerAddr(ctx, ownerAddr)
	if len(names) == 0 {
		return method.Outputs.Pack("", "")
	}

	return method.Outputs.Pack(names[0].Name, names[0].Tld)
}

// GetPrimaryName handles the getPrimaryName(owner) view function.
// Returns the explicitly set primary name and TLD for an address.
func (p Precompile) GetPrimaryName(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for owner: %T", args[0])
	}

	ownerAddr := sdk.AccAddress(addr.Bytes())

	if pTld, pName, found := p.xidKeeper.GetPrimaryNameEntry(ctx, ownerAddr); found {
		if record, exists := p.xidKeeper.GetNameRecord(ctx, pTld, pName); exists && record.Owner == ownerAddr.String() {
			return method.Outputs.Pack(pName, pTld)
		}
	}

	return method.Outputs.Pack("", "")
}

// GetProfile handles the getProfile(name, tld) view function.
// Returns the avatar and bio for a name.
func (p Precompile) GetProfile(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}

	profile, found := p.xidKeeper.GetProfileRecord(ctx, tld, name)
	if !found {
		return method.Outputs.Pack("", "")
	}

	return method.Outputs.Pack(profile.Avatar, profile.Bio)
}

// GetDNSRecord handles the getDNSRecord(name, tld, recordType) view function.
// Returns the value and TTL for a specific DNS record.
func (p Precompile) GetDNSRecord(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("expected 3 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}
	recordType, ok := args[2].(uint16)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for recordType: %T", args[2])
	}

	record, found := p.xidKeeper.GetDNSRecordEntry(ctx, tld, name, uint32(recordType))
	if !found {
		return method.Outputs.Pack("", uint32(0))
	}

	return method.Outputs.Pack(record.Value, record.Ttl)
}

// GetRegistrationFee handles the getRegistrationFee(name, tld) view function.
// Returns the fee amount in aepix as a uint256.
func (p Precompile) GetRegistrationFee(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}

	fee, err := p.xidKeeper.CalculateRegistrationFee(ctx, tld, name)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(fee.Amount.BigInt())
}

// GetLinkedIdentities handles the getLinkedIdentities(name, tld) view function.
// Returns arrays of identity addresses and labels.
func (p Precompile) GetLinkedIdentities(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}

	identities := p.xidKeeper.GetAllLinkedIdentities(ctx, tld, name)

	addresses := make([]string, len(identities))
	labels := make([]string, len(identities))
	addedAts := make([]uint64, len(identities))
	actives := make([]bool, len(identities))
	revokedAts := make([]uint64, len(identities))
	revokedAtTimes := make([]int64, len(identities))
	for i, identity := range identities {
		addresses[i] = identity.Address
		labels[i] = identity.Label
		addedAts[i] = identity.AddedAt
		actives[i] = identity.Active
		revokedAts[i] = identity.RevokedAt
		revokedAtTimes[i] = identity.RevokedAtTime
	}

	return method.Outputs.Pack(addresses, labels, addedAts, actives, revokedAts, revokedAtTimes)
}

// GetContentRoot handles the getContentRoot(name, tld) view function.
// Returns the content root hash and the block height when it was last updated.
func (p Precompile) GetContentRoot(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}

	root, found := p.xidKeeper.GetContentRoot(ctx, tld, name)
	if !found {
		return method.Outputs.Pack("", uint64(0))
	}

	return method.Outputs.Pack(root.Root, root.UpdatedAt)
}

// ReverseResolveBech32 handles the reverseResolveBech32(bech32Addr) view function.
// Accepts a bech32 string (e.g. "epix1...") and returns the primary name and TLD.
func (p Precompile) ReverseResolveBech32(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	bech32Addr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for bech32Addr: %T", args[0])
	}

	ownerAddr, err := sdk.AccAddressFromBech32(bech32Addr)
	if err != nil {
		return nil, fmt.Errorf("invalid bech32 address: %s", bech32Addr)
	}

	// Check explicit primary name first
	if pTld, pName, found := p.xidKeeper.GetPrimaryNameEntry(ctx, ownerAddr); found {
		if record, exists := p.xidKeeper.GetNameRecord(ctx, pTld, pName); exists && record.Owner == ownerAddr.String() {
			return method.Outputs.Pack(pName, pTld)
		}
	}

	// Fall back to first name from owner index
	names := p.xidKeeper.GetNamesByOwnerAddr(ctx, ownerAddr)
	if len(names) == 0 {
		return method.Outputs.Pack("", "")
	}

	return method.Outputs.Pack(names[0].Name, names[0].Tld)
}

// GetStateDigest handles the getStateDigest() view function.
// Returns the current state digest, block height, and number of names.
func (p Precompile) GetStateDigest(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	digest, found := p.xidKeeper.GetStateDigest(ctx)
	if !found {
		return method.Outputs.Pack("", uint64(0), uint64(0))
	}

	return method.Outputs.Pack(digest.Digest, digest.Height, digest.NumNames)
}

// GetAttestations handles the getAttestations() view function.
// Returns all attestations for the current digest.
func (p Precompile) GetAttestations(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	digest, found := p.xidKeeper.GetStateDigest(ctx)
	if !found {
		return method.Outputs.Pack([]string{}, []string{}, []uint64{}, false)
	}

	attestations := p.xidKeeper.GetAttestations(ctx, digest.Digest)
	finalized := p.xidKeeper.IsDigestFinalized(ctx, digest.Digest)

	validators := make([]string, len(attestations))
	signatures := make([]string, len(attestations))
	heights := make([]uint64, len(attestations))

	for i, att := range attestations {
		validators[i] = att.ValidatorAddr
		signatures[i] = att.Signature
		heights[i] = att.Height
	}

	return method.Outputs.Pack(validators, signatures, heights, finalized)
}

// ReverseResolveByIdentity handles the reverseResolveByIdentity(identityAddress) view function.
// Looks up the xID name associated with a linked identity address.
func (p Precompile) ReverseResolveByIdentity(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	identityAddress, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for identityAddress: %T", args[0])
	}

	tld, name, found := p.xidKeeper.GetLinkedIdentityOwner(ctx, identityAddress)
	if !found {
		return method.Outputs.Pack("", "", false)
	}

	if !p.xidKeeper.HasNameRecord(ctx, tld, name) {
		return method.Outputs.Pack("", "", false)
	}

	return method.Outputs.Pack(name, tld, true)
}

// ensure big.Int is used
var _ = (*big.Int)(nil)
