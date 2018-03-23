package dhcp6opts

import (
	"io"
	"math"
	"net"
	"net/url"
	"time"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// A OptionRequestOption is a list OptionCode, as defined in RFC 3315, Section 22.7.
//
// The Option Request option is used to identify a list of options in a
// message between a client and a server.
type OptionRequestOption []dhcp6.OptionCode

// MarshalBinary allocates a byte slice containing the data from a OptionRequestOption.
func (oro OptionRequestOption) MarshalBinary() ([]byte, error) {
	b := buffer.New(nil)
	for _, opt := range oro {
		b.Write16(uint16(opt))
	}
	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a OptionRequestOption.
//
// If the length of byte slice is not be be divisible by 2,
// errInvalidOptionRequest is returned.
func (oro *OptionRequestOption) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Length must be divisible by 2.
	if b.Len()%2 != 0 {
		return io.ErrUnexpectedEOF
	}

	// Fill slice by parsing every two bytes using index i.
	*oro = make(OptionRequestOption, 0, b.Len()/2)
	for b.Len() > 1 {
		*oro = append(*oro, dhcp6.OptionCode(b.Read16()))
	}
	return nil
}

// A Preference is a preference value, as defined in RFC 3315, Section 22.8.
//
// A preference value is sent by a server to a client to affect the selection
// of a server by the client.
type Preference uint8

// MarshalBinary allocates a byte slice containing the data from a Preference.
func (p Preference) MarshalBinary() ([]byte, error) {
	return []byte{byte(p)}, nil
}

// UnmarshalBinary unmarshals a raw byte slice into a Preference.
//
// If the byte slice is not exactly 1 byte in length, io.ErrUnexpectedEOF is
// returned.
func (p *Preference) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return io.ErrUnexpectedEOF
	}

	*p = Preference(b[0])
	return nil
}

// An ElapsedTime is a client's elapsed request time value, as defined in RFC
// 3315, Section 22.9.
//
// The duration returned reports the time elapsed during a DHCP transaction,
// as reported by a client.
type ElapsedTime time.Duration

// MarshalBinary allocates a byte slice containing the data from an
// ElapsedTime.
func (t ElapsedTime) MarshalBinary() ([]byte, error) {
	b := buffer.New(nil)

	unit := 10 * time.Millisecond
	// The elapsed time value is an unsigned, 16 bit integer.
	// The client uses the value 0xffff to represent any
	// elapsed time values greater than the largest time value
	// that can be represented in the Elapsed Time option.
	if max := time.Duration(math.MaxUint16) * unit; time.Duration(t) > max {
		t = ElapsedTime(max)
	}
	b.Write16(uint16(time.Duration(t) / unit))
	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a ElapsedTime.
//
// If the byte slice is not exactly 2 bytes in length, io.ErrUnexpectedEOF is
// returned.
func (t *ElapsedTime) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	if b.Len() != 2 {
		return io.ErrUnexpectedEOF
	}

	// Time is reported in hundredths of seconds, so we convert it to a more
	// manageable milliseconds
	*t = ElapsedTime(time.Duration(b.Read16()) * 10 * time.Millisecond)
	return nil
}

// An IP is an IPv6 address.  The IP type is provided for convenience.
// It can be used to easily add IPv6 addresses to an Options map.
type IP net.IP

// MarshalBinary allocates a byte slice containing the data from a IP.
func (i IP) MarshalBinary() ([]byte, error) {
	ip := make([]byte, net.IPv6len)
	copy(ip, i)
	return ip, nil
}

// UnmarshalBinary unmarshals a raw byte slice into an IP.
//
// If the byte slice is not an IPv6 address, io.ErrUnexpectedEOF is
// returned.
func (i *IP) UnmarshalBinary(b []byte) error {
	if len(b) != net.IPv6len {
		return io.ErrUnexpectedEOF
	}

	if ip := net.IP(b); ip.To4() != nil {
		return io.ErrUnexpectedEOF
	}

	*i = make(IP, net.IPv6len)
	copy(*i, b)
	return nil
}

// IPs represents a list of IPv6 addresses.
type IPs []net.IP

// MarshalBinary allocates a byte slice containing the consecutive data of all
// IPs.
func (i IPs) MarshalBinary() ([]byte, error) {
	ips := make([]byte, 0, len(i)*net.IPv6len)
	for _, ip := range i {
		ips = append(ips, ip.To16()...)
	}
	return ips, nil
}

// UnmarshalBinary unmarshals a raw byte slice into a list of IPs.
//
// If the byte slice contains any non-IPv6 addresses, io.ErrUnexpectedEOF is
// returned.
func (i *IPs) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	if b.Len()%net.IPv6len != 0 || b.Len() == 0 {
		return io.ErrUnexpectedEOF
	}

	*i = make(IPs, 0, b.Len()/net.IPv6len)
	for b.Len() > 0 {
		ip := make(net.IP, net.IPv6len)
		b.ReadBytes(ip)
		*i = append(*i, ip)
	}
	return nil
}

// Data is a raw collection of byte slices, typically carrying user class
// data, vendor class data, or PXE boot file parameters.
type Data [][]byte

// MarshalBinary allocates a byte slice containing the data from a Data
// structure.
func (d Data) MarshalBinary() ([]byte, error) {
	// Count number of bytes needed to allocate at once
	var c int
	for _, dd := range d {
		c += 2 + len(dd)
	}

	b := buffer.New(nil)
	d.Marshal(b)
	return b.Data(), nil
}

// Marshal marshals to a given buffer from a Data structure.
func (d Data) Marshal(b *buffer.Buffer) {
	for _, dd := range d {
		// 2 byte: length of data
		b.Write16(uint16(len(dd)))

		// N bytes: actual raw data
		b.WriteBytes(dd)
	}
}

// UnmarshalBinary unmarshals a raw byte slice into a Data structure.
func (d *Data) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	return d.Unmarshal(b)
}

// Unmarshal marshals from a given buffer into a Data structure.
// Data is packed in the form:
//   - 2 bytes: data length
//   - N bytes: raw data
func (d *Data) Unmarshal(b *buffer.Buffer) error {
	data := make(Data, 0, b.Len())

	// Iterate until not enough bytes remain to parse another length value
	for b.Len() > 1 {
		// 2 bytes: length of data.
		length := int(b.Read16())

		// N bytes: actual data.
		data = append(data, b.Consume(length))
	}

	// At least one instance of class data must be present
	if len(data) == 0 {
		return io.ErrUnexpectedEOF
	}

	// If we encounter any trailing bytes, report an error
	if b.Len() != 0 {
		return io.ErrUnexpectedEOF
	}

	*d = data
	return nil
}

// A URL is a uniform resource locater.  The URL type is provided for
// convenience. It can be used to easily add URLs to an Options map.
type URL url.URL

// MarshalBinary allocates a byte slice containing the data from a URL.
func (u URL) MarshalBinary() ([]byte, error) {
	uu := url.URL(u)
	return []byte(uu.String()), nil
}

// UnmarshalBinary unmarshals a raw byte slice into an URL.
//
// If the byte slice is not an URLv6 address, io.ErrUnexpectedEOF is
// returned.
func (u *URL) UnmarshalBinary(b []byte) error {
	uu, err := url.Parse(string(b))
	if err != nil {
		return err
	}

	*u = URL(*uu)
	return nil
}

// ArchTypes is a slice of ArchType values.  It is provided for convenient
// marshaling and unmarshaling of a slice of ArchType values from an Options
// map.
type ArchTypes []ArchType

// MarshalBinary allocates a byte slice containing the data from ArchTypes.
func (a ArchTypes) MarshalBinary() ([]byte, error) {
	b := buffer.New(nil)
	for _, aType := range a {
		b.Write16(uint16(aType))
	}

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into an ArchTypes slice.
//
// If the byte slice is less than 2 bytes in length, or is not a length that
// is divisible by 2, io.ErrUnexpectedEOF is returned.
func (a *ArchTypes) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Length must be at least 2, and divisible by 2.
	if b.Len() < 2 || b.Len()%2 != 0 {
		return io.ErrUnexpectedEOF
	}

	// Allocate ArchTypes at once and unpack every two bytes into an element
	arch := make(ArchTypes, 0, b.Len()/2)
	for b.Len() > 1 {
		arch = append(arch, ArchType(b.Read16()))
	}

	*a = arch
	return nil
}

// A NII is a Client Network Interface Identifier, as defined in RFC 5970,
// Section 3.4.
//
// A NII is used to indicate a client's level of Universal Network Device
// Interface (UNDI) support.
type NII struct {
	// Type specifies a network interface type.
	Type uint8

	// Major specifies the UNDI major revisision which this client supports.
	Major uint8

	// Minor specifies the UNDI minor revision which this client supports.
	Minor uint8
}

// MarshalBinary allocates a byte slice containing the data from a NII.
func (n *NII) MarshalBinary() ([]byte, error) {
	b := make([]byte, 3)

	b[0] = n.Type
	b[1] = n.Major
	b[2] = n.Minor

	return b, nil
}

// UnmarshalBinary unmarshals a raw byte slice into a NII.
//
// If the byte slice is not exactly 3 bytes in length, io.ErrUnexpectedEOF
// is returned.
func (n *NII) UnmarshalBinary(b []byte) error {
	// Length must be exactly 3
	if len(b) != 3 {
		return io.ErrUnexpectedEOF
	}

	n.Type = b[0]
	n.Major = b[1]
	n.Minor = b[2]

	return nil
}

// A RelayMessageOption is used by a DHCPv6 Relay Agent to relay messages
// between clients and servers or other relay agents through Relay-Forward
// and Relay-Reply message types. The original client DHCP message (i.e.,
// the packet payload,  excluding UDP and IP headers) is encapsulated in a
// Relay Message option.
type RelayMessageOption []byte

// MarshalBinary allocates a byte slice containing the data from a RelayMessageOption.
func (r *RelayMessageOption) MarshalBinary() ([]byte, error) {
	return *r, nil
}

// UnmarshalBinary unmarshals a raw byte slice into a RelayMessageOption.
func (r *RelayMessageOption) UnmarshalBinary(b []byte) error {
	*r = make([]byte, len(b))
	copy(*r, b)
	return nil
}

// SetClientServerMessage sets a Packet (e.g. Solicit, Advertise ...) into this option.
func (r *RelayMessageOption) SetClientServerMessage(p *dhcp6.Packet) error {
	b, err := p.MarshalBinary()
	if err != nil {
		return err
	}

	*r = b
	return nil
}

// SetRelayMessage sets a RelayMessage (e.g. Relay Forward, Relay Reply) into this option.
func (r *RelayMessageOption) SetRelayMessage(p *RelayMessage) error {
	b, err := p.MarshalBinary()
	if err != nil {
		return err
	}

	*r = b
	return nil
}

// ClientServerMessage gets the client server message (e.g. Solicit,
// Advertise ...) into this option (when hopcount = 0 of outer RelayMessage).
func (r *RelayMessageOption) ClientServerMessage() (*dhcp6.Packet, error) {
	p := new(dhcp6.Packet)
	err := p.UnmarshalBinary(*r)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// RelayMessage gets the relay message (e.g. Relay Forward, Relay Reply) into
// this option (when hopcount > 0 of outer RelayMessage).
func (r *RelayMessageOption) RelayMessage() (*RelayMessage, error) {
	rm := new(RelayMessage)
	err := rm.UnmarshalBinary(*r)
	if err != nil {
		return nil, err
	}

	return rm, nil
}

// An InterfaceID is an opaque value of arbitrary length generated
// by the relay agent to identify one of the
// relay agent's interfaces.
type InterfaceID []byte

// MarshalBinary allocates a byte slice containing the data from a InterfaceID.
func (i *InterfaceID) MarshalBinary() ([]byte, error) {
	return *i, nil
}

// UnmarshalBinary unmarshals a raw byte slice into a InterfaceID.
func (i *InterfaceID) UnmarshalBinary(b []byte) error {
	*i = make([]byte, len(b))
	copy(*i, b)
	return nil
}

// A BootFileParam are boot file parameters.
type BootFileParam []string

// MarshalBinary allocates a byte slice containing the data from a
// BootFileParam.
func (bfp BootFileParam) MarshalBinary() ([]byte, error) {
	// Convert []string to [][]byte.
	bb := make(Data, 0, len(bfp))
	for _, param := range bfp {
		bb = append(bb, []byte(param))
	}
	return bb.MarshalBinary()
}

// UnmarshalBinary unmarshals a raw byte slice into a BootFileParam.
func (bfp *BootFileParam) UnmarshalBinary(b []byte) error {
	var d Data
	if err := (&d).UnmarshalBinary(b); err != nil {
		return err
	}
	// Convert [][]byte to []string.
	*bfp = make([]string, 0, len(d))
	for _, param := range d {
		*bfp = append(*bfp, string(param))
	}
	return nil
}
