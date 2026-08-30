package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import "unsafe"

// callForTest is used by call_test.go. It cannot live in a *_test.go file:
// package main does not support cgo in tests. This filename must not end in
// _test.go or the compiler treats it as a test file.
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
	st := provenencia_call(C.int32_t(method), inPtr, inLen, &outPtr, &outLen)
	outWasNil = outPtr == nil
	if outPtr != nil {
		out = C.GoBytes(unsafe.Pointer(outPtr), C.int(outLen))
		provenencia_free(unsafe.Pointer(outPtr))
	}
	return int(st), out, outWasNil
}
