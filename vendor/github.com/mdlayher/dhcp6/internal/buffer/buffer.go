package buffer

import (
	"encoding/binary"
)

var order = binary.BigEndian

// Buffer encapsulates marshaling unsigned integer and byte slice values.
type Buffer struct {
	// data is the underlying data.
	data []byte
}

// New consumes b for marshaling or unmarshaling.
func New(b []byte) *Buffer {
	return &Buffer{b}
}

// append appends n bytes to the Buffer and returns a slice pointing to the
// newly appended bytes.
func (b *Buffer) append(n int) []byte {
	b.data = append(b.data, make([]byte, n)...)
	return b.data[len(b.data)-n:]
}

// Data is unconsumed data remaining in the Buffer.
func (b *Buffer) Data() []byte {
	return b.data
}

// Remaining consumes and returns a copy of all remaining bytes in the Buffer.
func (b *Buffer) Remaining() []byte {
	p := b.Consume(len(b.Data()))
	cp := make([]byte, len(p))
	copy(cp, p)
	return cp
}

// consume consumes n bytes from the Buffer. It returns nil, false if there
// aren't enough bytes left.
func (b *Buffer) consume(n int) ([]byte, bool) {
	if !b.Has(n) {
		return nil, false
	}
	rval := b.data[:n]
	b.data = b.data[n:]
	return rval, true
}

// Consume consumes n bytes from the Buffer. It returns nil if there aren't
// enough bytes left.
func (b *Buffer) Consume(n int) []byte {
	v, ok := b.consume(n)
	if !ok {
		return nil
	}
	return v
}

// Has returns true if n bytes are available.
func (b *Buffer) Has(n int) bool {
	return len(b.data) >= n
}

// Len returns the length of the remaining bytes.
func (b *Buffer) Len() int {
	return len(b.data)
}

// Read8 reads a byte from the Buffer.
func (b *Buffer) Read8() uint8 {
	v, ok := b.consume(1)
	if !ok {
		return 0
	}
	return uint8(v[0])
}

// Read16 reads a 16-bit value from the Buffer.
func (b *Buffer) Read16() uint16 {
	v, ok := b.consume(2)
	if !ok {
		return 0
	}
	return order.Uint16(v)
}

// Read32 reads a 32-bit value from the Buffer.
func (b *Buffer) Read32() uint32 {
	v, ok := b.consume(4)
	if !ok {
		return 0
	}
	return order.Uint32(v)
}

// Read64 reads a 64-bit value from the Buffer.
func (b *Buffer) Read64() uint64 {
	v, ok := b.consume(8)
	if !ok {
		return 0
	}
	return order.Uint64(v)
}

// ReadBytes reads exactly len(p) values from the Buffer.
func (b *Buffer) ReadBytes(p []byte) {
	copy(p, b.Consume(len(p)))
}

// Write8 writes a byte to the Buffer.
func (b *Buffer) Write8(v uint8) {
	b.append(1)[0] = byte(v)
}

// Write16 writes a 16-bit value to the Buffer.
func (b *Buffer) Write16(v uint16) {
	order.PutUint16(b.append(2), v)
}

// Write32 writes a 32-bit value to the Buffer.
func (b *Buffer) Write32(v uint32) {
	order.PutUint32(b.append(4), v)
}

// Write64 writes a 64-bit value to the Buffer.
func (b *Buffer) Write64(v uint64) {
	order.PutUint64(b.append(8), v)
}

// WriteN returns a newly appended n-size Buffer to write to.
func (b *Buffer) WriteN(n int) []byte {
	return b.append(n)
}

// WriteBytes writes p to the Buffer.
func (b *Buffer) WriteBytes(p []byte) {
	copy(b.append(len(p)), p)
}
