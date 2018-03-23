package dhcp6opts

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/mdlayher/dhcp6"
)

// TestNewStatusCode verifies that NewStatusCode creates a proper StatusCode
// value for the input values.
func TestNewStatusCode(t *testing.T) {
	var tests = []struct {
		status  dhcp6.Status
		message string
		sc      *StatusCode
	}{
		{
			status:  dhcp6.StatusSuccess,
			message: "Success",
			sc: &StatusCode{
				Code:    dhcp6.StatusSuccess,
				Message: "Success",
			},
		},
	}

	for i, tt := range tests {
		if want, got := tt.sc, NewStatusCode(tt.status, tt.message); !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] unexpected StatusCode for NewStatusCode(%v, %q)\n- want: %v\n-  got: %v",
				i, tt.status, tt.message, want, got)
		}
	}
}

// TestStatusCodeUnmarshalBinary verifies that StatusCode.UnmarshalBinary
// returns correct StatusCode and error values for several input values.
func TestStatusCodeUnmarshalBinary(t *testing.T) {
	var tests = []struct {
		buf []byte
		sc  *StatusCode
		err error
	}{
		{
			buf: []byte{0},
			err: io.ErrUnexpectedEOF,
		},
		{
			buf: []byte{0, 0},
			sc: &StatusCode{
				Code: dhcp6.StatusSuccess,
			},
		},
		{
			buf: append([]byte{0, 1}, []byte("deadbeef")...),
			sc: &StatusCode{
				Code:    dhcp6.StatusUnspecFail,
				Message: "deadbeef",
			},
		},
	}

	for i, tt := range tests {
		sc := new(StatusCode)
		if err := sc.UnmarshalBinary(tt.buf); err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] unexpected error for parseStatusCode(%v): %v != %v",
					i, tt.buf, want, got)
			}

			continue
		}

		want, err := tt.sc.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		got, err := sc.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(want, got) {
			t.Fatalf("[%02d] unexpected StatusCode for parseStatusCode(%v)\n- want: %v\n-  got: %v",
				i, tt.buf, want, got)
		}
	}
}
