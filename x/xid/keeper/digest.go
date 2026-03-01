package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/xid/types"
)

// domainDigestEntry is the canonical representation of a single domain for digest computation.
type domainDigestEntry struct {
	Name        string              `json:"name"`
	Tld         string              `json:"tld"`
	Owner       string              `json:"owner"`
	Profile     *types.Profile      `json:"profile,omitempty"`
	DNS         []types.DNSRecord   `json:"dns,omitempty"`
	Peers       []types.EpixNetPeer `json:"peers,omitempty"`
	ContentRoot string              `json:"content_root,omitempty"`
}

// ComputeStateDigest computes a deterministic SHA-256 digest of all xID state.
// It iterates all names, collects their full data, sorts by "tld/name", and hashes.
func (k Keeper) ComputeStateDigest(ctx sdk.Context) (string, uint64) {
	var entries []domainDigestEntry
	var numNames uint64

	k.IterateNameRecords(ctx, func(record types.NameRecord) bool {
		numNames++

		entry := domainDigestEntry{
			Name:  record.Name,
			Tld:   record.Tld,
			Owner: record.Owner,
		}

		if profile, found := k.GetProfileRecord(ctx, record.Tld, record.Name); found {
			entry.Profile = &profile
		}

		dns := k.GetAllDNSRecords(ctx, record.Tld, record.Name)
		if len(dns) > 0 {
			entry.DNS = dns
		}

		peers := k.GetAllEpixNetPeers(ctx, record.Tld, record.Name)
		if len(peers) > 0 {
			entry.Peers = peers
		}

		if cr, found := k.GetContentRoot(ctx, record.Tld, record.Name); found && cr.Root != "" {
			entry.ContentRoot = cr.Root
		}

		entries = append(entries, entry)
		return false
	})

	// Sort by tld/name for determinism
	sort.Slice(entries, func(i, j int) bool {
		ki := entries[i].Tld + "/" + entries[i].Name
		kj := entries[j].Tld + "/" + entries[j].Name
		return ki < kj
	})

	// Hash all entries
	h := sha256.New()
	for _, entry := range entries {
		bz, _ := json.Marshal(entry)
		h.Write(bz)
	}

	return hex.EncodeToString(h.Sum(nil)), numNames
}

// RecomputeAndStoreStateDigest rebuilds the Merkle tree from scratch and stores
// the root as the state digest. This is used for genesis initialization and
// migration. For incremental updates use UpdateDomainInTree instead.
func (k Keeper) RecomputeAndStoreStateDigest(ctx sdk.Context) string {
	return k.RebuildTree(ctx)
}

// GetStateDigest retrieves the current state digest.
func (k Keeper) GetStateDigest(ctx sdk.Context) (types.StateDigest, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.StateDigestKey())
	if bz == nil {
		return types.StateDigest{}, false
	}
	var sd types.StateDigest
	if err := json.Unmarshal(bz, &sd); err != nil {
		return types.StateDigest{}, false
	}
	return sd, true
}

// SetStateDigest stores the state digest.
func (k Keeper) SetStateDigest(ctx sdk.Context, sd types.StateDigest) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(sd)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal state digest: %v", err))
	}
	store.Set(types.StateDigestKey(), bz)
}
