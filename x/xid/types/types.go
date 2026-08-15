package types

// This file holds hand-written helpers/constants for the xID types. The message
// types themselves (ContentRoot, LinkedIdentity, StateDigest, Attestation, …) are
// generated into xid.pb.go by `make proto-gen`; earlier they were hand-copied here
// as a placeholder ("run make proto-gen then replace") and have now been replaced.

// FullName returns the complete domain name (e.g. "alice.epix")
func (n NameRecord) FullName() string {
	return n.Name + "." + n.Tld
}

// DNS record type constants
const (
	DNSRecordTypeA       uint32 = 1
	DNSRecordTypeAAAA    uint32 = 28
	DNSRecordTypeCNAME   uint32 = 5
	DNSRecordTypeTXT     uint32 = 16
	DNSRecordTypeMX      uint32 = 15
	DNSRecordTypeNS      uint32 = 2
	DNSRecordTypeSRV     uint32 = 33
	DNSRecordTypeEpixNet uint32 = 65280 // Private-use range (RFC 6895) for linked identity discovery
)
