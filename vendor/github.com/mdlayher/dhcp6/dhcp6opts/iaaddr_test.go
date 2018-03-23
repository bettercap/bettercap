package dhcp6opts

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/mdlayher/dhcp6"
)

// TestNewIAAddr verifies that NewIAAddr creates a proper IAAddr value or returns
// a correct error for input values.
func TestNewIAAddr(t *testing.T) {
	var tests = []struct {
		desc      string
		ip        net.IP
		preferred time.Duration
		valid     time.Duration
		options   dhcp6.Options
		iaaddr    *IAAddr
		err       error
	}{
		{
			desc:   "all zero values",
			iaaddr: &IAAddr{},
		},
		{
			desc: "IPv4 address",
			ip:   net.IP([]byte{192, 168, 1, 1}),
			err:  ErrInvalidIP,
		},
		{
			desc:      "preferred greater than valid lifetime",
			ip:        net.IPv6loopback,
			preferred: 2 * time.Second,
			valid:     1 * time.Second,
			err:       ErrInvalidLifetimes,
		},
		{
			desc:      "IPv6 localhost, 1s preferred, 2s valid, no options",
			ip:        net.IPv6loopback,
			preferred: 1 * time.Second,
			valid:     2 * time.Second,
			iaaddr: &IAAddr{
				IP:                net.IPv6loopback,
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
			},
		},
		{
			desc:      "IPv6 localhost, 1s preferred, 2s valid, option client ID [0 1]",
			ip:        net.IPv6loopback,
			preferred: 1 * time.Second,
			valid:     2 * time.Second,
			options: dhcp6.Options{
				dhcp6.OptionClientID: [][]byte{{0, 1}},
			},
			iaaddr: &IAAddr{
				IP:                net.IPv6loopback,
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iaaddr, err := NewIAAddr(tt.ip, tt.preferred, tt.valid, tt.options)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error for NewIAAddr: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		want, err := tt.iaaddr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got, err := iaaddr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected IAAddr bytes:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestIAAddrUnmarshalBinary verifies that IAAddr.UnmarshalBinary produces a
// correct IAAddr value or error for an input buffer.
func TestIAAddrUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		desc   string
		buf    []byte
		iaaddr *IAAddr
		err    error
	}{
		{
			desc: "one byte IAAddr",
			buf:  []byte{0},
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "23 bytes IAAddr",
			buf:  bytes.Repeat([]byte{0}, 23),
			err:  io.ErrUnexpectedEOF,
		},
		{
			desc: "preferred greater than valid lifetime",
			buf: append(net.IPv6zero, []byte{
				0, 0, 0, 2,
				0, 0, 0, 1,
			}...),
			err: ErrInvalidLifetimes,
		},
		{
			desc: "invalid options (length mismatch)",
			buf: append(net.IPv6zero, []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				0, 1, 0, 1,
			}...),
			err: dhcp6.ErrInvalidOptions,
		},
		{
			desc: "IPv6 loopback, 1s preferred, 2s valid, no options",
			buf: append(net.IPv6loopback, []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
			}...),
			iaaddr: &IAAddr{
				IP:                net.IPv6loopback,
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
			},
		},
		{
			desc: "IPv6 loopback, 1s preferred, 2s valid, option client ID [0 1]",
			buf: append(net.IPv6loopback, []byte{
				0, 0, 0, 1,
				0, 0, 0, 2,
				0, 1, 0, 2, 0, 1,
			}...),
			iaaddr: &IAAddr{
				IP:                net.IPv6loopback,
				PreferredLifetime: 1 * time.Second,
				ValidLifetime:     2 * time.Second,
				Options: dhcp6.Options{
					dhcp6.OptionClientID: [][]byte{{0, 1}},
				},
			},
		},
	}

	for i, tt := range tests {
		iaaddr := new(IAAddr)
		if err := iaaddr.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error for parseIAAddr: %v != %v",
					i, tt.desc, want, got)
			}

			continue
		}

		want, err := tt.iaaddr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got, err := iaaddr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] test %q, unexpected IAAddr bytes for parseIAAddr:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}

		for _, v := range iaaddr.Options {
			for ii := range v {
				if want, got := cap(v[ii]), cap(v[ii]); want != got {
					t.Fatalf("[%02d] test %q, unexpected capacity option data:\n- want: %v\n-  got: %v",
						i, tt.desc, want, got)
				}
			}
		}
	}
}
