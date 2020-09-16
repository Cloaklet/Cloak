//+build linux

package extension

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

// revealPath reveals given path in macOS Finder app.
func revealPath(path string) {
	// FIXME
}

// isFuseAvailable detects FUSE availability for Linux.
func isFuseAvailable() bool {
	path, err := exec.LookPath("fusermount")
	return err == nil && path != ""
}

// locateLogDirectory returns the path in which log files should be stored.
// The directory gets created if it does not exist.
func locateLogDirectory() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	logDir := filepath.Join(currentUser.HomeDir, ".cloaklet.cloak", "logs")

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
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(currentUser.HomeDir, ".cloaklet.cloak", "data")

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