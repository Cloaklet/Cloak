package extension

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
)

var ReleaseMode string

func init() {
	timeFormat := "2006-01-02T15:04:05.000000 MST"
	// Ref: https://github.com/rs/zerolog/issues/114
	zerolog.TimeFieldFormat = timeFormat

	// Global zerolog settings
	if ReleaseMode == "true" {
		if logDir, err := locateLogDirectory(); err == nil {
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
	if self, err := os.Executable(); err != nil {
		return "", err
	} else {
		path := filepath.Join(filepath.Dir(self), executable)
		if _, err = os.Stat(path); err != nil {
			return exec.LookPath(executable)
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
