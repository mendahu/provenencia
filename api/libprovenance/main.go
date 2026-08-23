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

	"github.com/mendahu/provenance/api/ffi"
)

func bytesFromC(ptr *C.uint8_t, n C.size_t) []byte {
	if ptr == nil || n == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(ptr), C.int(n))
}

// callForTest runs provenance_call and provenance_free so tests can cover C malloc
// without cgo in *_test.go (not supported in package main).
func callForTest(method int32, in []byte) (status int, out []byte, outWasNil bool) {
	var inPtr *C.uint8_t
	var inLen C.size_t
	if len(in) > 0 {
		inPtr = (*C.uint8_t)(C.CBytes(in))
		defer C.free(unsafe.Pointer(inPtr))
		inLen = C.size_t(len(in))
	}
	var outPtr *C.uint8_t
	var outLen C.size_t
	st := provenance_call(C.int32_t(method), inPtr, inLen, &outPtr, &outLen)
	outWasNil = outPtr == nil
	if outPtr != nil {
		out = C.GoBytes(unsafe.Pointer(outPtr), C.int(outLen))
		provenance_free(unsafe.Pointer(outPtr))
	}
	return int(st), out, outWasNil
}

//export provenance_call
func provenance_call(method C.int32_t, in *C.uint8_t, inLen C.size_t, out **C.uint8_t, outLen *C.size_t) C.int {
	resp, err := ffi.Call(int32(method), bytesFromC(in, inLen))
	if err != nil {
		return 1
	}
	if len(resp) == 0 {
		*out = nil
		*outLen = 0
		return 0
	}
	buf := C.malloc(C.size_t(len(resp)))
	if buf == nil {
		return 2
	}
	C.memcpy(buf, unsafe.Pointer(&resp[0]), C.size_t(len(resp)))
	*out = (*C.uint8_t)(buf)
	*outLen = C.size_t(len(resp))
	return 0
}

//export provenance_free
func provenance_free(p unsafe.Pointer) {
	if p != nil {
		C.free(p)
	}
}

func main() {}
