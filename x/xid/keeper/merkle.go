package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/evm/x/xid/types"
)

// ComputePeerMerkleRoot computes a SHA-256 Merkle root from the active peers
// of a given name. Returns an empty string if there are no active peers.
// Peers are sorted lexicographically by address for determinism.
func (k Keeper) ComputePeerMerkleRoot(ctx sdk.Context, tld, name string) string {
	peers := k.GetAllEpixNetPeers(ctx, tld, name)

	var addresses []string
	for _, p := range peers {
		if p.Active {
			addresses = append(addresses, p.Address)
		}
	}

	if len(addresses) == 0 {
		return ""
	}

	sort.Strings(addresses)

	// Hash each address to form leaf nodes
	hashes := make([][]byte, len(addresses))
	for i, addr := range addresses {
		h := sha256.Sum256([]byte(addr))
		hashes[i] = h[:]
	}

	// Build Merkle tree by hashing pairs
	for len(hashes) > 1 {
		var next [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := append(hashes[i], hashes[i+1]...)
				h := sha256.Sum256(combined)
				next = append(next, h[:])
			} else {
				next = append(next, hashes[i])
			}
		}
		hashes = next
	}

	// Mix in the xID name as a domain separator so identical peer sets
	// under different names produce distinct roots.
	domain := sha256.Sum256([]byte(name + "." + tld))
	final := sha256.Sum256(append(domain[:], hashes[0]...))

	return hex.EncodeToString(final[:])
}

// RecomputeAndStoreContentRoot recomputes the Merkle root for a name's active
// peers and stores it. Returns the new root string.
func (k Keeper) RecomputeAndStoreContentRoot(ctx sdk.Context, tld, name string) string {
	root := k.ComputePeerMerkleRoot(ctx, tld, name)
	contentRoot := types.ContentRoot{
		Root:      root,
		UpdatedAt: uint64(ctx.BlockHeight()),
	}
	k.SetContentRoot(ctx, tld, name, contentRoot)
	return root
}
