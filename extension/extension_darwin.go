//+build darwin

package extension

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation
#include "extension_darwin.h"
*/
import "C"
import (
	"fmt"
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
	loadFuseBin := "/Library/Filesystems/osxfuse.fs/Contents/Resources/load_osxfuse"
	if info, err := os.Stat(loadFuseBin); err == nil {
		return !info.IsDir()
	}
	return false
}

// locateAppDataDirectory returns path of the "Application Support" of current user.
func locateAppDataDirectory() (string, error) {
	var err error
	appDataDir := filepath.Join(C.GoString(C.GetAppDataDirectory()), "Cloak")

	if _, err = os.Stat(appDataDir); os.IsNotExist(err) {
		if err = os.MkdirAll(appDataDir, 0700); err != nil {
			return "", err
		}
	}
	return appDataDir, err
}

// locateLogDirectory returns the path in which log files should be stored.
// The directory gets created if it does not exist.
func locateLogDirectory() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	logDir := filepath.Join(currentUser.HomeDir, "Library", "Logs")

	var info os.FileInfo
	if info, err = os.Stat(logDir); err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", logDir)
	}
	logDir = filepath.Join(logDir, "Cloak")
	if _, err := os.Stat(logDir); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		if err = os.Mkdir(logDir, 0750); err != nil {
			return "", err
		}
	}
	return logDir, nil
}
