package dhcp6opts

import (
	"io"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// IATA represents an Identity Association for Temporary Addresses, as
// defined in RFC 3315, Section 22.5.
//
// Multiple IATAs may be present in a single DHCP request.
type IATA struct {
	// IAID specifies a DHCP identity association identifier.  The IAID
	// is a unique, client-generated identifier.
	IAID [4]byte

	// Options specifies a map of DHCP options specific to this IATA.
	// Its methods can be used to retrieve data from an incoming IATA, or send
	// data with an outgoing IATA.
	Options dhcp6.Options
}

// NewIATA creates a new IATA from an IAID and an Options map.  If an Options
// map is not specified, a new one will be allocated.
func NewIATA(iaid [4]byte, options dhcp6.Options) *IATA {
	if options == nil {
		options = make(dhcp6.Options)
	}

	return &IATA{
		IAID:    iaid,
		Options: options,
	}
}

// MarshalBinary allocates a byte slice containing the data from a IATA.
func (i *IATA) MarshalBinary() ([]byte, error) {
	// 4 bytes: IAID
	// N bytes: options slice byte count
	b := buffer.New(nil)

	b.WriteBytes(i.IAID[:])
	opts, err := i.Options.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b.WriteBytes(opts)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a IATA.
//
// If the byte slice does not contain enough data to form a valid IATA,
// io.ErrUnexpectedEOF is returned.
func (i *IATA) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// IATA must contain at least an IAID.
	if b.Len() < 4 {
		return io.ErrUnexpectedEOF
	}

	b.ReadBytes(i.IAID[:])
	return (&i.Options).UnmarshalBinary(b.Remaining())
}
