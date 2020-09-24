package server

import (
	"bytes"
	"io"
	"os/exec"
)

/*
This file contains (hopefully) all the necessary utility functions to interact with gocryptfs tool.
All the utility functions:
  - return ApiError if possible;
  - do nothing but running gocryptfs process and logging;
*/

// GocryptfsCreateVault creates a new vault at `path` with `password`.
func (s *ApiServer) GocryptfsCreateVault(path string, password string) error {
	// Start a gocryptfs process to init this vault
	initProc := exec.Command(s.cmd, "-init", "--", path)
	// Password is piped through STDIN
	stdIn, err := initProc.StdinPipe()
	var errorOutput bytes.Buffer
	initProc.Stderr = &errorOutput
	if err != nil {
		logger.Error().Err(err).
			Str("vaultPath", path).
			Msg("Failed to create STDIN pipe when initializing new vault")
		return err
	}

	go func() {
		defer stdIn.Close()
		if _, err := io.WriteString(stdIn, password); err != nil {
			logger.Error().Err(err).
				Str("vaultPath", path).
				Msg("Failed to pipe vault password to gocryptfs when initializing new vault")
		}
	}()

	// Vault created, add to vault repository
	if err = initProc.Run(); err != nil { // Failed to init vault, inspect error and respond to UI
		rc := initProc.ProcessState.ExitCode()
		errString := errorOutput.String()
		errlog := logger.With().Err(err).
			Int("RC", rc).
			Str("vaultPath", path).
			Str("stdErr", errString).
			Logger()
		switch rc {
		case 6:
			errlog.Error().Msg("New vault directory (CIPHERDIR) is not empty")
			return ErrVaultDirNotEmpty
		case 22:
			errlog.Error().Msg("Password for new vault is empty")
			return ErrVaultPasswordEmpty
		case 24:
			errlog.Error().Msg("Gocryptfs could not create gocryptfs.conf")
			return ErrVaultInitConfFailed
		default:
			errlog.Error().Msg("Unknown error when initializing new vault")
			return ErrUnknown.Reformat(errString)
		}
	}
	return err
}

// GocryptfsChangeVaultPassword changes password for vault identified by `path` directory.
func GocryptfsChangeVaultPassword(path string, password string, newPassword string) error {
	return nil // FIXME
}

// GocryptfsShowVaultMasterkey reveals masterkey for vault identified by `path` directory.
// Returns (masterkey, error).
func GocryptfsShowVaultMasterkey(path string, password string) (string, error) {
	return "", nil // FIXME
}
