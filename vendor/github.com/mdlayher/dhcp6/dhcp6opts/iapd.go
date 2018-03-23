package dhcp6opts

import (
	"io"
	"time"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// IAPD represents an Identity Association for Prefix Delegation, as
// defined in RFC 3633, Section 9.
//
// Multiple IAPDs may be present in a single DHCP request.
type IAPD struct {
	// IAID specifies a DHCP identity association identifier.  The IAID
	// is a unique, client-generated identifier.
	IAID [4]byte

	// T1 specifies how long a requesting router will wait to contact a
	// delegating router, to extend the lifetimes of the prefixes delegated
	// to this IAPD, by the delegating router.
	T1 time.Duration

	// T2 specifies how long a requesting router will wait to contact any
	// available delegating router, to extend the lifetimes of the prefixes
	// delegated to this IAPD.
	T2 time.Duration

	// Options specifies a map of DHCP options specific to this IAPD.
	// Its methods can be used to retrieve data from an incoming IAPD, or send
	// data with an outgoing IAPD.
	Options dhcp6.Options
}

// NewIAPD creates a new IAPD from an IAID, T1 and T2 durations, and an
// Options map.  If an Options map is not specified, a new one will be
// allocated.
func NewIAPD(iaid [4]byte, t1 time.Duration, t2 time.Duration, options dhcp6.Options) *IAPD {
	if options == nil {
		options = make(dhcp6.Options)
	}

	return &IAPD{
		IAID:    iaid,
		T1:      t1,
		T2:      t2,
		Options: options,
	}
}

// MarshalBinary allocates a byte slice containing the data from a IAPD.
func (i *IAPD) MarshalBinary() ([]byte, error) {
	// 4 bytes: IAID
	// 4 bytes: T1
	// 4 bytes: T2
	// N bytes: options slice byte count
	buf := buffer.New(nil)

	buf.WriteBytes(i.IAID[:])
	buf.Write32(uint32(i.T1 / time.Second))
	buf.Write32(uint32(i.T2 / time.Second))
	opts, err := i.Options.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.WriteBytes(opts)

	return buf.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a IAPD.
//
// If the byte slice does not contain enough data to form a valid IAPD,
// io.ErrUnexpectedEOF is returned.
func (i *IAPD) UnmarshalBinary(b []byte) error {
	// IAPD must contain at least an IAID, T1, and T2.
	buf := buffer.New(b)
	if buf.Len() < 12 {
		return io.ErrUnexpectedEOF
	}

	copy(i.IAID[:], buf.Consume(4))
	i.T1 = time.Duration(buf.Read32()) * time.Second
	i.T2 = time.Duration(buf.Read32()) * time.Second

	return (&i.Options).UnmarshalBinary(buf.Remaining())
}
