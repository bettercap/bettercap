package dhcp6opts

import (
	"io"
	"net"
	"time"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// IAAddr represents an Identity Association Address, as defined in RFC 3315,
// Section 22.6.
//
// DHCP clients use identity association addresses (IAAddrs) to request IPv6
// addresses from a DHCP server, using the lifetimes specified in the preferred
// lifetime and valid lifetime fields.  Multiple IAAddrs may be present in a
// single DHCP request, but only enscapsulated within an IANA or IATA options
// field.
type IAAddr struct {
	// IP specifies the IPv6 address to offer to a client.  The validity of the
	// address is controlled by the PreferredLifetime and ValidLifetime fields.
	IP net.IP

	// PreferredLifetime specifies the preferred lifetime of an IPv6 address.
	// When the preferred lifetime of an address expires, the address becomes
	// deprecated, and should not be used in new communications.
	//
	// The preferred lifetime of an address must not be greater than its
	// valid lifetime.
	PreferredLifetime time.Duration

	// ValidLifetime specifies the valid lifetime of an IPv6 address.  When the
	// valid lifetime of an address expires, the address should not be used for
	// any further communication.
	//
	// The valid lifetime of an address must be greater than its preferred
	// lifetime.
	ValidLifetime time.Duration

	// Options specifies a map of DHCP options specific to this IAAddr.
	// Its methods can be used to retrieve data from an incoming IAAddr, or
	// send data with an outgoing IAAddr.
	Options dhcp6.Options
}

// NewIAAddr creates a new IAAddr from an IPv6 address, preferred and valid lifetime
// durations, and an optional Options map.
//
// The IP must be exactly 16 bytes, the correct length for an IPv6 address.
// The preferred lifetime duration must be less than the valid lifetime
// duration.  Failure to meet either of these conditions will result in an error.
// If an Options map is not specified, a new one will be allocated.
func NewIAAddr(ip net.IP, preferred time.Duration, valid time.Duration, options dhcp6.Options) (*IAAddr, error) {
	// From documentation: If ip is not an IPv4 address, To4 returns nil.
	if ip.To4() != nil {
		return nil, ErrInvalidIP
	}

	// Preferred lifetime must always be less than valid lifetime.
	if preferred > valid {
		return nil, ErrInvalidLifetimes
	}

	// If no options set, make empty map
	if options == nil {
		options = make(dhcp6.Options)
	}

	return &IAAddr{
		IP:                ip,
		PreferredLifetime: preferred,
		ValidLifetime:     valid,
		Options:           options,
	}, nil
}

// MarshalBinary allocates a byte slice containing the data from a IAAddr.
func (i *IAAddr) MarshalBinary() ([]byte, error) {
	// 16 bytes: IPv6 address
	//  4 bytes: preferred lifetime
	//  4 bytes: valid lifetime
	//  N bytes: options
	b := buffer.New(nil)

	copy(b.WriteN(net.IPv6len), i.IP)
	b.Write32(uint32(i.PreferredLifetime / time.Second))
	b.Write32(uint32(i.ValidLifetime / time.Second))
	opts, err := i.Options.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b.WriteBytes(opts)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a IAAddr.
//
// If the byte slice does not contain enough data to form a valid IAAddr,
// io.ErrUnexpectedEOF is returned.  If the preferred lifetime value in the
// byte slice is less than the valid lifetime, ErrInvalidLifetimes is returned.
func (i *IAAddr) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	if b.Len() < 24 {
		return io.ErrUnexpectedEOF
	}

	i.IP = make(net.IP, net.IPv6len)
	copy(i.IP, b.Consume(net.IPv6len))

	i.PreferredLifetime = time.Duration(b.Read32()) * time.Second
	i.ValidLifetime = time.Duration(b.Read32()) * time.Second

	// Preferred lifetime must always be less than valid lifetime.
	if i.PreferredLifetime > i.ValidLifetime {
		return ErrInvalidLifetimes
	}

	return (&i.Options).UnmarshalBinary(b.Remaining())
}
