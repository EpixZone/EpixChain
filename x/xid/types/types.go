package types

// FullName returns the complete domain name (e.g. "alice.epix")
func (n NameRecord) FullName() string {
	return n.Name + "." + n.Tld
}

// DNS record type constants
const (
	DNSRecordTypeA     uint32 = 1
	DNSRecordTypeAAAA  uint32 = 28
	DNSRecordTypeCNAME uint32 = 5
	DNSRecordTypeTXT   uint32 = 16
	DNSRecordTypeMX    uint32 = 15
	DNSRecordTypeNS    uint32 = 2
	DNSRecordTypeSRV   uint32 = 33
)
