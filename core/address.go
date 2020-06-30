package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// This is the main regex where an address should confirm to. Much simpler than an email address
	addressRegex string = "(^[a-z0-9][a-z0-9\\.\\-]{2,63})(?:@([a-z0-9][a-z0-9\\.\\-]{1,63}))?!$"
)

// HashAddress is a SHA256'd address
type HashAddress string

// String casts an hash address to string
func (ha HashAddress) String() string {
	return string(ha)
}

// Address represents a bitMaelum address
type Address struct {
	Local string
	Org   string
}

// StringToHash converts a string to a hash address
func StringToHash(address string) HashAddress {
	sum := sha256.Sum256([]byte(address))
	return HashAddress(hex.EncodeToString(sum[:]))
}

// IsValidAddress returns true when the given string is a valid BitMaelum address
func IsValidAddress(address string) bool {
	_, err := NewAddressFromString(address)
	return err == nil
}

// NewAddressFromString returns a valid address structure based on the given address
func NewAddressFromString(address string) (*Address, error) {
	re := regexp.MustCompile(addressRegex)
	if re == nil {
		return nil, errors.New("cannot compile regex")
	}

	if !re.MatchString(strings.ToLower(address)) {
		return nil, errors.New("incorrect address format specified")
	}

	matches := re.FindStringSubmatch(strings.ToLower(address))

	return &Address{
		Local: matches[1],
		Org:   matches[2],
	}, nil
}

// String converts an address to a string
func (a *Address) String() string {
	if len(a.Org) == 0 {
		return fmt.Sprintf("%s!", a.Local)
	}

	return fmt.Sprintf("%s@%s!", a.Local, a.Org)
}

// Hash converts an address to a hashed value
func (a *Address) Hash() HashAddress {
	return StringToHash(a.String())
}

// Bytes converts an address to []byte
func (a *Address) Bytes() []byte {
	return []byte(a.String())
}
