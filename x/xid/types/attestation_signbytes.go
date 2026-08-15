package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

// AttestDomain is the fixed 16-byte domain tag prefixed to every attestation
// sign-bytes so a validator's xID attestation signature can never be
// reinterpreted as a CometBFT consensus vote (whose canonical sign-bytes begin
// with a protobuf length prefix, never these ASCII bytes) or vice versa.
//
// MUST equal the client's `ATTEST_DOMAIN` in
// EpixNet crates/epix-chain/src/finality.rs byte-for-byte.
var AttestDomain = [16]byte{'E', 'P', 'I', 'X', '-', 'X', 'I', 'D', '-', 'A', 'T', 'T', 'E', 'S', 'T', '1'}

// AttestationSignBytes is the exact, canonical message a validator signs with its
// ed25519 attestation key for (chain_id, height, block_time, digest):
//
//	domain(16) ‖ len(chain_id) u32-BE ‖ chain_id ‖ height u64-BE ‖
//	block_time i64-BE ‖ len(digest) u32-BE ‖ digest
//
// `digestHex` is the hex tree root; the RAW 32 bytes are hashed (not the hex),
// matching the client which decodes the hex before verifying. Fully length-
// delimited + fixed-width so no two distinct tuples share a preimage.
//
// This function is the cross-repo contract: it MUST produce byte-identical output
// to `attest_sign_bytes` in EpixNet crates/epix-chain/src/finality.rs. A frozen
// KAT in both repos guards it.
func AttestationSignBytes(chainID string, height uint64, blockTime int64, digestHex string) []byte {
	digest, _ := hex.DecodeString(digestHex)
	var b bytes.Buffer
	b.Write(AttestDomain[:])
	_ = binary.Write(&b, binary.BigEndian, uint32(len(chainID)))
	b.WriteString(chainID)
	_ = binary.Write(&b, binary.BigEndian, height)
	_ = binary.Write(&b, binary.BigEndian, blockTime)
	_ = binary.Write(&b, binary.BigEndian, uint32(len(digest)))
	b.Write(digest)
	return b.Bytes()
}
