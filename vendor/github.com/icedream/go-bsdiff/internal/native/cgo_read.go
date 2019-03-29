package native

/*
#include "bspatch.h"
*/
import "C"
import (
	"io"
	"log"
	"unsafe"
)

//export cgo_read_buffer
func cgo_read_buffer(bufferIndex C.int,
	bufPtr unsafe.Pointer, length C.int) C.int {
	goLength := int(length)

	if goLength == 0 {
		return 0
	}

	sourceBuffer := readers.Get(int(bufferIndex))
	targetBuffer := cPtrToSlice(bufPtr, goLength)

	errCode := 0
	offset := 0
	for offset < goLength {
		n, err := sourceBuffer.Read(targetBuffer)

		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("cgo_read_buffer failed:", err)
			errCode = 1
			break
		}

		offset += n
		targetBuffer = targetBuffer[n:]
	}

	return C.int(errCode)
}
