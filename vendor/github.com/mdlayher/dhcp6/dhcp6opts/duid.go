package dhcp6opts

import (
	"encoding"
	"errors"
	"io"
	"net"
	"time"

	"github.com/mdlayher/dhcp6/internal/buffer"
)

var (
	// errInvalidDUIDLLT is returned when not enough bytes are present
	// to parse a valid DUIDLLT from a byte slice, or when the DUID type
	// found in the byte slice is incorrect.
	errInvalidDUIDLLT = errors.New("invalid DUID-LLT")

	// errInvalidDUIDEN is returned when not enough bytes are present
	// to parse a valid DUIDEN from a byte slice, or when the DUID type
	// found in the byte slice is incorrect.
	errInvalidDUIDEN = errors.New("invalid DUID-EN")

	// errInvalidDUIDLL is returned when not enough bytes are present
	// to parse a valid DUIDLL from a byte slice, or when the DUID type
	// found in the byte slice is incorrect.
	errInvalidDUIDLL = errors.New("invalid DUID-LL")

	// errInvalidDUIDUUID is returned when not enough bytes are present
	// to parse a valid DUIDUUID from a byte slice, or when the DUID type
	// found in the byte slice is incorrect.
	errInvalidDUIDUUID = errors.New("invalid DUID-UUID")

	// errUnknownDUID is returned when an unknown DUID type is
	// encountered, and thus, a DUID cannot be parsed.
	errUnknownDUID = errors.New("unknown DUID type")
)

var (
	// duidLLTTime is the date specified in RFC 3315, Section 9.2, for use
	// with DUID-LLT generation.  It is used to calculate a duration from an
	// input time after this date.  Dates before this time are not valid for
	// creation of DUIDLLT values.
	duidLLTTime = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
)

// DUIDType is a type of DHCP Unique Identifier, as defined in RFC
// 3315, Section 9.  DUIDs are used to uniquely identify a client to a
// server, or vice-versa.
type DUIDType uint16

// DUIDType constants which indicate DUID types described in RFCs 3315 and 6355.
//
// These DUID types are taken from IANA's DHCPv6 parameters registry:
// http://www.iana.org/assignments/dhcpv6-parameters/dhcpv6-parameters.xhtml.
const (
	// RFC 3315
	DUIDTypeLLT DUIDType = 1
	DUIDTypeEN  DUIDType = 2
	DUIDTypeLL  DUIDType = 3

	// RFC 6355
	DUIDTypeUUID DUIDType = 4
)

// DUID represents a DHCP Unique Identifier, as defined in RFC
// 3315, Section 9.  A DUID is used by a DHCP server to identify
// unique clients.  A DUID can also be used by a DHCP client to identify
// a unique server, when needed.
//
// The DUID interface represents a generic DUID, but DUIDs can be
// type-asserted to one of four specific types outlined in RFC 3315
// and RFC 6355:
//   -  DUIDLLT - DUID Based on Link-layer Address Plus Time
//   -   DUIDEN - DUID Assigned by Vendor Based on Enterprise Number
//   -   DUIDLL - DUID Based on Link-layer Address
//   - DUIDUUID - DUID Based on Universally Unique Identifier
//
// If further introspection of the DUID is needed, a type switch is
// recommended:
//	switch d := duid.(type) {
//	case *dhcp6.DUIDLLT:
//		fmt.Println(d.Time)
//	case *dhcp6.DUIDEN:
//		fmt.Println(d.EnterpriseNumber)
//	case *dhcp6.DUIDLL:
//		fmt.Println(d.HardwareAddr)
//	case *dhcp6.DUIDUUID:
//		fmt.Println(d.UUID)
//	}
type DUID interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// DUIDLLT represents a DUID Based on Link-layer Address Plus Time [DUID-LLT],
// as defined in RFC 3315, Section 9.2.
//
// This DUID type must only be used with clients and servers with stable,
// persistent storage.  It is the recommended DUID type for all general
// purpose computing devices.
type DUIDLLT struct {
	// Type specifies the DUID type.  For a DUIDLLT, this should always be
	// DUIDTypeLLT.
	Type DUIDType

	// HardwareType specifies an IANA-assigned hardware type, as described
	// in RFC 826.
	HardwareType uint16

	// Time specifies the duration of the time this DUID was generated, minus
	// midnight (UTC), January 1, 2000.
	Time time.Duration

	// HardwareAddr specifies the hardware address for an arbitrary link-layer
	// interface on a device, used in generating the DUIDLLT.  This value
	// could represent any arbitrary interface on a system, and should not be
	// treated as a client or server's communicating hardware address.
	HardwareAddr net.HardwareAddr
}

// NewDUIDLLT generates a new DUIDLLT from an input IANA-assigned hardware
// type, time value, and a hardware address.
//
// The time value must be greater than midnight (UTC), January 1, 2000.
func NewDUIDLLT(hardwareType uint16, time time.Time, hardwareAddr net.HardwareAddr) (*DUIDLLT, error) {
	// Do not accept dates before duidLLTTime.
	if time.Before(duidLLTTime) {
		return nil, ErrInvalidDUIDLLTTime
	}

	return &DUIDLLT{
		Type:         DUIDTypeLLT,
		HardwareType: hardwareType,
		Time:         time.Sub(duidLLTTime),
		HardwareAddr: hardwareAddr,
	}, nil
}

// MarshalBinary allocates a byte slice containing the data from a DUIDLLT.
func (d *DUIDLLT) MarshalBinary() ([]byte, error) {
	// 2 bytes: DUID type
	// 2 bytes: hardware type
	// 4 bytes: time duration
	// N bytes: hardware address
	b := buffer.New(nil)

	b.Write16(uint16(d.Type))
	b.Write16(d.HardwareType)
	b.Write32(uint32(d.Time / time.Second))
	b.WriteBytes(d.HardwareAddr)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a DUIDLLT.
// If the byte slice does not contain enough data to form a valid
// DUIDLLT, or another DUID type is indicated, errInvalidDUIDLLT is returned.
func (d *DUIDLLT) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Too short to be valid DUIDLLT
	if b.Len() < 8 {
		return io.ErrUnexpectedEOF
	}

	// Verify DUID type
	dType := DUIDType(b.Read16())
	if dType != DUIDTypeLLT {
		return errInvalidDUIDLLT
	}
	d.Type = dType
	d.HardwareType = b.Read16()
	d.Time = time.Duration(b.Read32()) * time.Second

	d.HardwareAddr = b.Remaining()

	return nil
}

// DUIDEN represents a DUID Assigned by Vendor Based on Enterprise Number
// [DUID-EN], as defined in RFC 3315, Section 9.3.  This DUID type
// uses an IANA-assigned Private Enterprise Number for a given vendor.
type DUIDEN struct {
	// Type specifies the DUID type.  For a DUIDEN, this should always be
	// DUIDTypeEN.
	Type DUIDType

	// EnterpriseNumber specifies an IANA-assigned vendor Private Enterprise
	// Number.
	EnterpriseNumber uint32

	// Identifier specifies a unique identifier of arbitrary length.  This
	// value is typically assigned when a device is manufactured.
	Identifier []byte
}

// NewDUIDEN generates a new DUIDEN from an input IANA-assigned Private
// Enterprise Number and a variable length unique identifier byte slice.
func NewDUIDEN(enterpriseNumber uint32, identifier []byte) *DUIDEN {
	return &DUIDEN{
		Type:             DUIDTypeEN,
		EnterpriseNumber: enterpriseNumber,
		Identifier:       identifier,
	}
}

// MarshalBinary allocates a byte slice containing the data from a DUIDEN.
func (d *DUIDEN) MarshalBinary() ([]byte, error) {
	// 2 bytes: DUID type
	// 4 bytes: enterprise number
	// N bytes: identifier
	b := buffer.New(nil)

	b.Write16(uint16(d.Type))
	b.Write32(d.EnterpriseNumber)
	b.WriteBytes(d.Identifier)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a DUIDEN.
// If the byte slice does not contain enough data to form a valid
// DUIDEN, or another DUID type is indicated, errInvalidDUIDEN is returned.
func (d *DUIDEN) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Too short to be valid DUIDEN
	if b.Len() < 6 {
		return io.ErrUnexpectedEOF
	}

	// Verify DUID type
	dType := DUIDType(b.Read16())
	if dType != DUIDTypeEN {
		return errInvalidDUIDEN
	}
	d.Type = dType
	d.EnterpriseNumber = b.Read32()
	d.Identifier = b.Remaining()
	return nil
}

// DUIDLL represents a DUID Based on Link-layer Address [DUID-LL],
// as defined in RFC 3315, Section 9.4.
//
// This DUID type is recommended for devices with a
// permanently-connected network interface, but without stable,
// persistent storage.
//
// DUIDLL values are generated automatically for Servers which are not
// created with a ServerID, using the hardware type found by HardwareType
// and the hardware address of the listening network interface.
type DUIDLL struct {
	// Type specifies the DUID type.  For a DUIDLL, this should always be
	// DUIDTypeLL.
	Type DUIDType

	// HardwareType specifies an IANA-assigned hardware type, as described
	// in RFC 826.
	HardwareType uint16

	// HardwareAddr specifies the hardware address for an arbitrary link-layer
	// interface on a device, used in generating the DUIDLL.  This value
	// could represent any arbitrary interface on a system, and should not be
	// treated as a client or server's communicating hardware address.
	HardwareAddr net.HardwareAddr
}

// NewDUIDLL generates a new DUIDLL from an input IANA-assigned hardware
// type and a hardware address.
func NewDUIDLL(hardwareType uint16, hardwareAddr net.HardwareAddr) *DUIDLL {
	return &DUIDLL{
		Type:         DUIDTypeLL,
		HardwareType: hardwareType,
		HardwareAddr: hardwareAddr,
	}
}

// MarshalBinary allocates a byte slice containing the data from a DUIDLL.
func (d *DUIDLL) MarshalBinary() ([]byte, error) {
	// 2 bytes: DUID type
	// 2 bytes: hardware type
	// N bytes: hardware address
	b := buffer.New(nil)

	b.Write16(uint16(d.Type))
	b.Write16(d.HardwareType)
	b.WriteBytes(d.HardwareAddr)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a DUIDLL.
// If the byte slice does not contain enough data to form a valid
// DUIDLL, or another DUID type is indicated, errInvalidDUIDLL is returned.
func (d *DUIDLL) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Too short to be DUIDLL
	if b.Len() < 4 {
		return io.ErrUnexpectedEOF
	}

	// Verify DUID type
	dType := DUIDType(b.Read16())
	if dType != DUIDTypeLL {
		return errInvalidDUIDLL
	}
	d.Type = dType
	d.HardwareType = b.Read16()
	d.HardwareAddr = b.Remaining()

	return nil
}

// DUIDUUID represents a DUID based on Universally Unique Identifier
// [DUID-UUID], as defined in RFC 6355.  This DUID type uses a UUID to
// identify clients or servers.
type DUIDUUID struct {
	// Type specifies the DUID type.  For a DUIDUUID, this should always be
	// DUIDTypeUUID.
	Type DUIDType

	// UUID specifies a Universally Unique Identifier, as described in RFC 4578.
	UUID [16]byte
}

// NewDUIDUUID generates a new DUIDUUID using an input UUID.
func NewDUIDUUID(uuid [16]byte) *DUIDUUID {
	return &DUIDUUID{
		Type: DUIDTypeUUID,
		UUID: uuid,
	}
}

// MarshalBinary allocates a byte slice containing the data from a DUIDUUID.
func (d *DUIDUUID) MarshalBinary() ([]byte, error) {
	//  2 bytes: DUID type
	// 16 bytes: UUID
	b := buffer.New(nil)

	b.Write16(uint16(d.Type))
	b.WriteBytes(d.UUID[:])

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a DUIDUUID.
// If the byte slice does not contain the exact number of bytes
// needed to form a valid DUIDUUID, or another DUID type is indicated,
// errInvalidDUIDUUID is returned.
func (d *DUIDUUID) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// DUIDUUIDs are fixed-length structures
	if b.Len() != 18 {
		return io.ErrUnexpectedEOF
	}

	// Verify DUID type
	dType := DUIDType(b.Read16())
	if dType != DUIDTypeUUID {
		return errInvalidDUIDUUID
	}
	d.Type = dType
	b.ReadBytes(d.UUID[:])
	return nil
}

// parseDUID returns the correct DUID type of the input byte slice as a
// DUID interface type.
func parseDUID(p []byte) (DUID, error) {
	b := buffer.New(p)
	// DUID must have enough bytes to determine its type
	if b.Len() < 2 {
		return nil, io.ErrUnexpectedEOF
	}

	var d DUID
	switch DUIDType(b.Read16()) {
	case DUIDTypeLLT:
		d = new(DUIDLLT)
	case DUIDTypeEN:
		d = new(DUIDEN)
	case DUIDTypeLL:
		d = new(DUIDLL)
	case DUIDTypeUUID:
		d = new(DUIDUUID)
	default:
		return nil, errUnknownDUID
	}

	return d, d.UnmarshalBinary(p)
}
