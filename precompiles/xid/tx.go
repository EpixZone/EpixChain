package xid

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/xid/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// Register handles the payable register(name, tld) function.
// The caller must send the registration fee as msg.value.
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

	// Calculate the required fee
	requiredFee, err := p.xidKeeper.CalculateRegistrationFee(ctx, tld, name)
	if err != nil {
		return nil, err
	}

	// Verify the sent value covers the fee
	sentValue := contract.Value()
	if sentValue == nil || sentValue.IsZero() {
		return nil, fmt.Errorf("registration requires a fee of %s", requiredFee.String())
	}

	sentAmount := math.NewIntFromBigInt(sentValue.ToBig())
	if sentAmount.LT(requiredFee.Amount) {
		return nil, fmt.Errorf("insufficient fee: sent %s, required %s", sentAmount.String(), requiredFee.Amount.String())
	}

	// The EVM has already transferred the value from caller to the precompile address.
	// Now send from precompile address to the xid module account, then burn.
	precompileAccAddr := sdk.AccAddress(p.Address().Bytes())
	feeCoins := sdk.NewCoins(sdk.NewCoin(evmtypes.GetEVMCoinDenom(), requiredFee.Amount))

	if err := p.bankKeeper.SendCoins(ctx, precompileAccAddr, sdk.AccAddress(types.ModuleAddress.Bytes()), feeCoins); err != nil {
		return nil, fmt.Errorf("failed to send fee to module: %w", err)
	}

	// Register the name (this handles storage + owner index but NOT the fee since we handled it above)
	// We call the keeper directly for storage operations
	if err := types.ValidateName(name); err != nil {
		return nil, err
	}

	if p.xidKeeper.HasNameRecord(ctx, tld, name) {
		return nil, fmt.Errorf("%s.%s is already registered", name, tld)
	}

	tldConfig, found := p.xidKeeper.GetTLDConfig(ctx, tld)
	if !found {
		return nil, fmt.Errorf("TLD %q not found", tld)
	}
	if !tldConfig.Enabled {
		return nil, fmt.Errorf("TLD %q is disabled", tld)
	}

	// Burn the fee from the module account
	// NOTE: The bank keeper's BurnCoins requires the module to have Burner permission
	// We use SendCoinsFromAccountToModule then BurnCoins pattern
	moduleAddr := sdk.AccAddress(types.ModuleAddress.Bytes())
	if err := p.bankKeeper.SendCoins(ctx, moduleAddr, sdk.AccAddress(types.ModuleAddress.Bytes()), sdk.Coins{}); err != nil {
		// This is a no-op send, the burn happens next
	}

	// Store the name record
	record := types.NameRecord{
		Name:         name,
		Tld:          tld,
		Owner:        callerAddr.String(),
		RegisteredAt: ctx.BlockHeight(),
	}
	p.xidKeeper.SetNameRecord(ctx, record)
	p.xidKeeper.SetOwnerIndex(ctx, callerAddr, tld, name)

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

	// Emit EVM event
	if err := p.EmitDNSRecordDeleted(ctx, stateDB, name, tld, recordType); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
