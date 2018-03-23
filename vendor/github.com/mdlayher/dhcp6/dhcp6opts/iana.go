package dhcp6opts

import (
	"io"
	"time"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// IANA represents an Identity Association for Non-temporary Addresses, as
// defined in RFC 3315, Section 22.4.
//
// Multiple IANAs may be present in a single DHCP request.
type IANA struct {
	// IAID specifies a DHCP identity association identifier.  The IAID
	// is a unique, client-generated identifier.
	IAID [4]byte

	// T1 specifies how long a DHCP client will wait to contact this server,
	// to extend the lifetimes of the addresses assigned to this IANA
	// by this server.
	T1 time.Duration

	// T2 specifies how long a DHCP client will wait to contact any server,
	// to extend the lifetimes of the addresses assigned to this IANA
	// by this server.
	T2 time.Duration

	// Options specifies a map of DHCP options specific to this IANA.
	// Its methods can be used to retrieve data from an incoming IANA, or send
	// data with an outgoing IANA.
	Options dhcp6.Options
}

// NewIANA creates a new IANA from an IAID, T1 and T2 durations, and an
// Options map.  If an Options map is not specified, a new one will be
// allocated.
func NewIANA(iaid [4]byte, t1 time.Duration, t2 time.Duration, options dhcp6.Options) *IANA {
	if options == nil {
		options = make(dhcp6.Options)
	}

	return &IANA{
		IAID:    iaid,
		T1:      t1,
		T2:      t2,
		Options: options,
	}
}

// MarshalBinary allocates a byte slice containing the data from a IANA.
func (i IANA) MarshalBinary() ([]byte, error) {
	// 4 bytes: IAID
	// 4 bytes: T1
	// 4 bytes: T2
	// N bytes: options slice byte count
	b := buffer.New(nil)

	b.WriteBytes(i.IAID[:])
	b.Write32(uint32(i.T1 / time.Second))
	b.Write32(uint32(i.T2 / time.Second))
	opts, err := i.Options.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b.WriteBytes(opts)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a IANA.
//
// If the byte slice does not contain enough data to form a valid IANA,
// io.ErrUnexpectedEOF is returned.
func (i *IANA) UnmarshalBinary(p []byte) error {
	// IANA must contain at least an IAID, T1, and T2.
	b := buffer.New(p)
	if b.Len() < 12 {
		return io.ErrUnexpectedEOF
	}

	b.ReadBytes(i.IAID[:])
	i.T1 = time.Duration(b.Read32()) * time.Second
	i.T2 = time.Duration(b.Read32()) * time.Second

	return (&i.Options).UnmarshalBinary(b.Remaining())
}
