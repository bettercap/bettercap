package dhcp6opts

import (
	"io"
	"net"
	"time"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// IAPrefix represents an Identity Association Prefix, as defined in RFC 3633,
// Section 10.
//
// Routers may use identity association prefixes (IAPrefixes) to request IPv6
// prefixes to assign individual address to IPv6 clients, using the lifetimes
// specified in the preferred lifetime and valid lifetime fields.  Multiple
// IAPrefixes may be present in a single DHCP request, but only enscapsulated
// within an IAPD's options.
type IAPrefix struct {
	// PreferredLifetime specifies the preferred lifetime of an IPv6 prefix.
	// When the preferred lifetime of a prefix expires, the prefix becomes
	// deprecated, and addresses from the prefix should not be used in new
	// communications.
	//
	// The preferred lifetime of a prefix must not be greater than its valid
	// lifetime.
	PreferredLifetime time.Duration

	// ValidLifetime specifies the valid lifetime of an IPv6 prefix.  When the
	// valid lifetime of a prefix expires, addresses from the prefix the address
	// should not be used for any further communication.
	//
	// The valid lifetime of a prefix must be greater than its preferred
	// lifetime.
	ValidLifetime time.Duration

	// PrefixLength specifies the length in bits of an IPv6 address prefix, such
	// as 32, 64, etc.
	PrefixLength uint8

	// Prefix specifies the IPv6 address prefix from which IPv6 addresses can
	// be allocated.
	Prefix net.IP

	// Options specifies a map of DHCP options specific to this IAPrefix.
	// Its methods can be used to retrieve data from an incoming IAPrefix, or
	// send data with an outgoing IAPrefix.
	Options dhcp6.Options
}

// NewIAPrefix creates a new IAPrefix from preferred and valid lifetime
// durations, an IPv6 prefix length, an IPv6 prefix, and an optional Options
// map.
//
// The preferred lifetime duration must be less than the valid lifetime
// duration.  The IPv6 prefix must be exactly 16 bytes, the correct length
// for an IPv6 address.  Failure to meet either of these conditions will result
// in an error.  If an Options map is not specified, a new one will be
// allocated.
func NewIAPrefix(preferred time.Duration, valid time.Duration, prefixLength uint8, prefix net.IP, options dhcp6.Options) (*IAPrefix, error) {
	// Preferred lifetime must always be less than valid lifetime.
	if preferred > valid {
		return nil, ErrInvalidLifetimes
	}

	// From documentation: If ip is not an IPv4 address, To4 returns nil.
	if prefix.To4() != nil {
		return nil, ErrInvalidIP
	}

	// If no options set, make empty map
	if options == nil {
		options = make(dhcp6.Options)
	}

	return &IAPrefix{
		PreferredLifetime: preferred,
		ValidLifetime:     valid,
		PrefixLength:      prefixLength,
		Prefix:            prefix,
		Options:           options,
	}, nil
}

// MarshalBinary allocates a byte slice containing the data from a IAPrefix.
func (i *IAPrefix) MarshalBinary() ([]byte, error) {
	//  4 bytes: preferred lifetime
	//  4 bytes: valid lifetime
	//  1 byte : prefix length
	// 16 bytes: IPv6 prefix
	//  N bytes: options
	b := buffer.New(nil)

	b.Write32(uint32(i.PreferredLifetime / time.Second))
	b.Write32(uint32(i.ValidLifetime / time.Second))
	b.Write8(i.PrefixLength)
	copy(b.WriteN(net.IPv6len), i.Prefix)
	opts, err := i.Options.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b.WriteBytes(opts)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a IAPrefix.
//
// If the byte slice does not contain enough data to form a valid IAPrefix,
// io.ErrUnexpectedEOF is returned.  If the preferred lifetime value in the
// byte slice is less than the valid lifetime, ErrInvalidLifetimes is
// returned.
func (i *IAPrefix) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// IAPrefix must at least contain lifetimes, prefix length, and prefix
	if b.Len() < 25 {
		return io.ErrUnexpectedEOF
	}

	i.PreferredLifetime = time.Duration(b.Read32()) * time.Second
	i.ValidLifetime = time.Duration(b.Read32()) * time.Second

	// Preferred lifetime must always be less than valid lifetime.
	if i.PreferredLifetime > i.ValidLifetime {
		return ErrInvalidLifetimes
	}

	i.PrefixLength = b.Read8()
	i.Prefix = make(net.IP, net.IPv6len)
	copy(i.Prefix, b.Consume(net.IPv6len))

	return (&i.Options).UnmarshalBinary(b.Remaining())
}
