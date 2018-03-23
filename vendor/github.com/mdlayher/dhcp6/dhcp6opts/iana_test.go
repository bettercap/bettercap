package dhcp6opts

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/mdlayher/dhcp6"
)

// TestNewIANA verifies that NewIANA creates a proper IANA value for
// input values.
func TestNewIANA(t *testing.T) {
	var tests = []struct {
		desc    string
		iaid    [4]byte
		t1      time.Duration
		t2      time.Duration
		options dhcp6.Options
		iana    *IANA
	}{
		{
			desc: "all zero values",
			iana: &IANA{},
		},
		{
			desc: "[0 1 2 3] IAID, 30s T1, 60s T2, option client ID [0 1]",
			iaid: [4]byte{0, 1, 2, 3},
			t1:   30 * time.Second,
			t2:   60 * time.Second,
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{0, 1}},
			},
			iana: &IANA{
				IAID: [4]byte{0, 1, 2, 3},
				T1:   30 * time.Second,
				T2:   60 * time.Second,
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iana := NewIANA(tt.iaid, tt.t1, tt.t2, tt.options)

		want, err := tt.iana.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got, err := iana.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected IANA bytes for NewIANA(%v, %v, %v, %v)\n- want: %v\n-  got: %v",
				i, tt.desc, tt.iaid, tt.t1, tt.t2, tt.options, want, got)
		}
	}
}

// TestIANAMarshalBinary verifies that IANA.MarshalBinary allocates and returns a correct
// byte slice for a variety of input data.
func TestIANAMarshalBinary(t *testing.T) {
	var tests = []struct {
		desc string
		iana *IANA
		buf  []byte
	}{
		{
			desc: "empty IANA",
			iana: &IANA{},
			buf: []byte{
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
		},
		{
			desc: "[1 2 3 4] IAID only",
			iana: &IANA{
				IAID: [4]byte{1, 2, 3, 4},
			},
			buf: []byte{
				1, 2, 3, 4,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
		},
		{
			desc: "[1 2 3 4] IAID, 30s T1, 60s T2",
			iana: &IANA{
				IAID: [4]byte{1, 2, 3, 4},
				T1:   30 * time.Second,
				T2:   60 * time.Second,
			},
			buf: []byte{
				1, 2, 3, 4,
				0, 0, 0, 30,
				0, 0, 0, 60,
			},
		},
		{
			desc: "[1 2 3 4] IAID, 30s T1, 60s T2, option client ID [0 1]",
			iana: &IANA{
				IAID: [4]byte{1, 2, 3, 4},
				T1:   30 * time.Second,
				T2:   60 * time.Second,
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
			buf: []byte{
				1, 2, 3, 4,
				0, 0, 0, 30,
				0, 0, 0, 60,
				0, 1, 0, 2, 0, 1,
			},
		},
	}

	for i, tt := range tests {
		got, err := tt.iana.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if want := tt.buf; !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected IANA bytes:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestIANAUnmarshalBinary verifies that IANA.UnmarshalBinary produces a correct
// IANA value or error for an input buffer.
func TestIANAUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		buf     []byte
		iana    *IANA
		options dhcp6.Options
		err     error
	}{
		{
			buf: []byte{0},
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: bytes.Repeat([]byte{0}, 11),
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: []byte{
				1, 2, 3, 4,
				0, 0, 1, 0,
				0, 0, 2, 0,
				0, 1, 0, 1,
			},
			err: dhcp6.ErrInvalidOptions,
		},
		{
			buf: []byte{
				1, 2, 3, 4,
				0, 0, 1, 0,
				0, 0, 2, 0,
				0, 1, 0, 2, 0, 1,
			},
			iana: &IANA{
				IAID: [4]byte{1, 2, 3, 4},
				T1:   (4 * time.Minute) + 16*time.Second,
				T2:   (8 * time.Minute) + 32*time.Second,
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iana := new(IANA)
		if err := iana.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] unexpected error for parseIANA(%v): %v != %v",
					i, tt.buf, want, got)
			}

			continue
		}

		if want, got := tt.iana, iana; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected IANA for parseIANA(%v):\n- want: %v\n-  got: %v",
				i, tt.buf, want, got)
		}
	}
}
