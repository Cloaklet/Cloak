//+build linux

package extension

import (
	"fmt"
	"github.com/adrg/xdg"
	"os"
	"os/exec"
	"path/filepath"
)

// openPath opens given path in OS file manager.
func openPath(path string) {
	xdgOpen, err := exec.LookPath("xdg-open")
	if err != nil {
		return
	}
	proc := exec.Command(xdgOpen, path)
	proc.Run()
}

// isFuseAvailable detects FUSE availability for Linux.
func isFuseAvailable() bool {
	path, err := exec.LookPath("fusermount")
	return err == nil && path != ""
}

// locateLogDirectory returns the path in which log files should be stored.
// The directory gets created if it does not exist.
func locateLogDirectory() (string, error) {
	dataDir, err := locateAppDataDirectory()
	if err != nil {
		return "", err
	}
	logDir := filepath.Join(dataDir, "logs")

	var info os.FileInfo
	if info, err = os.Stat(logDir); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil && !info.IsDir() {
		return "", fmt.Errorf("%s should be a directory but it is a file", logDir)
	}
	if err != nil && os.IsNotExist(err) {
		return logDir, os.MkdirAll(logDir, 0750)
	}
	return logDir, err
}

// locateAppDataDirectory returns path of where we should store our data for current user.
// The directory gets created if it does not exist.
func locateAppDataDirectory() (string, error) {
	var err error
	dataDir := filepath.Join(xdg.DataHome, "Cloak")

	var info os.FileInfo
	if info, err = os.Stat(dataDir); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil && !info.IsDir() {
		return "", fmt.Errorf("%s should be a directory but it is a file", dataDir)
	}
	if err != nil && os.IsNotExist(err) {
		return dataDir, os.MkdirAll(dataDir, 0750)
	}
	return dataDir, err
}
