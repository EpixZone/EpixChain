package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultFeeDenom      = "aepix"
	DefaultMinNameLength = 1
	DefaultMaxNameLength = 64
)

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
