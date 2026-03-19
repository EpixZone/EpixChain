package types

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// nameRegex matches lowercase alphanumeric names with hyphens (no leading/trailing hyphens)
var nameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// tldRegex matches lowercase alphabetic TLD strings
var tldRegex = regexp.MustCompile(`^[a-z]{2,16}$`)

// ValidateName checks that a name label is valid
func ValidateName(name string) error {
	if len(name) == 0 {
		return ErrInvalidName.Wrap("name cannot be empty")
	}
	if len(name) > 64 {
		return ErrInvalidName.Wrap("name exceeds maximum length of 64 characters")
	}
	name = strings.ToLower(name)
	if !nameRegex.MatchString(name) {
		return ErrInvalidName.Wrap("name must be lowercase alphanumeric with optional hyphens, cannot start or end with a hyphen")
	}
	return nil
}

// ValidateTLD checks that a TLD string is valid
func ValidateTLD(tld string) error {
	if len(tld) == 0 {
		return ErrInvalidTLD.Wrap("TLD cannot be empty")
	}
	tld = strings.ToLower(tld)
	if !tldRegex.MatchString(tld) {
		return ErrInvalidTLD.Wrap("TLD must be 2-16 lowercase alphabetic characters")
	}
	return nil
}

// ValidatePriceTiers checks that a set of price tiers is well-formed:
// at least one tier, positive prices, positive MaxLength, no duplicates, sorted ascending.
func ValidatePriceTiers(tiers []PriceTier) error {
	if len(tiers) == 0 {
		return ErrInvalidPriceTier.Wrap("at least one price tier is required")
	}

	seen := make(map[uint32]bool, len(tiers))
	var prevMax uint32
	for i, tier := range tiers {
		if tier.MaxLength == 0 {
			return ErrInvalidPriceTier.Wrapf("tier %d: max_length must be greater than 0", i)
		}
		if !tier.Price.IsPositive() {
			return ErrInvalidPriceTier.Wrapf("tier %d: price must be positive", i)
		}
		if seen[tier.MaxLength] {
			return ErrInvalidPriceTier.Wrapf("tier %d: duplicate max_length %d", i, tier.MaxLength)
		}
		seen[tier.MaxLength] = true
		if i > 0 && tier.MaxLength <= prevMax {
			return ErrInvalidPriceTier.Wrapf("tier %d: max_length %d must be greater than previous tier's %d (tiers must be sorted ascending)", i, tier.MaxLength, prevMax)
		}
		prevMax = tier.MaxLength
	}

	return nil
}

// validDNSRecordTypes is the set of supported DNS record types.
var validDNSRecordTypes = map[uint32]bool{
	DNSRecordTypeA:       true,
	DNSRecordTypeNS:      true,
	DNSRecordTypeCNAME:   true,
	DNSRecordTypeMX:      true,
	DNSRecordTypeTXT:     true,
	DNSRecordTypeAAAA:    true,
	DNSRecordTypeSRV:     true,
	DNSRecordTypeEpixNet: true,
}

// ValidateDNSRecordType checks that a record type is in the supported set.
func ValidateDNSRecordType(rt uint32) error {
	if !validDNSRecordTypes[rt] {
		return ErrInvalidDNSRecord.Wrapf("unsupported record type %d", rt)
	}
	return nil
}

const (
	// MaxDNSTTL is the maximum allowed TTL (7 days).
	MaxDNSTTL uint32 = 604800
	// MinDNSTTL is the minimum allowed TTL (60 seconds). TTL=0 is also allowed (resolver default).
	MinDNSTTL uint32 = 60
	// MaxDNSValueLength is the maximum length of a DNS record value.
	MaxDNSValueLength = 1024
)

// ValidateDNSTTL checks that a TTL is within acceptable bounds.
// TTL=0 is allowed (means "use resolver default").
func ValidateDNSTTL(ttl uint32) error {
	if ttl == 0 {
		return nil
	}
	if ttl < MinDNSTTL {
		return ErrInvalidDNSRecord.Wrapf("TTL %d is below minimum of %d seconds", ttl, MinDNSTTL)
	}
	if ttl > MaxDNSTTL {
		return ErrInvalidDNSRecord.Wrapf("TTL %d exceeds maximum of %d seconds (7 days)", ttl, MaxDNSTTL)
	}
	return nil
}

// ValidateHexDigest checks that a string is a valid hex-encoded digest of the expected length.
func ValidateHexDigest(digest string, expectedLen int) error {
	if len(digest) != expectedLen {
		return fmt.Errorf("digest must be %d characters, got %d", expectedLen, len(digest))
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return fmt.Errorf("digest must be valid hexadecimal: %w", err)
	}
	return nil
}

// Profile validation constants
const (
	MaxAvatarLength        = 256
	MaxBioLength           = 512
	MaxProfileMetadataKeys = 20
	MaxMetadataKeyLength   = 64
	MaxMetadataValueLength = 256
)

// ValidateProfile checks that a profile's fields are within size limits.
func ValidateProfile(profile Profile) error {
	if len(profile.Avatar) > MaxAvatarLength {
		return ErrInvalidProfile.Wrapf("avatar exceeds maximum length of %d characters", MaxAvatarLength)
	}
	if len(profile.Bio) > MaxBioLength {
		return ErrInvalidProfile.Wrapf("bio exceeds maximum length of %d characters", MaxBioLength)
	}
	if len(profile.Metadata) > MaxProfileMetadataKeys {
		return ErrInvalidProfile.Wrapf("metadata exceeds maximum of %d entries", MaxProfileMetadataKeys)
	}
	for i, kv := range profile.Metadata {
		if len(kv.Key) == 0 {
			return ErrInvalidProfile.Wrapf("metadata entry %d: key cannot be empty", i)
		}
		if len(kv.Key) > MaxMetadataKeyLength {
			return ErrInvalidProfile.Wrapf("metadata entry %d: key exceeds maximum length of %d characters", i, MaxMetadataKeyLength)
		}
		if len(kv.Value) > MaxMetadataValueLength {
			return ErrInvalidProfile.Wrapf("metadata entry %d: value exceeds maximum length of %d characters", i, MaxMetadataValueLength)
		}
	}
	return nil
}
