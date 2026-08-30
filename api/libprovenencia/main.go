// Package main is the C-shared library entrypoint (not a CLI).
// go build -buildmode=c-shared requires package main; dispatch lives in api/ffi.
package main

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"unsafe"

	"github.com/mendahu/provenencia/api/ffi"
)

func bytesFromC(ptr *C.uint8_t, n C.size_t) []byte {
	if ptr == nil || n == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(ptr), C.int(n))
}

//export provenencia_call
func provenencia_call(method C.int32_t, in *C.uint8_t, inLen C.size_t, out **C.uint8_t, outLen *C.size_t) C.int {
	resp, err := ffi.Call(int32(method), bytesFromC(in, inLen))
	if err != nil {
		return copyOut(err.Error(), out, outLen, 1)
	}
	if len(resp) == 0 {
		*out = nil
		*outLen = 0
		return 0
	}
	return copyOutBytes(resp, out, outLen, 0)
}

func copyOut(s string, out **C.uint8_t, outLen *C.size_t, status C.int) C.int {
	return copyOutBytes([]byte(s), out, outLen, status)
}

func copyOutBytes(b []byte, out **C.uint8_t, outLen *C.size_t, status C.int) C.int {
	if len(b) == 0 {
		*out = nil
		*outLen = 0
		return status
	}
	buf := C.malloc(C.size_t(len(b)))
	if buf == nil {
		return 2
	}
	C.memcpy(buf, unsafe.Pointer(&b[0]), C.size_t(len(b)))
	*out = (*C.uint8_t)(buf)
	*outLen = C.size_t(len(b))
	return status
}

//export provenencia_free
func provenencia_free(p unsafe.Pointer) {
	if p != nil {
		C.free(p)
	}
}

func main() {}
