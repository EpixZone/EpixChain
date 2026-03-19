package types

import "fmt"

// DefaultGenesisState returns the default VRF genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:  DefaultParams(),
		Beacons: []RandomBeacon{},
	}
}

// Validate performs basic validation of the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	heightMap := make(map[uint64]bool)
	for _, beacon := range gs.Beacons {
		if heightMap[beacon.Height] {
			return fmt.Errorf("duplicate beacon at height %d", beacon.Height)
		}
		heightMap[beacon.Height] = true

		if len(beacon.Beacon) != 64 {
			return fmt.Errorf("beacon at height %d has invalid length %d (expected 64 hex chars)", beacon.Height, len(beacon.Beacon))
		}
	}

	return nil
}
