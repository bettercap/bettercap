package dhcp6opts

import (
	"io"

	"github.com/mdlayher/dhcp6"
	"github.com/mdlayher/dhcp6/internal/buffer"
)

// A VendorOpts is used by clients and servers to exchange
// VendorOpts information.
type VendorOpts struct {
	// EnterpriseNumber specifies an IANA-assigned vendor Private Enterprise
	// Number.
	EnterpriseNumber uint32

	// An opaque object of option-len octets,
	// interpreted by vendor-specific code on the
	// clients and servers
	Options dhcp6.Options
}

// MarshalBinary allocates a byte slice containing the data from a VendorOpts.
func (v *VendorOpts) MarshalBinary() ([]byte, error) {
	// 4 bytes: EnterpriseNumber
	// N bytes: options slice byte count
	b := buffer.New(nil)
	b.Write32(v.EnterpriseNumber)
	opts, err := v.Options.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b.WriteBytes(opts)

	return b.Data(), nil
}

// UnmarshalBinary unmarshals a raw byte slice into a VendorOpts.
// If the byte slice does not contain enough data to form a valid
// VendorOpts, io.ErrUnexpectedEOF is returned.
// If option-data are invalid, then ErrInvalidPacket is returned.
func (v *VendorOpts) UnmarshalBinary(p []byte) error {
	b := buffer.New(p)
	// Too short to be valid VendorOpts
	if b.Len() < 4 {
		return io.ErrUnexpectedEOF
	}

	v.EnterpriseNumber = b.Read32()
	if err := (&v.Options).UnmarshalBinary(b.Remaining()); err != nil {
		// Invalid options means an invalid RelayMessage
		return dhcp6.ErrInvalidPacket
	}
	return nil
}
