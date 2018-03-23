package dhcp6

import (
	"bytes"
	"reflect"
	"testing"
)

type option struct {
	code OptionCode
	data []byte
}

// TestOptionsAddRaw verifies that Options.AddRaw correctly creates or appends
// key/value Option pairs to an Options map.
func TestOptionsAddRaw(t *testing.T) {
	var tests = []struct {
		desc    string
		kv      []option
		options Options
	}{
		{
			desc: "one key/value pair",
			kv: []option{
				{
					code: 1,
					data: []byte("foo"),
				},
			},
			options: Options{
				1: [][]byte{[]byte("foo")},
			},
		},
		{
			desc: "two key/value pairs",
			kv: []option{
				{
					code: 1,
					data: []byte("foo"),
				},
				{
					code: 2,
					data: []byte("bar"),
				},
			},
			options: Options{
				1: [][]byte{[]byte("foo")},
				2: [][]byte{[]byte("bar")},
			},
		},
		{
			desc: "three key/value pairs, two with same key",
			kv: []option{
				{
					code: 1,
					data: []byte("foo"),
				},
				{
					code: 1,
					data: []byte("baz"),
				},
				{
					code: 2,
					data: []byte("bar"),
				},
			},
			options: Options{
				1: [][]byte{[]byte("foo"), []byte("baz")},
				2: [][]byte{[]byte("bar")},
			},
		},
	}

	for i, tt := range tests {
		o := make(Options)
		for _, p := range tt.kv {
			o.AddRaw(p.code, p.data)
		}

		if want, got := tt.options, o; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected Options map:\n- want: %v\n-  got: %v",
				i, tt.desc, want, got)
		}
	}
}

// TestOptionsGet verifies that Options.Get correctly selects the first value
// for a given key, if the value is not empty in an Options map.
func TestOptionsGet(t *testing.T) {
	var tests = []struct {
		desc    string
		options Options
		key     OptionCode
		value   []byte
		err     error
	}{
		{
			desc: "nil Options map",
			err:  ErrOptionNotPresent,
		},
		{
			desc:    "empty Options map",
			options: Options{},
			err:     ErrOptionNotPresent,
		},
		{
			desc: "value not present in Options map",
			options: Options{
				2: [][]byte{[]byte("foo")},
			},
			key: 1,
			err: ErrOptionNotPresent,
		},
		{
			desc: "value present in Options map, but zero length value for key",
			options: Options{
				1: [][]byte{},
			},
			key: 1,
		},
		{
			desc: "value present in Options map",
			options: Options{
				1: [][]byte{[]byte("foo")},
			},
			key:   1,
			value: []byte("foo"),
		},
		{
			desc: "value present in Options map, with multiple values",
			options: Options{
				1: [][]byte{[]byte("foo"), []byte("bar")},
			},
			key: 1,
			err: ErrInvalidPacket,
		},
	}

	for i, tt := range tests {
		value, err := tt.options.GetOne(tt.key)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected err for Options.GetOne(%v): %v != %v",
					i, tt.desc, tt.key, want, got)
				continue
			}
		}

		if want, got := tt.value, value; !bytes.Equal(want, got) {
			t.Errorf("[%02d] test %q, unexpected value for Options.GetOne(%v):\n- want: %v\n-  got: %v",
				i, tt.desc, tt.key, want, got)
		}
	}
}

// Test_parseOptions verifies that parseOptions parses correct option values
// from a slice of bytes, and that it returns an empty Options map if the byte
// slice cannot contain options.
func Test_parseOptions(t *testing.T) {
	var tests = []struct {
		desc    string
		buf     []byte
		options Options
		err     error
	}{
		{
			desc:    "nil options bytes",
			options: Options{},
		},
		{
			desc:    "empty options bytes",
			buf:     []byte{},
			options: Options{},
		},
		{
			desc: "too short options bytes",
			buf:  []byte{0},
			err:  ErrInvalidOptions,
		},
		{
			desc:    "zero code, zero length option bytes",
			buf:     []byte{0, 0, 0, 0},
			options: Options{},
		},
		{
			desc: "zero code, zero length option bytes with trailing byte",
			buf:  []byte{0, 0, 0, 0, 1},
			err:  ErrInvalidOptions,
		},
		{
			desc: "zero code, length 3, incorrect length for data",
			buf:  []byte{0, 0, 0, 3, 1, 2},
			err:  ErrInvalidOptions,
		},
		{
			desc: "client ID, length 1, value [1]",
			buf:  []byte{0, 1, 0, 1, 1},
			options: Options{
				OptionClientID: [][]byte{{1}},
			},
		},
		{
			desc: "client ID, length 2, value [1 1] + server ID, length 3, value [1 2 3]",
			buf: []byte{
				0, 1, 0, 2, 1, 1,
				0, 2, 0, 3, 1, 2, 3,
			},
			options: Options{
				OptionClientID: [][]byte{{1, 1}},
				OptionServerID: [][]byte{{1, 2, 3}},
			},
		},
	}

	for i, tt := range tests {
		var options Options
		err := (&options).UnmarshalBinary(tt.buf)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Errorf("[%02d] test %q, unexpected error for parseOptions(%v): %v != %v",
					i, tt.desc, tt.buf, want, got)
			}
			continue
		}

		if want, got := tt.options, options; !reflect.DeepEqual(want, got) {
			t.Errorf("[%02d] test %q, unexpected Options map for parseOptions(%v):\n- want: %v\n-  got: %v",
				i, tt.desc, tt.buf, want, got)
		}

		for k, v := range tt.options {
			for ii := range v {
				if want, got := cap(v[ii]), cap(options[k][ii]); want != got {
					t.Errorf("[%02d] test %q, option %d, unexpected capacity option data:\n- want: %v\n-  got: %v",
						i, tt.desc, ii, want, got)
				}
			}
		}
	}
}
