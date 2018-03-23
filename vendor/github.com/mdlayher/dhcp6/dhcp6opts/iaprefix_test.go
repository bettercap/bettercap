package dhcp6opts

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/mdlayher/dhcp6"
)

// TestNewIAPrefix verifies that NewIAPrefix creates a proper IAPrefix value
// or returns a correct error for input values.
func TestNewIAPrefix(t *testing.T) {
	var tests = []struct {
		desc      string
		preferred time.Duration
		valid     time.Duration
		pLength   uint8
		prefix    net.IP
		options   dhcp6.Options
		iaprefix  *IAPrefix
		err       error
	}{
		{
			desc:     "all zero values",
			iaprefix: &IAPrefix{},
		},
		{
			desc:      "preferred greater than valid lifetime",
			preferred: 2 * time.Second,
			valid:     1 * time.Second,
			err:       ErrInvalidLifetimes,
		},
		{
			desc:   "IPv4 address",
			prefix: net.IP([]byte{192, 168, 1, 1}),
			err:    ErrInvalidIP,
		},
		{
			desc:      "1s preferred, 2s valid, '2001:db8::/32', no options",
			preferred: 1 * time.Second,
			valid:     2 * time.Second,
			pLength:   32,
			prefix:    net.ParseIP("2001:db8::"),
			iaprefix: &IAPrefix{
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
				PrefixLength:      32,
				Prefix:            net.ParseIP("2001:db8::"),
			},
		},
		{
			desc:      "1s preferred, 2s valid, '2001:db8::6:1/64', option client ID [0 1]",
			preferred: 1 * time.Second,
			valid:     2 * time.Second,
			pLength:   64,
			prefix:    net.ParseIP("2001:db8::6:1"),
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{0, 1}},
			},
			iaprefix: &IAPrefix{
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
				PrefixLength:      64,
				Prefix:            net.ParseIP("2001:db8::6:1"),
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iaprefix, err := NewIAPrefix(tt.preferred, tt.valid, tt.pLength, tt.prefix, tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error for NewIAPrefix: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		want, err := tt.iaprefix.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got, err := iaprefix.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected IAPrefix bytes:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestIAPrefixUnmarshalBinary verifies that IAPrefix.UnmarshalBinary produces
// a correct IAPrefix value or error for an input buffer.
func TestIAPrefixUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc     string
		buf      []byte
		iaprefix *IAPrefix
		err      error
	}{
		{
			desc: "one byte IAPrefix",
			buf:  []byte{0},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "24 bytes IAPrefix",
			buf:  bytes.Repeat([]byte{0}, 24),
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "preferred greater than valid lifetime",
			buf: append([]byte{
				0, 0, 0, 2,
				0, 0, 0, 1,
			}, bytes.Repeat([]byte{0}, 17)...),
			err: ErrInvalidLifetimes,
		},
		{
			desc: "invalid options (length mismatch)",
			buf: []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				0,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 1, 0, 1,
			},
			err: dhcp6.ErrInvalidOptions,
		},
		{
			desc: "1s preferred, 2s valid, '2001:db8::/32', no options",
			buf: []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				32,
				32, 1, 13, 184, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0,
			},
			iaprefix: &IAPrefix{
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
				PrefixLength:      32,
				Prefix:            net.ParseIP("2001:db8::"),
			},
		},
		{
			desc: "1s preferred, 2s valid, '2001:db8::6:1/64', option client ID [0 1]",
			buf: []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				64,
				32, 1, 13, 184, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 6, 0, 1,
				0, 1, 0, 2, 0, 1,
			},
			iaprefix: &IAPrefix{
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
				PrefixLength:      64,
				Prefix:            net.ParseIP("2001:db8::6:1"),
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iaprefix := new(IAPrefix)
		if err := iaprefix.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error for parseIAPrefix: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		want, err := tt.iaprefix.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got, err := iaprefix.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected IAPrefix bytes for parseIAPrefix:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}
