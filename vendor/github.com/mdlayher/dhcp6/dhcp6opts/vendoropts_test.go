package dhcp6opts

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/mdlayher/dhcp6"
)

// TestVendorOptsMarshalBinary verifies that VendorOpts marshals properly for the input values.
func TestVendorOptsMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc       string
		buf        []byte
		vendorOpts *VendorOpts
	}{
		{
			desc:       "all zero values",
			buf:        bytes.Repeat([]byte{0}, 4),
			vendorOpts: &VendorOpts{},
		},
		{
			desc: "[0, 0, 5, 0x58] EnterpriseNumber, [1: []byte{3, 4}, 2: []byte{0x04, 0xa3, 0x9e}] vendorOpts",
			buf: []byte{
				0, 0, 5, 0x58,
				0, 1, 0, 2, 3, 4,
				0, 2, 0, 3, 0x04, 0xa3, 0x9e,
			},
			vendorOpts: &VendorOpts{
				EnterpriseNumber: 1368,
				Options: dhcp6.Options{
					1: [][]byte{
						{3, 4},
					},
					2: [][]byte{
						{0x04, 0xa3, 0x9e},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		want := tt.buf
		got, err := tt.vendorOpts.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected VendorOpts bytes for MarshalBinary(%v)\n- want: %v\n-  got: %v",
				i, tt.desc, tt.buf, want, got)
		}
	}
}

// TestVendorOptsUnmarshalBinary verifies that VendorOpts unmarshals properly for the input values.
func TestVendorOptsUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		buf        []byte
		vendorOpts *VendorOpts
		err        error
	}{
		{
			buf: bytes.Repeat([]byte{0}, 3),
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: []byte{
				0, 0, 5, 0x58,
				0, 1, 0, 0xa,
			},
			err: dhcp6.ErrInvalidPacket,
		},
		{
			buf: []byte{
				0, 0, 5, 0x58,
				0, 1, 0, 2, 3, 4,
				0, 2, 0, 3, 0x04, 0xa3, 0x9e,
			},
			vendorOpts: &VendorOpts{
				EnterpriseNumber: 1368,
				Options: dhcp6.Options{
					1: [][]byte{
						{3, 4},
					},
					2: [][]byte{
						{0x04, 0xa3, 0x9e},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		vendorOpts := new(VendorOpts)
		if want, got := tt.err, vendorOpts.UnmarshalBinary(tt.buf); want != got {
			t.Fatalf("[%02d] unexpected error for parseVendorOpts(%v):\n- want: %v\n-  got: %v", i, tt.buf, want, got)
		}

		if tt.err != nil {
			continue
		}

		if want, got := tt.vendorOpts, vendorOpts; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected VendorOpts for parseVendorOpts(%v):\n- want: %v\n-  got: %v", i, tt.buf, want, got)
		}
	}
}
