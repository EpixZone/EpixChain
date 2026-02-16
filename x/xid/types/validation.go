package types

import (
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
