package xid

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/cosmos/evm/x/xid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Register handles the register(name, tld) function.
// The registration fee is deducted directly from the caller's bank balance
// and burned via the xID module account.
func (p Precompile) Register(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
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

	caller := contract.Caller()
	callerAddr := sdk.AccAddress(caller.Bytes())

	// Delegate to the keeper which handles validation, fee collection
	// (SendCoinsFromAccountToModule), burning (BurnCoins), and storage.
	if err := p.xidKeeper.RegisterNameRecord(ctx, callerAddr, tld, name); err != nil {
		return nil, err
	}

	p.xidKeeper.RecomputeAndStoreStateDigest(ctx)

	// Emit EVM event
	if err := p.EmitNameRegistered(ctx, stateDB, caller, name, tld); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// TransferName handles the transferName(name, tld, newOwner) function.
func (p Precompile) TransferName(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
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
	newOwnerAddr, ok := args[2].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for newOwner: %T", args[2])
	}

	caller := contract.Caller()
	callerAddr := sdk.AccAddress(caller.Bytes())
	newOwner := sdk.AccAddress(newOwnerAddr.Bytes())

	if err := p.xidKeeper.TransferNameRecord(ctx, callerAddr, newOwner, tld, name); err != nil {
		return nil, err
	}

	p.xidKeeper.RecomputeAndStoreStateDigest(ctx)

	// Emit EVM event
	if err := p.EmitNameTransferred(ctx, stateDB, caller, newOwnerAddr, name, tld); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// UpdateProfile handles the updateProfile(name, tld, avatar, bio) function.
func (p Precompile) UpdateProfile(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("expected 4 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}
	avatar, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for avatar: %T", args[2])
	}
	bio, ok := args[3].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for bio: %T", args[3])
	}

	caller := contract.Caller()

	// Verify ownership
	record, found := p.xidKeeper.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, fmt.Errorf("%s.%s not found", name, tld)
	}

	callerAddr := sdk.AccAddress(caller.Bytes())
	if record.Owner != callerAddr.String() {
		return nil, fmt.Errorf("caller is not the owner of %s.%s", name, tld)
	}

	profile := types.Profile{
		Avatar: avatar,
		Bio:    bio,
	}
	p.xidKeeper.SetProfileRecord(ctx, tld, name, profile)

	p.xidKeeper.RecomputeAndStoreStateDigest(ctx)

	// Emit EVM event
	if err := p.EmitProfileUpdated(ctx, stateDB, caller, name, tld); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// SetDNSRecord handles the setDNSRecord(name, tld, recordType, value, ttl) function.
func (p Precompile) SetDNSRecord(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 5 {
		return nil, fmt.Errorf("expected 5 arguments, got %d", len(args))
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
	value, ok := args[3].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for value: %T", args[3])
	}
	ttl, ok := args[4].(uint32)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for ttl: %T", args[4])
	}

	caller := contract.Caller()

	// Verify ownership
	record, found := p.xidKeeper.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, fmt.Errorf("%s.%s not found", name, tld)
	}

	callerAddr := sdk.AccAddress(caller.Bytes())
	if record.Owner != callerAddr.String() {
		return nil, fmt.Errorf("caller is not the owner of %s.%s", name, tld)
	}

	dnsRecord := types.DNSRecord{
		RecordType: uint32(recordType),
		Value:      value,
		Ttl:        ttl,
	}
	p.xidKeeper.SetDNSRecordEntry(ctx, tld, name, dnsRecord)

	p.xidKeeper.RecomputeAndStoreStateDigest(ctx)

	// Emit EVM event
	if err := p.EmitDNSRecordSet(ctx, stateDB, name, tld, recordType, value); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// DeleteDNSRecord handles the deleteDNSRecord(name, tld, recordType) function.
func (p Precompile) DeleteDNSRecord(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
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

	caller := contract.Caller()

	// Verify ownership
	record, found := p.xidKeeper.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, fmt.Errorf("%s.%s not found", name, tld)
	}

	callerAddr := sdk.AccAddress(caller.Bytes())
	if record.Owner != callerAddr.String() {
		return nil, fmt.Errorf("caller is not the owner of %s.%s", name, tld)
	}

	p.xidKeeper.DeleteDNSRecordEntry(ctx, tld, name, uint32(recordType))

	p.xidKeeper.RecomputeAndStoreStateDigest(ctx)

	// Emit EVM event
	if err := p.EmitDNSRecordDeleted(ctx, stateDB, name, tld, recordType); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// LinkIdentity handles the linkIdentity(name, tld, identityAddress, label) function.
func (p Precompile) LinkIdentity(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("expected 4 arguments, got %d", len(args))
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for name: %T", args[0])
	}
	tld, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for tld: %T", args[1])
	}
	identityAddress, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for identityAddress: %T", args[2])
	}
	label, ok := args[3].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for label: %T", args[3])
	}

	caller := contract.Caller()

	// Verify ownership
	record, found := p.xidKeeper.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, fmt.Errorf("%s.%s not found", name, tld)
	}

	callerAddr := sdk.AccAddress(caller.Bytes())
	if record.Owner != callerAddr.String() {
		return nil, fmt.Errorf("caller is not the owner of %s.%s", name, tld)
	}

	identity := types.LinkedIdentity{
		Address: identityAddress,
		Label:   label,
	}
	if err := p.xidKeeper.SetLinkedIdentityEntry(ctx, tld, name, identity); err != nil {
		return nil, err
	}

	newRoot := p.xidKeeper.RecomputeAndStoreContentRoot(ctx, tld, name)
	p.xidKeeper.RecomputeAndStoreStateDigest(ctx)

	if err := p.EmitIdentityLinked(ctx, stateDB, name, tld, identityAddress, label); err != nil {
		return nil, err
	}
	if err := p.EmitContentRootUpdated(ctx, stateDB, name, tld, newRoot); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// UnlinkIdentity handles the unlinkIdentity(name, tld, identityAddress) function.
func (p Precompile) UnlinkIdentity(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
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
	identityAddress, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for identityAddress: %T", args[2])
	}

	caller := contract.Caller()

	// Verify ownership
	record, found := p.xidKeeper.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, fmt.Errorf("%s.%s not found", name, tld)
	}

	callerAddr := sdk.AccAddress(caller.Bytes())
	if record.Owner != callerAddr.String() {
		return nil, fmt.Errorf("caller is not the owner of %s.%s", name, tld)
	}

	if err := p.xidKeeper.RevokeLinkedIdentityEntry(ctx, tld, name, identityAddress); err != nil {
		return nil, err
	}

	newRoot := p.xidKeeper.RecomputeAndStoreContentRoot(ctx, tld, name)
	p.xidKeeper.RecomputeAndStoreStateDigest(ctx)

	if err := p.EmitIdentityUnlinked(ctx, stateDB, name, tld, identityAddress); err != nil {
		return nil, err
	}
	if err := p.EmitContentRootUpdated(ctx, stateDB, name, tld, newRoot); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// AttestStateDigest handles the attestStateDigest(digest, signature) function.
func (p Precompile) AttestStateDigest(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	digest, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for digest: %T", args[0])
	}
	signature, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for signature: %T", args[1])
	}

	caller := contract.Caller()
	callerAddr := sdk.AccAddress(caller.Bytes())

	msg := &types.MsgAttestStateDigest{
		Signer:    callerAddr.String(),
		Digest:    digest,
		Signature: signature,
	}

	if err := p.xidKeeper.SubmitAttestation(ctx, msg); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// SetPrimaryName handles the setPrimaryName(name, tld) function.
func (p Precompile) SetPrimaryName(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
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

	caller := contract.Caller()
	callerAddr := sdk.AccAddress(caller.Bytes())

	// Verify ownership
	record, found := p.xidKeeper.GetNameRecord(ctx, tld, name)
	if !found {
		return nil, fmt.Errorf("%s.%s not found", name, tld)
	}
	if record.Owner != callerAddr.String() {
		return nil, fmt.Errorf("caller is not the owner of %s.%s", name, tld)
	}

	p.xidKeeper.SetPrimaryNameEntry(ctx, callerAddr, tld, name)

	// Emit EVM event
	if err := p.EmitPrimaryNameSet(ctx, stateDB, caller, name, tld); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// UpdateContentRoot is deprecated — content root is auto-computed from active linked identities.
func (p Precompile) UpdateContentRoot(
	_ sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	_ *abi.Method,
	_ []interface{},
) ([]byte, error) {
	return nil, fmt.Errorf("updateContentRoot is deprecated; content root is auto-computed from active linked identities")
}
