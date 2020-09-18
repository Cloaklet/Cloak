package extension

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var ReleaseMode string

func init() {
	// Global zerolog settings
	if ReleaseMode == "true" {
		if logDir, err := locateLogDirectory(); err == nil {
			logFile, err := os.OpenFile(filepath.Join(logDir, "Cloak.log"), os.O_WRONLY|os.O_CREATE, 0640)
			if err == nil {
				_ = logFile.Truncate(0)
				zlog.Logger = zlog.Output(logFile)
				return
			}
		}
	}
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
func LocateGocryptfsBinary() (string, error) {
	if executable, err := os.Executable(); err != nil {
		return "", err
	} else {
		path := filepath.Join(filepath.Dir(executable), "gocryptfs")
		if _, err = os.Stat(path); err != nil {
			return exec.LookPath("gocryptfs")
		}
		return path, nil
	}
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

func GetAppDataDirectory() (string, error) {
	return locateAppDataDirectory()
}
