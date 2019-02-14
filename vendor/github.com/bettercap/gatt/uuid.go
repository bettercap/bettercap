package gatt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// A UUID is a BLE UUID.
type UUID struct {
	// Hide the bytes, so that we can enforce that they have length 2 or 16,
	// and that they are immutable. This simplifies the code and API.
	b []byte
}

// UUID16 converts a uint16 (such as 0x1800) to a UUID.
func UUID16(i uint16) UUID {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return UUID{b}
}

// ParseUUID parses a standard-format UUID string, such
// as "1800" or "34DA3AD1-7110-41A1-B1EF-4430F509CDE7".
func ParseUUID(s string) (UUID, error) {
	s = strings.Replace(s, "-", "", -1)
	b, err := hex.DecodeString(s)
	if err != nil {
		return UUID{}, err
	}
	if err := lenErr(len(b)); err != nil {
		return UUID{}, err
	}
	return UUID{reverse(b)}, nil
}

// MustParseUUID parses a standard-format UUID string,
// like ParseUUID, but panics in case of error.
func MustParseUUID(s string) UUID {
	u, err := ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// lenErr returns an error if n is an invalid UUID length.
func lenErr(n int) error {
	switch n {
	case 2, 16:
		return nil
	}
	return fmt.Errorf("UUIDs must have length 2 or 16, got %d", n)
}

// Len returns the length of the UUID, in bytes.
// BLE UUIDs are either 2 or 16 bytes.
func (u UUID) Len() int {
	return len(u.b)
}

// String hex-encodes a UUID.
func (u UUID) String() string {
	return fmt.Sprintf("%x", reverse(u.b))
}

func (u UUID) Bytes() []byte {
	return u.b
}

// Equal returns a boolean reporting whether v represent the same UUID as u.
func (u UUID) Equal(v UUID) bool {
	return bytes.Equal(u.b, v.b)
}

// UUIDContains returns a boolean reporting whether u is in the slice s.
func UUIDContains(s []UUID, u UUID) bool {
	if s == nil {
		return true
	}

	for _, a := range s {
		if a.Equal(u) {
			return true
		}
	}

	return false
}

// reverse returns a reversed copy of u.
func reverse(u []byte) []byte {
	// Special-case 16 bit UUIDS for speed.
	l := len(u)
	if l == 2 {
		return []byte{u[1], u[0]}
	}
	b := make([]byte, l)
	for i := 0; i < l/2+1; i++ {
		b[i], b[l-i-1] = u[l-i-1], u[i]
	}
	return b
}
