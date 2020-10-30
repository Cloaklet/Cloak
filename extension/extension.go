package extension

import (
	"fmt"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
)

// ReleaseMode is set in build time (to either "true" or "false"), it
// indicates whether the app runs in release (production) mode.
var ReleaseMode string

func init() {
	timeFormat := "2006-01-02T15:04:05.000000 MST"
	// Ref: https://github.com/rs/zerolog/issues/114
	zerolog.TimeFieldFormat = timeFormat

	// Global zerolog settings
	if ReleaseMode == "true" {
		if logDir, err := EnsureDirectoryExists(locateLogDirectory()); err == nil {
			logFile, err := os.OpenFile(filepath.Join(logDir, "Cloak.log"), os.O_WRONLY|os.O_CREATE, 0640)
			if err == nil {
				_ = logFile.Truncate(0)
				zlog.Logger = zlog.Output(zerolog.ConsoleWriter{
					NoColor:    true,
					Out:        logFile,
					TimeFormat: timeFormat,
				})
				return
			}
		}
	}
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat})

}

// OpenPath opens given path in OS file manager
func OpenPath(path string) {
	openPath(path)
}

// LocateBinary locates executable binary of given name and returns its absolute path.
// It first looks for the binary in the same directory as current running executable,
// then falls back to anything found in PATH environment.
func LocateBinary(executable string) (string, error) {
	var (
		self string
		err  error
	)
	if self, err = os.Executable(); err != nil {
		return "", err
	}
	path := filepath.Join(filepath.Dir(self), executable)
	if _, err = os.Stat(path); err != nil {
		return exec.LookPath(executable)
	}
	return path, nil
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

// GetAppDataDirectory locates a directory in which we can store our data.
// The directory might not exist yet.
func GetAppDataDirectory() string {
	return locateAppDataDirectory()
}

// GetConfigDirectory locates a directory in which we can store our configuration file.
// The directory might not exist yet.
func GetConfigDirectory() string {
	return locateConfigDirectory()
}

// EnsureDirectoryExists makes sure given directory path exists.
// If the directory cannot be created, or it is an existing file, an error is returned.
func EnsureDirectoryExists(path string) (string, error) {
	var info os.FileInfo
	var err error

	if info, err = os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		if err = os.Mkdir(path, 0750); err != nil {
			return "", err
		}
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	return path, nil

}
