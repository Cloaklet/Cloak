package extension

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func init() {
	// Global zerolog settings
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

// RevealInFileManager is a general function which calls platform-dependent implementations
// to reveal given path in its parent directory in OS file manager.
func RevealInFileManager(path string) {
	revealPath(path)
}

// LocateGocryptfsBinary locates gocryptfs binary and returns its absolute path.
// It first looks for the binary in the same directory as current running executable,
// then falls back to anything found in PATH environment.
func LocateGocryptfsBinary() (path string, err error) {
	if executable, err := os.Executable(); err == nil {
		path = filepath.Join(filepath.Dir(executable), "gocryptfs")
		if _, err = os.Stat(path); os.IsNotExist(err) {
			path, err = exec.LookPath("gocryptfs")
		}
	}
	return
}

// GetLogger creates a new zerolog logger with given string as vaule for `module` key.
func GetLogger(module string) zerolog.Logger {
	// Derive from the global logger so all settings are unified
	return zlog.Logger.With().Str("module", module).Logger()
}

// IsFuseAvailable returns a bool value indicating FUSE ability of current OS.
func IsFuseAvailable() bool {
	return isFuseAvailable()
}
