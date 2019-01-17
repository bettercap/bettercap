package dhcp6

import (
	"encoding"
	"sort"

	"github.com/mdlayher/dhcp6/internal/buffer"
)

// Options is a map of OptionCode keys with a slice of byte slice values.
//
// Its methods can be used to easily check for additional information from a
// packet. Get and GetOne should be used to access data from Options.
type Options map[OptionCode][][]byte

// Add adds a new OptionCode key and BinaryMarshaler struct's bytes to the
// Options map.
func (o Options) Add(key OptionCode, value encoding.BinaryMarshaler) error {
	// Special case: since OptionRapidCommit actually has zero length, it is
	// possible for an option key to appear with no value.
	if value == nil {
		o.AddRaw(key, nil)
		return nil
	}

	b, err := value.MarshalBinary()
	if err != nil {
		return err
	}

	o.AddRaw(key, b)
	return nil
}

// AddRaw adds a new OptionCode key and raw value byte slice to the
// Options map.
func (o Options) AddRaw(key OptionCode, value []byte) {
	o[key] = append(o[key], value)
}

// Get attempts to retrieve all values specified by an OptionCode key.
//
// If a value is found, get returns a non-nil [][]byte and nil. If it is not
// found, Get returns nil and ErrOptionNotPresent.
func (o Options) Get(key OptionCode) ([][]byte, error) {
	// Check for value by key.
	v, ok := o[key]
	if !ok {
		return nil, ErrOptionNotPresent
	}

	// Some options can actually have zero length (OptionRapidCommit), so
	// just return an empty byte slice if this is the case.
	if len(v) == 0 {
		return [][]byte{{}}, nil
	}
	return v, nil
}

// GetOne attempts to retrieve the first and only value specified by an
// OptionCode key. GetOne should only be used for OptionCode keys that must
// have at most one value.
//
// GetOne works just like Get, but if there is more than one value for the
// OptionCode key, ErrInvalidPacket will be returned.
func (o Options) GetOne(key OptionCode) ([]byte, error) {
	vv, err := o.Get(key)
	if err != nil {
		return nil, err
	}

	if len(vv) != 1 {
		return nil, ErrInvalidPacket
	}
	return vv[0], nil
}

// MarshalBinary allocates a buffer and writes options in their DHCPv6 binary
// format into the buffer.
func (o Options) MarshalBinary() ([]byte, error) {
	b := buffer.New(nil)
	for _, code := range o.sortedCodes() {
		for _, data := range o[code] {
			// 2 bytes: option code
			b.Write16(uint16(code))

			// 2 bytes: option length
			b.Write16(uint16(len(data)))

			// N bytes: option data
			b.WriteBytes(data)
		}
	}
	return b.Data(), nil
}

// UnmarshalBinary fills opts with option codes and corresponding values from
// an input byte slice.
//
// It is used with various different types to enable parsing of both top-level
// options, and options embedded within other options. If options data is
// malformed, it returns ErrInvalidOptions.
func (o *Options) UnmarshalBinary(p []byte) error {
	buf := buffer.New(p)
	*o = make(Options)

	for buf.Len() >= 4 {
		// 2 bytes: option code
		// 2 bytes: option length n
		// n bytes: data
		code := OptionCode(buf.Read16())
		length := buf.Read16()

		// N bytes: option data
		data := buf.Consume(int(length))
		if data == nil {
			return ErrInvalidOptions
		}
		data = data[:int(length):int(length)]

		o.AddRaw(code, data)
	}

	// Report error for any trailing bytes
	if buf.Len() != 0 {
		return ErrInvalidOptions
	}
	return nil
}

// optionCodes implements sort.Interface.
type optionCodes []OptionCode

func (b optionCodes) Len() int               { return len(b) }
func (b optionCodes) Less(i int, j int) bool { return b[i] < b[j] }
func (b optionCodes) Swap(i int, j int)      { b[i], b[j] = b[j], b[i] }

func (o Options) sortedCodes() optionCodes {
	var codes optionCodes
	for code := range o {
		codes = append(codes, code)
	}

	sort.Sort(codes)
	return codes
}
