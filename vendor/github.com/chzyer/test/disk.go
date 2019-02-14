package test

import (
	"encoding/hex"
	"io"
)

type MemDisk struct {
	data       [][]byte
	size       int64
	woff, roff int
}

func NewMemDisk() *MemDisk {
	return &MemDisk{}
}

func (w *MemDisk) Write(b []byte) (int, error) {
	n, err := w.WriteAt(b, int64(w.woff))
	w.woff += n
	return n, err
}

func (w *MemDisk) getData(off int64) []byte {
	idx := int(off >> 20)
	if idx >= cap(w.data) {
		newdata := make([][]byte, idx+1)
		copy(newdata, w.data)
		w.data = newdata
	}
	if len(w.data[idx]) == 0 {
		w.data[idx] = make([]byte, 1<<20)
	}

	return w.data[idx][off&((1<<20)-1):]
}

func (w *MemDisk) WriteAt(b []byte, off int64) (int, error) {
	n := len(b)
	for len(b) > 0 {
		buf := w.getData(off)
		m := copy(buf, b)
		if off+int64(m) > w.size {
			w.size = off + int64(m)
		}
		b = b[m:]
		off += int64(m)
	}
	return n, nil
}

func (w *MemDisk) ReadAt(b []byte, off int64) (int, error) {
	byteRead := 0
	for byteRead < len(b) {
		if off >= w.size {
			return 0, io.EOF
		}
		buf := w.getData(off)
		if int64(len(buf))+off > w.size {
			buf = buf[:w.size-off]
		}
		if len(buf) == 0 {
			return byteRead, io.EOF
		}
		n := copy(b[byteRead:], buf)
		off += int64(n)
		byteRead += n
	}
	return byteRead, nil
}

func (w *MemDisk) Dump() string {
	return hex.Dump(w.getData(0))
}

func (w *MemDisk) SeekRead(offset int64, whence int) (ret int64) {
	switch whence {
	case 0:
		w.roff += int(offset)
	case 1:
		w.roff = int(offset)
	default:
	}
	return int64(w.roff)
}

func (w *MemDisk) SeekWrite(offset int64, whence int) (ret int64) {
	switch whence {
	case 0:
		w.woff += int(offset)
	case 1:
		w.woff = int(offset)
	default:
	}
	return int64(w.woff)
}

func (w *MemDisk) Read(b []byte) (int, error) {
	n, err := w.ReadAt(b, int64(w.roff))
	w.roff += n
	return n, err
}

func (w *MemDisk) Close() error {
	return nil
}
