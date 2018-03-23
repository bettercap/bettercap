package dhcp6opts

import (
	"io"

	"github.com/mdlayher/dhcp6/internal/buffer"
)

// A RemoteIdentifier carries vendor-specific options.
//
// The vendor is indicated in the enterprise-number field.
// The remote-id field may be used to encode, for instance:
// - a "caller ID" telephone number for dial-up connection
// - a "user name" prompted for by a Remote Access Server
// - a remote caller ATM address
// - a "modem ID" of a cable data modem
// - the remote IP address of a point-to-point link
// - a remote X.25 address for X.25 connections
// - an interface or port identifier
type RemoteIdentifier struct {
	// EnterpriseNumber specifies an IANA-assigned vendor Private Enterprise
	// Number.
	EnterpriseNumber uint32

	// The opaque value for the remote-id.
	RemoteID []byte
}

// MarshalBinary allocates a byte slice containing the data
// from a RemoteIdentifier.
func (r *RemoteIdentifier) MarshalBinary() ([]byte, error) {
	// 4 bytes: EnterpriseNumber
	// N bytes: RemoteId
	b := buffer.New(nil)
	b.Write32(r.EnterpriseNumber)
	b.WriteBytes(r.RemoteID)
	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a RemoteIdentifier.
// If the byte slice does not contain enough data to form a valid
// RemoteIdentifier, io.ErrUnexpectedEOF is returned.
func (r *RemoteIdentifier) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Too short to be valid RemoteIdentifier
	if b.Len() < 5 {
		return io.ErrUnexpectedEOF
	}

	r.EnterpriseNumber = b.Read32()
	r.RemoteID = b.Remaining()
	return nil
}
