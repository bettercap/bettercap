package dhcp6opts

import (
	"io"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// StatusCode represents a Status Code, as defined in RFC 3315, Section 5.4.
// DHCP clients and servers can use status codes to communicate successes
// or failures, and provide additional information using a message to describe
// specific failures.
type StatusCode struct {
	// Code specifies the Status value stored within this StatusCode, such as
	// StatusSuccess, StatusUnspecFail, etc.
	Code dhcp6.Status

	// Message specifies a human-readable message within this StatusCode, useful
	// for providing information about successes or failures.
	Message string
}

// NewStatusCode creates a new StatusCode from an input Status value and a
// string message.
func NewStatusCode(code dhcp6.Status, message string) *StatusCode {
	return &StatusCode{
		Code:    code,
		Message: message,
	}
}

// MarshalBinary allocates a byte slice containing the data from a StatusCode.
func (s *StatusCode) MarshalBinary() ([]byte, error) {
	// 2 bytes: status code
	// N bytes: message
	b := buffer.New(nil)
	b.Write16(uint16(s.Code))
	b.WriteBytes([]byte(s.Message))
	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a StatusCode.
//
// If the byte slice does not contain enough data to form a valid StatusCode,
// errInvalidStatusCode is returned.
func (s *StatusCode) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Too short to contain valid StatusCode
	if b.Len() < 2 {
		return io.ErrUnexpectedEOF
	}

	s.Code = dhcp6.Status(b.Read16())
	s.Message = string(b.Remaining())
	return nil
}
