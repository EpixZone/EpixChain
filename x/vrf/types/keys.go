package types

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	ModuleName = "vrf"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the VRF persistent store
const (
	prefixBeacon = iota + 1
	prefixLatestHeight
	prefixParams
)

var (
	KeyPrefixBeacon = []byte{prefixBeacon}
	KeyLatestHeight = []byte{prefixLatestHeight}
	KeyPrefixParams = []byte{prefixParams}
)

// BeaconKey returns the store key for a beacon at a given height: [prefix][height_8bytes_BE]
func BeaconKey(height uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = prefixBeacon
	binary.BigEndian.PutUint64(key[1:9], height)
	return key
}
