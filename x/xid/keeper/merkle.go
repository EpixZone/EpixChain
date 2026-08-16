package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/cosmos/evm/x/xid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ComputePeerMerkleRoot computes a SHA-256 Merkle root from the active linked
// identities of a given name. Returns an empty string if there are no active
// identities. Identities are sorted lexicographically by address for determinism.
func (k Keeper) ComputePeerMerkleRoot(ctx sdk.Context, tld, name string) string {
	identities := k.GetAllLinkedIdentities(ctx, tld, name)

	var addresses []string
	for _, id := range identities {
		if id.Active {
			addresses = append(addresses, id.Address)
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

	// Mix in the xID name as a domain separator so identical identity sets
	// under different names produce distinct roots.
	domain := sha256.Sum256([]byte(name + "." + tld))
	final := sha256.Sum256(append(domain[:], hashes[0]...))

	return hex.EncodeToString(final[:])
}

// RecomputeAndStoreContentRoot recomputes the Merkle root for a name's active
// linked identities and stores it. Returns the new root string.
func (k Keeper) RecomputeAndStoreContentRoot(ctx sdk.Context, tld, name string) string {
	root := k.ComputePeerMerkleRoot(ctx, tld, name)
	contentRoot := types.ContentRoot{
		Root:      root,
		UpdatedAt: uint64(ctx.BlockHeight()),
	}
	k.SetContentRoot(ctx, tld, name, contentRoot)
	return root
}
