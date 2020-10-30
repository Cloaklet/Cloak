package extension

import (
	"fmt"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
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

	migrateLegacyDirectories()
}

// migrateLegacyDirectories migrates legacy directories for Linux (< v0.8.0).
// It might panic.
func migrateLegacyDirectories() {
	logger := GetLogger("extension")
	if runtime.GOOS != "linux" {
		return
	}

	currentUser, err := user.Current()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to get current user")
	}

	// Make sure new locations exist
	for dirName, dirPathFunc := range map[string]func() string{
		"appDataDir": GetAppDataDirectory,
		"configDir":  GetConfigDirectory,
		"logDir":     locateLogDirectory,
	} {
		dirPath, dirErr := EnsureDirectoryExists(dirPathFunc())
		if dirErr != nil {
			logger.Fatal().Err(dirErr).
				Str("name", dirName).
				Str("path", dirPath).
				Msg("Failed to ensure directory existence")
		}
	}

	// Move database file
	// Move files
	for fromPath, toPath := range map[string]string{
		filepath.Join(currentUser.HomeDir, ".cloaklet.cloak", "data", "vaults.db"):   filepath.Join(GetAppDataDirectory(), "vaults.db"),
		filepath.Join(currentUser.HomeDir, ".cloaklet.cloak", "data", "options.ini"): filepath.Join(GetConfigDirectory(), "options.ini"),
	} {
		if err := os.Rename(fromPath, toPath); err != nil && !os.IsNotExist(err) {
			logger.Fatal().Err(err).
				Str("from", fromPath).
				Str("to", toPath).
				Msg("Failed to move file to its new location")
		}
	}

	// Remove old directory
	legacyDataRoot := filepath.Join(currentUser.HomeDir, ".cloaklet.cloak")
	if err := os.RemoveAll(legacyDataRoot); err != nil && !os.IsNotExist(err) {
		logger.Warn().Err(err).
			Str("path", legacyDataRoot).
			Msg("Failed to delete legacy directory")
	} else {
		logger.Info().Str("path", legacyDataRoot).
			Msg("Migrated old data files")
	}
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
	if info != nil && !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	return path, nil

}
