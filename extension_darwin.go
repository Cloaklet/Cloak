//+build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#include "extension_darwin.h"
*/
import "C"
import "unsafe"

// revealPath reveals given path in macOS Finder app.
// The actual implementation is in Objective-C.
func revealPath(path string) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.RevealInFinder(cPath)
}
