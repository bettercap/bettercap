package dhcp6opts

import (
	"io"

	"github.com/mdlayher/dhcp6/internal/buffer"
)

// The Authentication option carries authentication information to
// authenticate the identity and contents of DHCP messages. The use of
// the Authentication option is described in section 21.
type Authentication struct {
	// The authentication protocol used in this authentication option
	Protocol byte

	// The algorithm used in the authentication protocol
	Algorithm byte

	// The replay detection method used in this authentication option
	RDM byte

	// The replay detection information for the RDM
	ReplayDetection uint64

	// The authentication information, as specified by the protocol and
	// algorithm used in this authentication option.
	AuthenticationInformation []byte
}

// MarshalBinary allocates a byte slice containing the data from a Authentication.
func (a *Authentication) MarshalBinary() ([]byte, error) {
	// 1 byte:  Protocol
	// 1 byte:  Algorithm
	// 1 byte:  RDM
	// 8 bytes: ReplayDetection
	// N bytes: AuthenticationInformation (can have 0 len byte)
	b := buffer.New(nil)
	b.Write8(a.Protocol)
	b.Write8(a.Algorithm)
	b.Write8(a.RDM)
	b.Write64(a.ReplayDetection)
	b.WriteBytes(a.AuthenticationInformation)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a Authentication.
// If the byte slice does not contain enough data to form a valid
// Authentication, io.ErrUnexpectedEOF is returned.
func (a *Authentication) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Too short to be valid Authentication
	if b.Len() < 11 {
		return io.ErrUnexpectedEOF
	}

	a.Protocol = b.Read8()
	a.Algorithm = b.Read8()
	a.RDM = b.Read8()
	a.ReplayDetection = b.Read64()
	a.AuthenticationInformation = b.Remaining()
	return nil
}
