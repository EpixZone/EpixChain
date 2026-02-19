package xid

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/xid/keeper"
	"github.com/cosmos/evm/x/xid/types"
)

// InitGenesis initializes the xid module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	// Validate genesis
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	// Set params
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Set TLD configs
	for _, tld := range genState.Tlds {
		k.SetTLDConfig(ctx, tld)
	}

	// Set name entries and build owner counts
	ownerCounts := make(map[string]uint64)
	for _, entry := range genState.Names {
		k.SetNameRecord(ctx, entry.Record)

		ownerAddr, err := sdk.AccAddressFromBech32(entry.Record.Owner)
		if err != nil {
			panic(err)
		}
		k.SetOwnerIndex(ctx, ownerAddr, entry.Record.Tld, entry.Record.Name)
		ownerCounts[entry.Record.Owner]++

		if entry.Profile != nil {
			k.SetProfileRecord(ctx, entry.Record.Tld, entry.Record.Name, *entry.Profile)
		}

		for _, dns := range entry.DnsRecords {
			k.SetDNSRecordEntry(ctx, entry.Record.Tld, entry.Record.Name, dns)
		}
	}

	// Persist owner counts
	for ownerBech32, count := range ownerCounts {
		ownerAddr, _ := sdk.AccAddressFromBech32(ownerBech32)
		k.SetOwnerCount(ctx, ownerAddr, count)
	}
}

// ExportGenesis returns the xid module's exported genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	tlds := k.GetAllTLDs(ctx)

	var names []types.NameEntry
	k.IterateNameRecords(ctx, func(record types.NameRecord) bool {
		entry := types.NameEntry{
			Record:     record,
			DnsRecords: k.GetAllDNSRecords(ctx, record.Tld, record.Name),
		}

		profile, found := k.GetProfileRecord(ctx, record.Tld, record.Name)
		if found {
			entry.Profile = &profile
		}

		names = append(names, entry)
		return false
	})

	return &types.GenesisState{
		Params: params,
		Tlds:   tlds,
		Names:  names,
	}
}
