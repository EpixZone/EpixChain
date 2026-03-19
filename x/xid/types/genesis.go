package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultGenesisState returns the default genesis state with the .epix TLD
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Tlds: []TLDConfig{
			DefaultEpixTLD(),
		},
		Names: []NameEntry{},
	}
}

// DefaultEpixTLD returns the default .epix TLD configuration
func DefaultEpixTLD() TLDConfig {
	epix18 := math.NewIntWithDecimal(1, 18) // 10^18 (1 EPIX in aepix)

	return TLDConfig{
		Tld:     "epix",
		Enabled: true,
		PriceTiers: []PriceTier{
			{MaxLength: 1, Price: math.NewInt(100_000_000).Mul(epix18)},     // 100M EPIX
			{MaxLength: 2, Price: math.NewInt(1_000_000).Mul(epix18)},       // 1M EPIX
			{MaxLength: 3, Price: math.NewInt(500_000).Mul(epix18)},         // 500K EPIX
			{MaxLength: 4, Price: math.NewInt(100_000).Mul(epix18)},         // 100K EPIX
			{MaxLength: 4294967295, Price: math.NewInt(10_000).Mul(epix18)}, // 10K EPIX (5+ chars)
		},
	}
}

// Validate performs basic validation of the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	tldMap := make(map[string]bool)
	for _, tld := range gs.Tlds {
		if err := ValidateTLD(tld.Tld); err != nil {
			return fmt.Errorf("invalid TLD %q: %w", tld.Tld, err)
		}
		if tldMap[tld.Tld] {
			return fmt.Errorf("duplicate TLD: %s", tld.Tld)
		}
		tldMap[tld.Tld] = true

		if len(tld.PriceTiers) == 0 {
			return fmt.Errorf("TLD %q must have at least one price tier", tld.Tld)
		}
		for _, tier := range tld.PriceTiers {
			if tier.Price.IsNegative() {
				return fmt.Errorf("TLD %q has negative price tier", tld.Tld)
			}
		}
	}

	nameMap := make(map[string]bool)
	for _, entry := range gs.Names {
		fullName := entry.Record.Name + "." + entry.Record.Tld
		if nameMap[fullName] {
			return fmt.Errorf("duplicate name: %s", fullName)
		}
		nameMap[fullName] = true

		if !tldMap[entry.Record.Tld] {
			return fmt.Errorf("name %q references non-existent TLD %q", fullName, entry.Record.Tld)
		}
	}

	return nil
}
