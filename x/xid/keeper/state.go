package keeper

import (
	"encoding/json"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/evm/x/xid/types"
)

// ---------------------------------------------------------------------------
// Name Records
// ---------------------------------------------------------------------------

// SetNameRecord stores a name record
func (k Keeper) SetNameRecord(ctx sdk.Context, record types.NameRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(record)
	store.Set(types.NameRecordKey(record.Tld, record.Name), bz)
}

// GetNameRecord retrieves a name record
func (k Keeper) GetNameRecord(ctx sdk.Context, tld, name string) (types.NameRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NameRecordKey(tld, name))
	if bz == nil {
		return types.NameRecord{}, false
	}

	var record types.NameRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.NameRecord{}, false
	}
	return record, true
}

// HasNameRecord checks if a name is already registered
func (k Keeper) HasNameRecord(ctx sdk.Context, tld, name string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.NameRecordKey(tld, name))
}


// IterateNameRecords iterates over all name records
func (k Keeper) IterateNameRecords(ctx sdk.Context, cb func(record types.NameRecord) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixNameRecord)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.NameRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		if cb(record) {
			break
		}
	}
}

// ---------------------------------------------------------------------------
// Owner Index
// ---------------------------------------------------------------------------

// SetOwnerIndex adds an entry to the owner index
func (k Keeper) SetOwnerIndex(ctx sdk.Context, owner sdk.AccAddress, tld, name string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OwnerIndexKey(owner.Bytes(), tld, name), []byte{1})
}

// DeleteOwnerIndex removes an entry from the owner index
func (k Keeper) DeleteOwnerIndex(ctx sdk.Context, owner sdk.AccAddress, tld, name string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.OwnerIndexKey(owner.Bytes(), tld, name))
}

// GetNamesByOwnerAddr returns all name records owned by an address
func (k Keeper) GetNamesByOwnerAddr(ctx sdk.Context, owner sdk.AccAddress) []types.NameRecord {
	store := ctx.KVStore(k.storeKey)
	prefix := types.OwnerIndexPrefix(owner.Bytes())
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var records []types.NameRecord
	for ; iterator.Valid(); iterator.Next() {
		// KVStorePrefixIterator returns full keys (not stripped).
		// Full key: [prefixByte][owner 20 bytes][len(tld)][tld][name]
		// Skip past the prefix to get [len(tld)][tld][name]
		key := iterator.Key()
		remaining := key[len(prefix):]
		if len(remaining) < 1 {
			continue
		}
		tldLen := int(remaining[0])
		if len(remaining) < 1+tldLen {
			continue
		}
		tld := string(remaining[1 : 1+tldLen])
		name := string(remaining[1+tldLen:])

		record, found := k.GetNameRecord(ctx, tld, name)
		if found {
			records = append(records, record)
		}
	}

	return records
}

// GetNamesByOwnerPaginated returns a paginated list of name records owned by an address.
func (k Keeper) GetNamesByOwnerPaginated(ctx sdk.Context, owner sdk.AccAddress, pageReq *query.PageRequest) ([]types.NameRecord, *query.PageResponse, error) {
	ownerPrefix := types.OwnerIndexPrefix(owner.Bytes())
	store := prefix.NewStore(ctx.KVStore(k.storeKey), ownerPrefix)

	var records []types.NameRecord
	pageRes, err := query.Paginate(store, pageReq, func(key, _ []byte) error {
		// key is [len(tld)][tld][name] (prefix already stripped)
		if len(key) < 1 {
			return nil
		}
		tldLen := int(key[0])
		if len(key) < 1+tldLen {
			return nil
		}
		tld := string(key[1 : 1+tldLen])
		name := string(key[1+tldLen:])

		record, found := k.GetNameRecord(ctx, tld, name)
		if found {
			records = append(records, record)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return records, pageRes, nil
}

// ---------------------------------------------------------------------------
// Profiles
// ---------------------------------------------------------------------------

// SetProfileRecord stores a profile for a name
func (k Keeper) SetProfileRecord(ctx sdk.Context, tld, name string, profile types.Profile) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(profile)
	store.Set(types.ProfileKey(tld, name), bz)
}

// GetProfileRecord retrieves the profile for a name
func (k Keeper) GetProfileRecord(ctx sdk.Context, tld, name string) (types.Profile, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProfileKey(tld, name))
	if bz == nil {
		return types.Profile{}, false
	}

	var profile types.Profile
	if err := json.Unmarshal(bz, &profile); err != nil {
		return types.Profile{}, false
	}
	return profile, true
}

// DeleteProfile removes a profile from the store
func (k Keeper) DeleteProfile(ctx sdk.Context, tld, name string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.ProfileKey(tld, name))
}

// ---------------------------------------------------------------------------
// DNS Records
// ---------------------------------------------------------------------------

// SetDNSRecordEntry stores a DNS record for a name
func (k Keeper) SetDNSRecordEntry(ctx sdk.Context, tld, name string, record types.DNSRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(record)
	store.Set(types.DNSRecordKey(tld, name, record.RecordType), bz)
}

// GetDNSRecordEntry retrieves a specific DNS record
func (k Keeper) GetDNSRecordEntry(ctx sdk.Context, tld, name string, recordType uint32) (types.DNSRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DNSRecordKey(tld, name, recordType))
	if bz == nil {
		return types.DNSRecord{}, false
	}

	var record types.DNSRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.DNSRecord{}, false
	}
	return record, true
}

// DeleteDNSRecordEntry removes a DNS record from the store
func (k Keeper) DeleteDNSRecordEntry(ctx sdk.Context, tld, name string, recordType uint32) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.DNSRecordKey(tld, name, recordType))
}

// GetAllDNSRecords returns all DNS records for a name
func (k Keeper) GetAllDNSRecords(ctx sdk.Context, tld, name string) []types.DNSRecord {
	store := ctx.KVStore(k.storeKey)
	prefix := types.DNSRecordPrefix(tld, name)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var records []types.DNSRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.DNSRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records
}

// ---------------------------------------------------------------------------
// TLD Config
// ---------------------------------------------------------------------------

// SetTLDConfig stores a TLD configuration
func (k Keeper) SetTLDConfig(ctx sdk.Context, config types.TLDConfig) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(config)
	store.Set(types.TLDConfigKey(config.Tld), bz)
}

// GetTLDConfig retrieves a TLD configuration
func (k Keeper) GetTLDConfig(ctx sdk.Context, tld string) (types.TLDConfig, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TLDConfigKey(tld))
	if bz == nil {
		return types.TLDConfig{}, false
	}

	var config types.TLDConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return types.TLDConfig{}, false
	}
	return config, true
}

// HasTLDConfig checks if a TLD is registered
func (k Keeper) HasTLDConfig(ctx sdk.Context, tld string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.TLDConfigKey(tld))
}

// GetAllTLDs returns all registered TLD configurations
func (k Keeper) GetAllTLDs(ctx sdk.Context) []types.TLDConfig {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixTLDConfig)
	defer iterator.Close()

	var configs []types.TLDConfig
	for ; iterator.Valid(); iterator.Next() {
		var config types.TLDConfig
		if err := json.Unmarshal(iterator.Value(), &config); err != nil {
			continue
		}
		configs = append(configs, config)
	}
	return configs
}
