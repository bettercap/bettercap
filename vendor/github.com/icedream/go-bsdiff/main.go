package bsdiff

import (
	"io"

	"github.com/icedream/go-bsdiff/diff"
	"github.com/icedream/go-bsdiff/patch"
)

func Diff(oldReader, newReader io.Reader, patchWriter io.Writer) (err error) {
	return diff.Diff(oldReader, newReader, patchWriter)
}

func Patch(oldReader io.Reader, newWriter io.Writer, patchReader io.Reader) (err error) {
	return patch.Patch(oldReader, newWriter, patchReader)
}
