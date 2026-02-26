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

// RecomputeAndStoreStateDigest recomputes the state digest and stores it.
// Returns the digest string. Emits an event.
// Clears any attestations for the previous digest since they are now stale.
func (k Keeper) RecomputeAndStoreStateDigest(ctx sdk.Context) string {
	// Clear attestations for the old digest — they are stale after a state change
	if oldDigest, found := k.GetStateDigest(ctx); found {
		k.ClearAttestationsForDigest(ctx, oldDigest.Digest)
	}

	digest, numNames := k.ComputeStateDigest(ctx)

	sd := types.StateDigest{
		Digest:   digest,
		Height:   uint64(ctx.BlockHeight()),
		NumNames: numNames,
	}
	k.SetStateDigest(ctx, sd)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"xid_state_digest_updated",
		sdk.NewAttribute("digest", digest),
		sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
		sdk.NewAttribute("num_names", fmt.Sprintf("%d", numNames)),
	))

	return digest
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
