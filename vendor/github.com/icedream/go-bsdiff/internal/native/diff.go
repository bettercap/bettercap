package native

/*
#cgo CFLAGS: -I../../bsdiff

#include "bsdiff.h"
#include "cgo.h"
*/
import "C"
import (
	"errors"
	"io"
)

func Diff(oldbytes, newbytes []byte, patch io.Writer) (err error) {
	oldptr, oldsize := bytesToUint8PtrAndSize(oldbytes)
	newptr, newsize := bytesToUint8PtrAndSize(newbytes)

	bufferIndex := writers.Add(patch)
	defer writers.Free(bufferIndex)

	errCode := int(C.bsdiff_cgo(oldptr, oldsize, newptr, newsize, C.int(bufferIndex)))
	if errCode != 0 {
		err = errors.New("bsdiff failed")
		return
	}

	return
}
