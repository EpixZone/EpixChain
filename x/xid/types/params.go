package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultFeeDenom             = "aepix"
	DefaultMinNameLength        = 1
	DefaultMaxNameLength        = 64
	DefaultAttestationThreshold = uint64(1)
	DefaultAttestationEnabled   = true
)

// AttestationConfig holds governance-adjustable attestation parameters.
// Stored separately from proto-generated Params to avoid modifying generated code.
type AttestationConfig struct {
	Threshold uint64 `json:"threshold"` // min validators needed to finalize
	Enabled   bool   `json:"enabled"`   // whether attestation system is active
}

// DefaultAttestationConfig returns the default attestation config
func DefaultAttestationConfig() AttestationConfig {
	return AttestationConfig{
		Threshold: DefaultAttestationThreshold,
		Enabled:   DefaultAttestationEnabled,
	}
}

// Validate performs basic validation
func (c AttestationConfig) Validate() error {
	if c.Threshold == 0 {
		return fmt.Errorf("attestation threshold must be greater than 0")
	}
	return nil
}

// DefaultParams returns the default module parameters
func DefaultParams() Params {
	return Params{
		FeeDenom:      DefaultFeeDenom,
		MinNameLength: DefaultMinNameLength,
		MaxNameLength: DefaultMaxNameLength,
	}
}

// String implements the Stringer interface
func (p Params) String() string {
	return fmt.Sprintf("Params{FeeDenom: %s, MinNameLength: %d, MaxNameLength: %d}",
		p.FeeDenom, p.MinNameLength, p.MaxNameLength)
}

// Validate performs basic validation of module parameters
func (p Params) Validate() error {
	if err := sdk.ValidateDenom(p.FeeDenom); err != nil {
		return fmt.Errorf("invalid fee denom: %w", err)
	}
	if p.MinNameLength == 0 {
		return fmt.Errorf("min name length must be greater than 0")
	}
	if p.MaxNameLength == 0 {
		return fmt.Errorf("max name length must be greater than 0")
	}
	if p.MinNameLength > p.MaxNameLength {
		return fmt.Errorf("min name length (%d) must not exceed max name length (%d)", p.MinNameLength, p.MaxNameLength)
	}
	return nil
}
