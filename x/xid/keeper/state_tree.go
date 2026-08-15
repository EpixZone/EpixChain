package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/xid/types"
)

// emptyLeafHash is the sentinel hash for unoccupied Merkle tree leaves.
var emptyLeafHash [32]byte

func init() {
	emptyLeafHash = sha256.Sum256(nil)
}

// MerkleMetadata tracks the state of the global Merkle tree.
type MerkleMetadata struct {
	NumLeaves uint32 `json:"num_leaves"` // Number of occupied leaves
	Capacity  uint32 `json:"capacity"`   // Tree capacity (power of 2, minimum 2)
}

// MerkleProof is an inclusion proof for a single domain in the global Merkle tree.
type MerkleProof struct {
	LeafIndex uint32   `json:"leaf_index"`
	LeafHash  string   `json:"leaf_hash"`
	Siblings  []string `json:"siblings"`
	Root      string   `json:"root"`
}

// treeHeight returns the number of levels above the leaves for the given capacity.
// A capacity of 2 has height 1, capacity 4 has height 2, etc.
func treeHeight(capacity uint32) uint16 {
	var h uint16
	c := capacity
	for c > 1 {
		c >>= 1
		h++
	}
	return h
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func (k Keeper) GetMerkleMetadata(ctx sdk.Context) MerkleMetadata {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MerkleMetadataKey())
	if bz == nil {
		return MerkleMetadata{}
	}
	var meta MerkleMetadata
	if err := json.Unmarshal(bz, &meta); err != nil {
		return MerkleMetadata{}
	}
	return meta
}

func (k Keeper) SetMerkleMetadata(ctx sdk.Context, meta MerkleMetadata) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(meta)
	store.Set(types.MerkleMetadataKey(), bz)
}

// ---------------------------------------------------------------------------
// Node storage
// ---------------------------------------------------------------------------

func (k Keeper) GetMerkleNode(ctx sdk.Context, level uint16, index uint32) [32]byte {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MerkleNodeKey(level, index))
	if bz == nil {
		return emptyLeafHash
	}
	var h [32]byte
	copy(h[:], bz)
	return h
}

func (k Keeper) SetMerkleNode(ctx sdk.Context, level uint16, index uint32, hash [32]byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.MerkleNodeKey(level, index), hash[:])
}

// ---------------------------------------------------------------------------
// Leaf index mapping: tld/name → leaf position
// ---------------------------------------------------------------------------

func (k Keeper) GetLeafIndex(ctx sdk.Context, tld, name string) (uint32, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MerkleLeafIndexKey(tld, name))
	if bz == nil || len(bz) < 4 {
		return 0, false
	}
	return binary.BigEndian.Uint32(bz), true
}

func (k Keeper) SetLeafIndex(ctx sdk.Context, tld, name string, index uint32) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, index)
	store.Set(types.MerkleLeafIndexKey(tld, name), bz)
}

// ---------------------------------------------------------------------------
// Leaf hash computation
// ---------------------------------------------------------------------------

// computeLeafHash builds the canonical domainDigestEntry for a domain and
// returns its SHA-256 hash. This reuses the same struct as the old flat digest
// so that the data model is consistent.
func (k Keeper) computeLeafHash(ctx sdk.Context, tld, name string) [32]byte {
	pre := k.computeLeafPreimage(ctx, tld, name)
	if pre == nil {
		return emptyLeafHash
	}
	return sha256.Sum256(pre)
}

// computeLeafPreimage returns the exact canonical bytes hashed into a domain's
// leaf — `json.Marshal(domainDigestEntry)` — or nil if the name does not exist.
// This is served to clients as `leaf_preimage` so they can hash it to bind the
// returned data to the proven leaf (see docs/xid-lightclient-finality.md). Any
// change here MUST be mirrored by the client's leaf parser + a frozen KAT.
func (k Keeper) computeLeafPreimage(ctx sdk.Context, tld, name string) []byte {
	record, found := k.GetNameRecord(ctx, tld, name)
	if !found {
		return nil
	}

	entry := domainDigestEntry{
		Name:  record.Name,
		Tld:   record.Tld,
		Owner: record.Owner,
	}

	if profile, found := k.GetProfileRecord(ctx, tld, name); found {
		entry.Profile = &profile
	}

	dns := k.GetAllDNSRecords(ctx, tld, name)
	if len(dns) > 0 {
		entry.DNS = dns
	}

	identities := k.GetAllLinkedIdentities(ctx, tld, name)
	if len(identities) > 0 {
		entry.Identities = identities
	}

	if cr, found := k.GetContentRoot(ctx, tld, name); found && cr.Root != "" {
		entry.ContentRoot = cr.Root
	}

	bz, _ := json.Marshal(entry)
	return bz
}

// hashPair hashes two 32-byte children into a parent node.
func hashPair(left, right [32]byte) [32]byte {
	var combined [64]byte
	copy(combined[:32], left[:])
	copy(combined[32:], right[:])
	return sha256.Sum256(combined[:])
}

// ---------------------------------------------------------------------------
// Tree operations
// ---------------------------------------------------------------------------

// ensureCapacity doubles the tree capacity if needed. When expanding, the new
// right subtree is filled with empty hashes and internal nodes are recomputed
// along the spine.
func (k Keeper) ensureCapacity(ctx sdk.Context, meta *MerkleMetadata) {
	if meta.Capacity == 0 {
		meta.Capacity = 2
		k.SetMerkleMetadata(ctx, *meta)
		return
	}
	if meta.NumLeaves < meta.Capacity {
		return
	}

	oldCap := meta.Capacity
	newCap := oldCap * 2
	oldHeight := treeHeight(oldCap)
	newHeight := treeHeight(newCap)

	// The old root becomes the left child of the new root.
	// The right child is the Merkle root of an all-empty subtree of the old size.
	// We only need to set the new top-level node; all existing nodes remain valid.
	// The empty subtree root is computed iteratively.
	emptySubRoot := emptyLeafHash
	for level := uint16(0); level < oldHeight; level++ {
		emptySubRoot = hashPair(emptySubRoot, emptySubRoot)
	}
	// Store the empty subtree root as the right child at oldHeight level, index 1
	k.SetMerkleNode(ctx, oldHeight, 1, emptySubRoot)

	// Compute new root
	oldRoot := k.GetMerkleNode(ctx, oldHeight, 0)
	newRoot := hashPair(oldRoot, emptySubRoot)
	k.SetMerkleNode(ctx, newHeight, 0, newRoot)

	meta.Capacity = newCap
	k.SetMerkleMetadata(ctx, *meta)
}

// UpdateLeaf recomputes the leaf hash for a domain and propagates changes up
// the Merkle tree. Returns the new root hash.
func (k Keeper) UpdateLeaf(ctx sdk.Context, tld, name string) string {
	meta := k.GetMerkleMetadata(ctx)
	if meta.Capacity == 0 {
		// Tree not initialized yet — do a full rebuild
		return k.RebuildTree(ctx)
	}

	leafIdx, found := k.GetLeafIndex(ctx, tld, name)
	if !found {
		// New leaf — assign next index
		leafIdx = meta.NumLeaves
		meta.NumLeaves++
		k.ensureCapacity(ctx, &meta)
		k.SetLeafIndex(ctx, tld, name, leafIdx)
		k.SetMerkleMetadata(ctx, meta)
	}

	// Compute and store new leaf hash
	leafHash := k.computeLeafHash(ctx, tld, name)
	k.SetMerkleNode(ctx, 0, leafIdx, leafHash)

	// Walk up the tree, recomputing parents
	height := treeHeight(meta.Capacity)
	idx := leafIdx
	for level := uint16(0); level < height; level++ {
		var left, right [32]byte
		if idx%2 == 0 {
			left = k.GetMerkleNode(ctx, level, idx)
			right = k.GetMerkleNode(ctx, level, idx+1)
		} else {
			left = k.GetMerkleNode(ctx, level, idx-1)
			right = k.GetMerkleNode(ctx, level, idx)
		}
		parentHash := hashPair(left, right)
		parentIdx := idx / 2
		k.SetMerkleNode(ctx, level+1, parentIdx, parentHash)
		idx = parentIdx
	}

	root := k.GetMerkleNode(ctx, height, 0)
	return hex.EncodeToString(root[:])
}

// GenerateProof produces a Merkle inclusion proof for a domain.
func (k Keeper) GenerateProof(ctx sdk.Context, tld, name string) (MerkleProof, error) {
	meta := k.GetMerkleMetadata(ctx)
	if meta.Capacity == 0 {
		return MerkleProof{}, fmt.Errorf("merkle tree not initialized")
	}

	leafIdx, found := k.GetLeafIndex(ctx, tld, name)
	if !found {
		return MerkleProof{}, fmt.Errorf("domain %s.%s not found in merkle tree", name, tld)
	}

	height := treeHeight(meta.Capacity)
	leafHash := k.GetMerkleNode(ctx, 0, leafIdx)

	siblings := make([]string, 0, height)
	idx := leafIdx
	for level := uint16(0); level < height; level++ {
		var siblingIdx uint32
		if idx%2 == 0 {
			siblingIdx = idx + 1
		} else {
			siblingIdx = idx - 1
		}
		sibling := k.GetMerkleNode(ctx, level, siblingIdx)
		siblings = append(siblings, hex.EncodeToString(sibling[:]))
		idx /= 2
	}

	root := k.GetMerkleNode(ctx, height, 0)

	return MerkleProof{
		LeafIndex: leafIdx,
		LeafHash:  hex.EncodeToString(leafHash[:]),
		Siblings:  siblings,
		Root:      hex.EncodeToString(root[:]),
	}, nil
}

// RebuildTree performs a full O(n) rebuild of the Merkle tree from all name records.
// Used for genesis initialization and migration. Returns the root hash hex string.
func (k Keeper) RebuildTree(ctx sdk.Context) string {
	// Collect all name records and sort by tld/name
	type nameKey struct {
		Tld  string
		Name string
	}
	var keys []nameKey
	k.IterateNameRecords(ctx, func(record types.NameRecord) bool {
		keys = append(keys, nameKey{Tld: record.Tld, Name: record.Name})
		return false
	})

	sort.Slice(keys, func(i, j int) bool {
		ki := keys[i].Tld + "/" + keys[i].Name
		kj := keys[j].Tld + "/" + keys[j].Name
		return ki < kj
	})

	numLeaves := uint32(len(keys))
	// Capacity must be a power of 2, minimum 2
	capacity := uint32(2)
	for capacity < numLeaves {
		capacity *= 2
	}

	// Assign leaf indices and compute leaf hashes
	for i, nk := range keys {
		k.SetLeafIndex(ctx, nk.Tld, nk.Name, uint32(i))
		leafHash := k.computeLeafHash(ctx, nk.Tld, nk.Name)
		k.SetMerkleNode(ctx, 0, uint32(i), leafHash)
	}

	// Fill remaining leaves with empty hash
	for i := numLeaves; i < capacity; i++ {
		k.SetMerkleNode(ctx, 0, i, emptyLeafHash)
	}

	// Build tree bottom-up
	height := treeHeight(capacity)
	for level := uint16(0); level < height; level++ {
		nodesAtLevel := capacity >> level
		for i := uint32(0); i < nodesAtLevel; i += 2 {
			left := k.GetMerkleNode(ctx, level, i)
			right := k.GetMerkleNode(ctx, level, i+1)
			parent := hashPair(left, right)
			k.SetMerkleNode(ctx, level+1, i/2, parent)
		}
	}

	meta := MerkleMetadata{
		NumLeaves: numLeaves,
		Capacity:  capacity,
	}
	k.SetMerkleMetadata(ctx, meta)

	root := k.GetMerkleNode(ctx, height, 0)
	rootHex := hex.EncodeToString(root[:])

	// Update the state digest with the Merkle root
	sd := types.StateDigest{
		Digest:   rootHex,
		Height:   uint64(ctx.BlockHeight()),
		NumNames: uint64(numLeaves),
	}
	k.SetStateDigest(ctx, sd)

	return rootHex
}

// UpdateDomainInTree updates a domain's leaf in the Merkle tree, stores the new
// root as the state digest, clears stale attestations, and emits an event.
func (k Keeper) UpdateDomainInTree(ctx sdk.Context, tld, name string) {
	// Clear attestations for the old digest
	if oldDigest, found := k.GetStateDigest(ctx); found {
		k.ClearAttestationsForDigest(ctx, oldDigest.Digest)
	}

	rootHex := k.UpdateLeaf(ctx, tld, name)
	meta := k.GetMerkleMetadata(ctx)

	sd := types.StateDigest{
		Digest:   rootHex,
		Height:   uint64(ctx.BlockHeight()),
		NumNames: uint64(meta.NumLeaves),
	}
	k.SetStateDigest(ctx, sd)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"xid_state_digest_updated",
		sdk.NewAttribute("digest", rootHex),
		sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
		sdk.NewAttribute("num_names", fmt.Sprintf("%d", meta.NumLeaves)),
	))
}
