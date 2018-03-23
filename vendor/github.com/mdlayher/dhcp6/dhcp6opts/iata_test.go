package dhcp6opts

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/mdlayher/dhcp6"
)

// TestNewIATA verifies that NewIATA creates a proper IATA value for
// input values.
func TestNewIATA(t *testing.T) {
	var tests = []struct {
		desc    string
		iaid    [4]byte
		options dhcp6.Options
		iata    *IATA
	}{
		{
			desc: "all zero values",
			iata: &IATA{},
		},
		{
			desc: "[0 1 2 3] IAID, option client ID [0 1]",
			iaid: [4]byte{0, 1, 2, 3},
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{0, 1}},
			},
			iata: &IATA{
				IAID: [4]byte{0, 1, 2, 3},
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iata := NewIATA(tt.iaid, tt.options)

		want, err := tt.iata.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got, err := iata.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected IATA bytes for NewIATA(%v, %v)\n- want: %v\n-  got: %v",
				i, tt.desc, tt.iaid, tt.options, want, got)
		}
	}
}

// TestIATAUnmarshalBinary verifies that IATAUnmarshalBinary produces a
// correct IATA value or error for an input buffer.
func TestIATAUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		buf     []byte
		iata    *IATA
		options dhcp6.Options
		err     error
	}{
		{
			buf: []byte{0},
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: bytes.Repeat([]byte{0}, 3),
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: []byte{
				1, 2, 3, 4,
				0, 1, 0, 1,
			},
			err: dhcp6.ErrInvalidOptions,
		},
		{
			buf: []byte{
				1, 2, 3, 4,
				0, 1, 0, 2, 0, 1,
			},
			iata: &IATA{
				IAID: [4]byte{1, 2, 3, 4},
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iata := new(IATA)
		if err := iata.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] unexpected error for parseIATA(%v): %v != %v",
					i, tt.buf, want, got)
			}

			continue
		}

		if want, got := tt.iata, iata; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected IATA for parseIATA(%v):\n- want: %v\n-  got: %v",
				i, tt.buf, want, got)
		}
	}
}
