package internal

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrInvalidMagic = errors.New("Invalid magic")

	sizeEncoding = binary.BigEndian

	magicText = []byte("ENDSLEY/BSDIFF43")
)

func WriteHeader(w io.Writer, size uint64) (err error) {
	if _, err = w.Write(magicText); err != nil {
		return
	}
	err = binary.Write(w, sizeEncoding, size)
	return
}

func ReadHeader(r io.Reader) (size uint64, err error) {
	magicBuf := make([]byte, len(magicText))
	n, err := r.Read(magicBuf)
	if err != nil {
		return
	}
	if n < len(magicText) {
		err = ErrInvalidMagic
		return
	}

	err = binary.Read(r, sizeEncoding, &size)

	return
}
