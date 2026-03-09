package types_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/evm/x/vrf/types"
)

func TestBeaconKey(t *testing.T) {
	// Height 0: prefix + 8 zero bytes
	key0 := types.BeaconKey(0)
	require.Len(t, key0, 9)
	require.Equal(t, byte(1), key0[0]) // prefixBeacon = iota + 1 = 1
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, key0[1:9])

	// Height 1
	key1 := types.BeaconKey(1)
	require.Len(t, key1, 9)
	require.Equal(t, byte(1), key1[0])
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 1}, key1[1:9])

	// Max height
	keyMax := types.BeaconKey(math.MaxUint64)
	require.Len(t, keyMax, 9)
	require.Equal(t, byte(1), keyMax[0])
	require.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, keyMax[1:9])
}

func TestBeaconKeyOrdering(t *testing.T) {
	// Big-endian encoding ensures lexicographic ordering matches numeric ordering
	key100 := types.BeaconKey(100)
	key200 := types.BeaconKey(200)
	require.True(t, bytes.Compare(key100, key200) < 0, "BeaconKey(100) should be < BeaconKey(200)")

	key0 := types.BeaconKey(0)
	key1 := types.BeaconKey(1)
	require.True(t, bytes.Compare(key0, key1) < 0, "BeaconKey(0) should be < BeaconKey(1)")

	keyMaxMinus1 := types.BeaconKey(math.MaxUint64 - 1)
	keyMax := types.BeaconKey(math.MaxUint64)
	require.True(t, bytes.Compare(keyMaxMinus1, keyMax) < 0, "BeaconKey(max-1) should be < BeaconKey(max)")
}

func TestKeyConstants(t *testing.T) {
	// All prefix keys should be distinct
	require.NotEqual(t, types.KeyPrefixBeacon, types.KeyLatestHeight)
	require.NotEqual(t, types.KeyPrefixBeacon, types.KeyPrefixParams)
	require.NotEqual(t, types.KeyLatestHeight, types.KeyPrefixParams)

	// Verify expected values
	require.Equal(t, []byte{1}, types.KeyPrefixBeacon)
	require.Equal(t, []byte{2}, types.KeyLatestHeight)
	require.Equal(t, []byte{3}, types.KeyPrefixParams)
}
