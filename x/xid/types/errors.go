package types

import errorsmod "cosmossdk.io/errors"

// xid module sentinel errors
var (
	ErrNameTaken             = errorsmod.Register(ModuleName, 2, "name is already registered")
	ErrNameNotFound          = errorsmod.Register(ModuleName, 3, "name not found")
	ErrTLDNotFound           = errorsmod.Register(ModuleName, 4, "TLD not found")
	ErrTLDDisabled           = errorsmod.Register(ModuleName, 5, "TLD is disabled for new registrations")
	ErrTLDAlreadyExists      = errorsmod.Register(ModuleName, 6, "TLD already exists")
	ErrNotOwner              = errorsmod.Register(ModuleName, 7, "sender is not the name owner")
	ErrInvalidName           = errorsmod.Register(ModuleName, 8, "invalid name")
	ErrInvalidTLD            = errorsmod.Register(ModuleName, 9, "invalid TLD")
	ErrInsufficientFee       = errorsmod.Register(ModuleName, 10, "insufficient registration fee")
	ErrInvalidDNSRecord      = errorsmod.Register(ModuleName, 11, "invalid DNS record")
	ErrDNSRecordNotFound     = errorsmod.Register(ModuleName, 12, "DNS record not found")
	ErrInvalidProfile        = errorsmod.Register(ModuleName, 13, "invalid profile")
	ErrInvalidParams         = errorsmod.Register(ModuleName, 14, "invalid module parameters")
	ErrInvalidPriceTier      = errorsmod.Register(ModuleName, 15, "invalid price tier configuration")
	ErrIdentityNotFound      = errorsmod.Register(ModuleName, 16, "linked identity not found")
	ErrIdentityAlreadyLinked = errorsmod.Register(ModuleName, 17, "identity address already linked")
	ErrInvalidContentRoot    = errorsmod.Register(ModuleName, 18, "invalid content root")
	ErrNotValidator          = errorsmod.Register(ModuleName, 19, "signer is not an active bonded validator")
	ErrDigestMismatch        = errorsmod.Register(ModuleName, 20, "digest does not match current state digest")
	ErrInvalidAttestation    = errorsmod.Register(ModuleName, 21, "invalid attestation")
	ErrAttestationDisabled   = errorsmod.Register(ModuleName, 22, "attestation system is not enabled")
	ErrAlreadyAttested       = errorsmod.Register(ModuleName, 23, "validator has already attested to this digest")
	ErrInvalidIdentity       = errorsmod.Register(ModuleName, 24, "invalid linked identity")
	ErrOwnerCountUnderflow   = errorsmod.Register(ModuleName, 25, "owner count is already zero")
	ErrInvalidAttestKey      = errorsmod.Register(ModuleName, 26, "invalid attestation key registration")
)
