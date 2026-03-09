package keeper

import (
	"encoding/binary"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/vrf/types"
)

// SetBeacon stores a random beacon at a given height
func (k Keeper) SetBeacon(ctx sdk.Context, beacon types.RandomBeacon) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(beacon)
	if err != nil {
		panic(err)
	}
	store.Set(types.BeaconKey(beacon.Height), bz)
}

// GetBeaconByHeight retrieves a random beacon for a given height
func (k Keeper) GetBeaconByHeight(ctx sdk.Context, height uint64) (types.RandomBeacon, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BeaconKey(height))
	if bz == nil {
		return types.RandomBeacon{}, false
	}
	var beacon types.RandomBeacon
	if err := json.Unmarshal(bz, &beacon); err != nil {
		return types.RandomBeacon{}, false
	}
	return beacon, true
}

// GetLatestBeacon retrieves the most recent beacon
func (k Keeper) GetLatestBeacon(ctx sdk.Context) (types.RandomBeacon, bool) {
	height, found := k.GetLatestHeight(ctx)
	if !found {
		return types.RandomBeacon{}, false
	}
	return k.GetBeaconByHeight(ctx, height)
}

// SetLatestHeight stores the height of the most recent beacon
func (k Keeper) SetLatestHeight(ctx sdk.Context, height uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, height)
	store.Set(types.KeyLatestHeight, bz)
}

// GetLatestHeight retrieves the height of the most recent beacon
func (k Keeper) GetLatestHeight(ctx sdk.Context) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyLatestHeight)
	if bz == nil {
		return 0, false
	}
	return binary.BigEndian.Uint64(bz), true
}

// PruneBeaconsBelow removes all beacons below the given height
func (k Keeper) PruneBeaconsBelow(ctx sdk.Context, minHeight uint64) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.KeyPrefixBeacon, types.BeaconKey(minHeight))
	defer iterator.Close()

	keysToDelete := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		keysToDelete = append(keysToDelete, append([]byte{}, iterator.Key()...))
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// IterateBeacons iterates over all beacons and calls the callback.
// If callback returns true, iteration stops.
func (k Keeper) IterateBeacons(ctx sdk.Context, cb func(beacon types.RandomBeacon) bool) {
	store := ctx.KVStore(k.storeKey)
	// Iterate from prefix to end of beacon range
	endKey := make([]byte, 1+8)
	endKey[0] = types.KeyPrefixBeacon[0]
	for i := 1; i < 9; i++ {
		endKey[i] = 0xFF
	}
	iterator := store.Iterator(types.KeyPrefixBeacon, endKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var beacon types.RandomBeacon
		if err := json.Unmarshal(iterator.Value(), &beacon); err != nil {
			continue
		}
		if cb(beacon) {
			break
		}
	}
}
