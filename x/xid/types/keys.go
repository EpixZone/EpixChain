package types

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// constants
const (
	// ModuleName defines the module name
	ModuleName = "xid"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// ModuleAddress is the native module address for xID
var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the xID persistent store
const (
	prefixNameRecord = iota + 1
	prefixOwnerIndex
	prefixProfile
	prefixDNSRecord
	prefixTLDConfig
	prefixParams
	prefixOwnerCount
	prefixGlobalNameCount
	prefixTLDNameCount
	prefixGlobalFeesBurned
	prefixTLDFeesBurned
	prefixEpixNetPeer
	prefixEpixNetPeerReverse
)

// KVStore key prefixes
var (
	KeyPrefixNameRecord = []byte{prefixNameRecord}
	KeyPrefixOwnerIndex = []byte{prefixOwnerIndex}
	KeyPrefixProfile    = []byte{prefixProfile}
	KeyPrefixDNSRecord  = []byte{prefixDNSRecord}
	KeyPrefixTLDConfig  = []byte{prefixTLDConfig}
	KeyPrefixParams     = []byte{prefixParams}
	KeyPrefixOwnerCount     = []byte{prefixOwnerCount}
	KeyGlobalNameCount      = []byte{prefixGlobalNameCount}
	KeyGlobalFeesBurned     = []byte{prefixGlobalFeesBurned}
)

// NameRecordKey returns the store key for a name record: [prefix][len(tld)][tld][name]
func NameRecordKey(tld, name string) []byte {
	tldBytes := []byte(tld)
	nameBytes := []byte(name)
	key := make([]byte, 0, 1+1+len(tldBytes)+len(nameBytes))
	key = append(key, prefixNameRecord)
	key = append(key, byte(len(tldBytes)))
	key = append(key, tldBytes...)
	key = append(key, nameBytes...)
	return key
}

// OwnerIndexKey returns the store key for the owner index: [prefix][owner_bytes(20)][len(tld)][tld][name]
func OwnerIndexKey(owner []byte, tld, name string) []byte {
	tldBytes := []byte(tld)
	nameBytes := []byte(name)
	key := make([]byte, 0, 1+20+1+len(tldBytes)+len(nameBytes))
	key = append(key, prefixOwnerIndex)
	key = append(key, padOrTruncate(owner, 20)...)
	key = append(key, byte(len(tldBytes)))
	key = append(key, tldBytes...)
	key = append(key, nameBytes...)
	return key
}

// OwnerIndexPrefix returns the prefix for iterating all names owned by an address
func OwnerIndexPrefix(owner []byte) []byte {
	key := make([]byte, 0, 1+20)
	key = append(key, prefixOwnerIndex)
	key = append(key, padOrTruncate(owner, 20)...)
	return key
}

// ProfileKey returns the store key for a profile: [prefix][len(tld)][tld][name]
func ProfileKey(tld, name string) []byte {
	tldBytes := []byte(tld)
	nameBytes := []byte(name)
	key := make([]byte, 0, 1+1+len(tldBytes)+len(nameBytes))
	key = append(key, prefixProfile)
	key = append(key, byte(len(tldBytes)))
	key = append(key, tldBytes...)
	key = append(key, nameBytes...)
	return key
}

// DNSRecordKey returns the store key for a DNS record: [prefix][len(tld)][tld][len(name)][name][recordType(2)]
func DNSRecordKey(tld, name string, recordType uint32) []byte {
	tldBytes := []byte(tld)
	nameBytes := []byte(name)
	key := make([]byte, 0, 1+1+len(tldBytes)+1+len(nameBytes)+2)
	key = append(key, prefixDNSRecord)
	key = append(key, byte(len(tldBytes)))
	key = append(key, tldBytes...)
	key = append(key, byte(len(nameBytes)))
	key = append(key, nameBytes...)
	rt := make([]byte, 2)
	binary.BigEndian.PutUint16(rt, uint16(recordType))
	key = append(key, rt...)
	return key
}

// DNSRecordPrefix returns the prefix for iterating all DNS records for a name
func DNSRecordPrefix(tld, name string) []byte {
	tldBytes := []byte(tld)
	nameBytes := []byte(name)
	key := make([]byte, 0, 1+1+len(tldBytes)+1+len(nameBytes))
	key = append(key, prefixDNSRecord)
	key = append(key, byte(len(tldBytes)))
	key = append(key, tldBytes...)
	key = append(key, byte(len(nameBytes)))
	key = append(key, nameBytes...)
	return key
}

// TLDConfigKey returns the store key for a TLD config: [prefix][tld]
func TLDConfigKey(tld string) []byte {
	key := make([]byte, 0, 1+len(tld))
	key = append(key, prefixTLDConfig)
	key = append(key, []byte(tld)...)
	return key
}

// OwnerCountKey returns the store key for an owner's name count: [prefix][owner_bytes(20)]
func OwnerCountKey(owner []byte) []byte {
	key := make([]byte, 0, 1+20)
	key = append(key, prefixOwnerCount)
	key = append(key, padOrTruncate(owner, 20)...)
	return key
}

// TLDNameCountKey returns the store key for a TLD's name count: [prefix][tld]
func TLDNameCountKey(tld string) []byte {
	key := make([]byte, 0, 1+len(tld))
	key = append(key, prefixTLDNameCount)
	key = append(key, []byte(tld)...)
	return key
}

// TLDFeesBurnedKey returns the store key for a TLD's burned fee total: [prefix][tld]
func TLDFeesBurnedKey(tld string) []byte {
	key := make([]byte, 0, 1+len(tld))
	key = append(key, prefixTLDFeesBurned)
	key = append(key, []byte(tld)...)
	return key
}

// EpixNetPeerKey returns the store key for an EpixNet peer:
// [prefix][len(tld)][tld][len(name)][name][sha256(address)[:8]]
func EpixNetPeerKey(tld, name, address string) []byte {
	tldBytes := []byte(tld)
	nameBytes := []byte(name)
	addrHash := sha256.Sum256([]byte(address))
	key := make([]byte, 0, 1+1+len(tldBytes)+1+len(nameBytes)+8)
	key = append(key, prefixEpixNetPeer)
	key = append(key, byte(len(tldBytes)))
	key = append(key, tldBytes...)
	key = append(key, byte(len(nameBytes)))
	key = append(key, nameBytes...)
	key = append(key, addrHash[:8]...)
	return key
}

// EpixNetPeerPrefix returns the prefix for iterating all EpixNet peers for a name
func EpixNetPeerPrefix(tld, name string) []byte {
	tldBytes := []byte(tld)
	nameBytes := []byte(name)
	key := make([]byte, 0, 1+1+len(tldBytes)+1+len(nameBytes))
	key = append(key, prefixEpixNetPeer)
	key = append(key, byte(len(tldBytes)))
	key = append(key, tldBytes...)
	key = append(key, byte(len(nameBytes)))
	key = append(key, nameBytes...)
	return key
}

// EpixNetPeerReverseKey returns the store key for the peer reverse index:
// [prefix][sha256(address)[:8]]
// The value stores the tld and name that this peer address is linked to.
func EpixNetPeerReverseKey(address string) []byte {
	addrHash := sha256.Sum256([]byte(address))
	key := make([]byte, 0, 1+8)
	key = append(key, prefixEpixNetPeerReverse)
	key = append(key, addrHash[:8]...)
	return key
}

// padOrTruncate ensures the byte slice is exactly the desired length
func padOrTruncate(b []byte, length int) []byte {
	if len(b) >= length {
		return b[:length]
	}
	result := make([]byte, length)
	copy(result[length-len(b):], b)
	return result
}
