package types

import (
	"bytes"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Msg type URLs
const (
	TypeMsgRegisterName      = "register_name"
	TypeMsgTransferName      = "transfer_name"
	TypeMsgUpdateProfile     = "update_profile"
	TypeMsgSetDNSRecord      = "set_dns_record"
	TypeMsgDeleteDNSRecord   = "delete_dns_record"
	TypeMsgCreateTLD         = "create_tld"
	TypeMsgUpdateTLDConfig   = "update_tld_config"
	TypeMsgUpdateParams      = "update_params"
	TypeMsgAttestStateDigest = "attest_state_digest"
	TypeMsgSetPrimaryName    = "set_primary_name"
)

// GetSigners returns the expected signers for MsgRegisterName.
func (msg *MsgRegisterName) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgRegisterName.
func (msg *MsgRegisterName) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	return nil
}

// GetSigners returns the expected signers for MsgTransferName.
func (msg *MsgTransferName) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgTransferName.
func (msg *MsgTransferName) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	newOwnerAddr, err := sdk.AccAddressFromBech32(msg.NewOwner)
	if err != nil {
		return err
	}
	if msg.Owner == msg.NewOwner {
		return ErrInvalidName.Wrap("cannot transfer name to yourself")
	}
	if bytes.Equal(newOwnerAddr, make([]byte, len(newOwnerAddr))) {
		return ErrInvalidName.Wrap("cannot transfer name to the zero address")
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	return nil
}

// GetSigners returns the expected signers for MsgUpdateProfile.
func (msg *MsgUpdateProfile) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgUpdateProfile.
func (msg *MsgUpdateProfile) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	if err := ValidateProfile(msg.Profile); err != nil {
		return err
	}
	return nil
}

// GetSigners returns the expected signers for MsgSetDNSRecord.
func (msg *MsgSetDNSRecord) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgSetDNSRecord.
func (msg *MsgSetDNSRecord) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	if err := ValidateDNSRecordType(msg.Record.RecordType); err != nil {
		return err
	}
	if msg.Record.Value == "" {
		return ErrInvalidDNSRecord.Wrap("record value cannot be empty")
	}
	if len(msg.Record.Value) > MaxDNSValueLength {
		return ErrInvalidDNSRecord.Wrapf("record value exceeds maximum length of %d characters", MaxDNSValueLength)
	}
	if err := ValidateDNSTTL(msg.Record.Ttl); err != nil {
		return err
	}
	return nil
}

// GetSigners returns the expected signers for MsgDeleteDNSRecord.
func (msg *MsgDeleteDNSRecord) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgDeleteDNSRecord.
func (msg *MsgDeleteDNSRecord) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	if msg.RecordType == 0 {
		return ErrInvalidDNSRecord.Wrap("record type cannot be 0")
	}
	return nil
}

// GetSigners returns the expected signers for MsgCreateTLD.
func (msg *MsgCreateTLD) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgCreateTLD.
func (msg *MsgCreateTLD) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if err := ValidateTLD(msg.TldConfig.Tld); err != nil {
		return err
	}
	if err := ValidatePriceTiers(msg.TldConfig.PriceTiers); err != nil {
		return err
	}
	return nil
}

// GetSigners returns the expected signers for MsgUpdateTLDConfig.
func (msg *MsgUpdateTLDConfig) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgUpdateTLDConfig.
func (msg *MsgUpdateTLDConfig) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	if err := ValidateTLD(msg.TldConfig.Tld); err != nil {
		return err
	}
	if err := ValidatePriceTiers(msg.TldConfig.PriceTiers); err != nil {
		return err
	}
	return nil
}

// GetSigners returns the expected signers for MsgUpdateParams.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgUpdateParams.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	return msg.Params.Validate()
}

// GetSigners returns the expected signers for MsgSetPrimaryName.
func (msg *MsgSetPrimaryName) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgSetPrimaryName.
func (msg *MsgSetPrimaryName) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	return nil
}

// GetSigners returns the expected signers for MsgAttestStateDigest.
func (msg *MsgAttestStateDigest) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgAttestStateDigest.
func (msg *MsgAttestStateDigest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return err
	}
	if len(msg.Digest) != 64 {
		return ErrInvalidAttestation.Wrap("digest must be a 64-character hex SHA-256 hash")
	}
	if _, err := hex.DecodeString(msg.Digest); err != nil {
		return ErrInvalidAttestation.Wrap("digest must contain only hexadecimal characters (0-9, a-f)")
	}
	if msg.Signature == "" {
		return ErrInvalidAttestation.Wrap("signature cannot be empty")
	}
	return nil
}


// GetSigners returns the expected signers for MsgLinkIdentity.
func (msg *MsgLinkIdentity) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgLinkIdentity.
func (msg *MsgLinkIdentity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	if msg.Identity.Address == "" {
		return ErrInvalidIdentity.Wrap("identity address cannot be empty")
	}
	return nil
}

// GetSigners returns the expected signers for MsgUnlinkIdentity.
func (msg *MsgUnlinkIdentity) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{signer}
}

// ValidateBasic performs stateless validation for MsgUnlinkIdentity.
func (msg *MsgUnlinkIdentity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if err := ValidateName(msg.Name); err != nil {
		return err
	}
	if err := ValidateTLD(msg.Tld); err != nil {
		return err
	}
	if msg.Address == "" {
		return ErrInvalidIdentity.Wrap("identity address cannot be empty")
	}
	return nil
}
