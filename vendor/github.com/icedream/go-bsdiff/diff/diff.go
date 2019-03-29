package diff

import (
	"io"
	"io/ioutil"

	"github.com/dsnet/compress/bzip2"
	"github.com/icedream/go-bsdiff/internal"
	"github.com/icedream/go-bsdiff/internal/native"
)

func Diff(oldReader, newReader io.Reader, patchWriter io.Writer) (err error) {
	oldBytes, err := ioutil.ReadAll(oldReader)
	if err != nil {
		return
	}
	newBytes, err := ioutil.ReadAll(newReader)
	if err != nil {
		return
	}

	if err = internal.WriteHeader(patchWriter, uint64(len(newBytes))); err != nil {
		return
	}

	// Compression
	bz2Writer, err := bzip2.NewWriter(patchWriter, nil)
	if err != nil {
		return
	}
	defer bz2Writer.Close()

	return native.Diff(oldBytes, newBytes, bz2Writer)
}
