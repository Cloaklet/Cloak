//go:build linux

package extension

import (
	"github.com/adrg/xdg"
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
// The directory might not exist yet.
func locateLogDirectory() string {
	return filepath.Join(xdg.DataHome, "Cloak", "logs")
}
