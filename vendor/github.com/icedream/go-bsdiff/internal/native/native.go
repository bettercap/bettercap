package native

/*
#include <stdint.h>
*/
import "C"
import (
	"reflect"
	"unsafe"
)

var (
	writers = writerTable{}
	readers = readerTable{}
)

func bytesToUint8PtrAndSize(bytes []byte) (ptr *C.uint8_t, size C.int64_t) {
	ptr = (*C.uint8_t)(unsafe.Pointer(&bytes[0]))
	size = C.int64_t(int64(len(bytes)))
	return
}

func cPtrToSlice(ptr unsafe.Pointer, size int) []byte {
	var slice []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = size
	sliceHeader.Len = size
	sliceHeader.Data = uintptr(ptr)

	return slice
}
