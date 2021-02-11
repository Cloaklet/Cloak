//+build darwin

package extension

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation
#include "extension_darwin.h"
*/
import "C"
import (
	"os"
	"os/user"
	"path/filepath"
	"unsafe"
)

// openPath opens given path in macOS Finder app.
// The actual implementation is in Objective-C.
func openPath(path string) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.OpenPath(cPath)
}

// isFuseAvailable detects FUSE availability for macOS.
// OSXFUSE provides FUSE support for macOS.
func isFuseAvailable() bool {
	for _, fuseBin := range []string{
		"/Library/Filesystems/macfuse.fs/Contents/Resources/load_macfuse",
		"/Library/Filesystems/osxfuse.fs/Contents/Resources/load_osxfuse",
	} {
		info, err := os.Stat(fuseBin)
		if err != nil {
			continue
		}
		return !info.IsDir()
	}
	return false
}

// locateAppDataDirectory returns path of the "Application Support" of current user.
func locateAppDataDirectory() string {
	return filepath.Join(C.GoString(C.GetAppDataDirectory()), "Cloak")
}

// locateConfigDirectory returns path of the "Application Support" of current user.
func locateConfigDirectory() string {
	return locateAppDataDirectory()
}

// locateLogDirectory returns the path in which log files should be stored.
// The directory gets created if it does not exist.
func locateLogDirectory() string {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(currentUser.HomeDir, "Library", "Logs")
}
