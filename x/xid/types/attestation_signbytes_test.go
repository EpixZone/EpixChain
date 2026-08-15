package types

import (
	"encoding/hex"
	"testing"
)

// TestAttestationSignBytesKAT is the frozen cross-repo Known-Answer Test. The
// exact same vector is asserted in EpixNet crates/epix-chain/src/finality.rs
// (test `attest_sign_bytes_kat`). If either side changes encoding, one of the two
// KATs breaks — that is the guard against silent Go<->Rust divergence.
func TestAttestationSignBytesKAT(t *testing.T) {
	// Vector: chain_id "epix_1916-1", height 200, block_time 1000090,
	// digest = 32 bytes of 0x11.
	digestHex := "1111111111111111111111111111111111111111111111111111111111111111"
	got := hex.EncodeToString(AttestationSignBytes("epix_1916-1", 200, 1000090, digestHex))

	const want = "455049582d5849442d41545445535431" + // domain "EPIX-XID-ATTEST1"
		"0000000b" + // len(chain_id)=11
		"657069785f313931362d31" + // "epix_1916-1"
		"00000000000000c8" + // height=200 (u64 BE)
		"00000000000f429a" + // block_time=1000090 (i64 BE)
		"00000020" + // len(digest)=32
		"1111111111111111111111111111111111111111111111111111111111111111"

	if got != want {
		t.Fatalf("sign-bytes KAT mismatch:\n got=%s\nwant=%s", got, want)
	}
}
