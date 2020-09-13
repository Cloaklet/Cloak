//+build darwin

package extension

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#include "extension_darwin.h"
*/
import "C"
import (
	"os"
	"unsafe"
)

// revealPath reveals given path in macOS Finder app.
// The actual implementation is in Objective-C.
func revealPath(path string) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.RevealInFinder(cPath)
}

// isFuseAvailable detects FUSE availability for macOS.
// OSXFUSE provides FUSE support for macOS.
func isFuseAvailable() bool {
	loadFuseBin := "/Library/Filesystems/osxfuse.fs/Contents/Resources/load_osxfuse"
	if info, err := os.Stat(loadFuseBin); err == nil {
		return !info.IsDir()
	}
	return false
}
