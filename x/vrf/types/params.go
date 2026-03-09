package types

import "fmt"

const (
	DefaultLookbackBlocks uint64 = 256
)

// DefaultParams returns the default VRF module params
func DefaultParams() Params {
	return Params{
		Enabled:        true,
		LookbackBlocks: DefaultLookbackBlocks,
	}
}

// Validate performs basic validation of VRF params
func (p Params) Validate() error {
	if p.LookbackBlocks == 0 {
		return fmt.Errorf("lookback_blocks must be greater than 0")
	}
	return nil
}
