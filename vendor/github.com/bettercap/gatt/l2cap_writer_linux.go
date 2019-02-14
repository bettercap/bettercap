package gatt

import "encoding/binary"

// l2capWriter helps create l2cap responses.
// It is not meant to be used with large writes.
// TODO: benchmark the number of allocs here.
// Reduce by letting WriteByteFit, WriteUint16Fit, etc.
// extend b/chunk and write into it directly.
type l2capWriter struct {
	mtu     int
	b       []byte
	chunk   []byte
	chunked bool
}

func newL2capWriter(mtu uint16) *l2capWriter {
	return &l2capWriter{mtu: int(mtu), b: make([]byte, 0, mtu)}
}

// Chunk starts writing a new chunk. This chunk
// is not committed until Commit is called.
// Chunk panics if another chunk has already been
// started and not committed.
func (w *l2capWriter) Chunk() {
	if w.chunked {
		panic("l2capWriter: chunk called twice without committing")
	}
	w.chunked = true
	if w.chunk == nil {
		w.chunk = make([]byte, 0, w.mtu)
	}
}

// Commit writes the current chunk and reports whether the
// write succeeded. The write succeeds iff there is enough room.
// Commit panics if no chunk has been started.
func (w *l2capWriter) Commit() bool {
	if !w.chunked {
		panic("l2capWriter: commit without starting a chunk")
	}
	var success bool
	if len(w.b)+len(w.chunk) <= w.mtu {
		success = true
		w.b = append(w.b, w.chunk...)
	}
	w.chunk = w.chunk[:0]
	w.chunked = false
	return success
}

// CommitFit writes as much of the current chunk as possible,
// truncating as needed.
// CommitFit panics if no chunk has been started.
func (w *l2capWriter) CommitFit() {
	if !w.chunked {
		panic("l2capWriter: CommitFit without starting a chunk")
	}
	writeable := w.mtu - len(w.b)
	if writeable > len(w.chunk) {
		writeable = len(w.chunk)
	}
	w.b = append(w.b, w.chunk[:writeable]...)
	w.chunk = w.chunk[:0]
	w.chunked = false
}

// WriteByteFit writes b.
// It reports whether the write succeeded,
// using the criteria of WriteFit.
func (w *l2capWriter) WriteByteFit(b byte) bool {
	return w.WriteFit([]byte{b})
}

// WriteUint16Fit writes v using BLE (LittleEndian) encoding.
// It reports whether the write succeeded, using the
// criteria of WriteFit.
func (w *l2capWriter) WriteUint16Fit(v uint16) bool {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return w.WriteFit(b)
}

// WriteUUIDFit writes uuid using BLE (reversed) encoding.
// It reports whether the write succeeded, using the
// criteria of WriteFit.
func (w *l2capWriter) WriteUUIDFit(u UUID) bool {
	return w.WriteFit(u.b)
}

// Writeable returns the number of bytes from b
// that would be written if pad bytes were written,
// then as much of b as fits were written. When
// writing to a chunk, any amount of bytes may be
// written.
func (w *l2capWriter) Writeable(pad int, b []byte) int {
	if w.chunked {
		return len(b)
	}
	avail := w.mtu - len(w.b) - pad
	if avail > len(b) {
		return len(b)
	}
	if avail < 0 {
		return 0
	}
	return avail
}

// WriteFit writes as much of b as fits.
// It reports whether the write succeeded without
// truncation. A write succeeds without truncation
// iff a chunk write is in progress or the entire
// contents were written (without exceeding the mtu).
func (w *l2capWriter) WriteFit(b []byte) bool {
	if w.chunked {
		w.chunk = append(w.chunk, b...)
		return true
	}
	avail := w.mtu - len(w.b)
	if avail >= len(b) {
		w.b = append(w.b, b...)
		return true
	}
	w.b = append(w.b, b[:avail]...)
	return false
}

// ChunkSeek discards the first offset bytes from the
// current chunk. It reports whether there were at least
// offset bytes available to discard.
// It panics if a chunked write is not in progress.
func (w *l2capWriter) ChunkSeek(offset uint16) bool {
	if !w.chunked {
		panic("l2capWriter: ChunkSeek requested without chunked write in progress")
	}
	if len(w.chunk) < int(offset) {
		w.chunk = w.chunk[:0]
		return false
	}
	w.chunk = w.chunk[offset:]
	return true
}

// Bytes returns the written bytes.
// It will panic if a chunked write
// is in progress.
// It is meant to be used when writing
// is completed. It does not return a copy.
// Don't abuse this, it's not worth it.
func (w *l2capWriter) Bytes() []byte {
	if w.chunked {
		panic("l2capWriter: Bytes requested while chunked write in progress")
	}
	return w.b
}
