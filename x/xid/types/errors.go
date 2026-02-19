package types

import errorsmod "cosmossdk.io/errors"

// xid module sentinel errors
var (
	ErrNameTaken         = errorsmod.Register(ModuleName, 2, "name is already registered")
	ErrNameNotFound      = errorsmod.Register(ModuleName, 3, "name not found")
	ErrTLDNotFound       = errorsmod.Register(ModuleName, 4, "TLD not found")
	ErrTLDDisabled       = errorsmod.Register(ModuleName, 5, "TLD is disabled for new registrations")
	ErrTLDAlreadyExists  = errorsmod.Register(ModuleName, 6, "TLD already exists")
	ErrNotOwner          = errorsmod.Register(ModuleName, 7, "sender is not the name owner")
	ErrInvalidName       = errorsmod.Register(ModuleName, 8, "invalid name")
	ErrInvalidTLD        = errorsmod.Register(ModuleName, 9, "invalid TLD")
	ErrInsufficientFee   = errorsmod.Register(ModuleName, 10, "insufficient registration fee")
	ErrInvalidDNSRecord  = errorsmod.Register(ModuleName, 11, "invalid DNS record")
	ErrDNSRecordNotFound = errorsmod.Register(ModuleName, 12, "DNS record not found")
	ErrInvalidProfile    = errorsmod.Register(ModuleName, 13, "invalid profile")
	ErrInvalidParams     = errorsmod.Register(ModuleName, 14, "invalid module parameters")
	ErrInvalidPriceTier    = errorsmod.Register(ModuleName, 15, "invalid price tier configuration")
	ErrEpixNetPeerNotFound      = errorsmod.Register(ModuleName, 16, "EpixNet peer not found")
	ErrEpixNetPeerAlreadyLinked = errorsmod.Register(ModuleName, 17, "EpixNet peer address is already linked to another xID")
	ErrInvalidContentRoot       = errorsmod.Register(ModuleName, 18, "invalid content root")
)
