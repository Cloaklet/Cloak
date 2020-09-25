package server

import (
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
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
func (s *ApiServer) GocryptfsChangeVaultPassword(path string, password string, newPassword string) error {
	chPwProc := exec.Command(s.cmd, "-passwd", "--", path)
	// Password is piped through STDIN
	stdIn, err := chPwProc.StdinPipe()
	var errorOutput bytes.Buffer
	chPwProc.Stderr = &errorOutput
	if err != nil {
		logger.Error().Err(err).
			Str("vaultPath", path).
			Msg("Failed to create STDIN pipe when changing password for vault")
		return err
	}

	go func() {
		defer stdIn.Close()
		passwords := strings.Join([]string{password, newPassword}, "\n")
		if _, err := io.WriteString(stdIn, passwords); err != nil {
			logger.Error().Err(err).
				Str("vaultPath", path).
				Msg("Failed to pipe passwords to gocryptfs when changing password for vault")
		}
	}()

	// Vault created, add to vault repository
	if err := chPwProc.Run(); err != nil { // Failed to init vault, inspect error and respond to UI
		rc := chPwProc.ProcessState.ExitCode()
		errString := errorOutput.String()
		errlog := logger.With().Err(err).
			Int("RC", rc).
			Str("vaultPath", path).
			Str("stdErr", errString).
			Logger()
		switch rc {
		case 12:
			errlog.Error().Msg("Password incorrect")
			return ErrWrongPassword
		case 23:
			errlog.Error().Msg("Gocryptfs could not open gocryptfs.conf for reading")
			return ErrCantOpenVaultConf
		case 24:
			errlog.Error().Msg("Gocryptfs could not write the updated gocryptfs.conf")
			return ErrVaultUpdateConfFailed
		default:
			errlog.Error().Msg("Unknown error when changing password for vault")
			return ErrUnknown.Reformat(errString)
		}
	}
	return nil
}

// GocryptfsShowVaultMasterkey reveals masterkey for vault identified by `path` directory.
// Returns (masterkey, error).
// Notice: `path` is the path to the vault directory.
func (s *ApiServer) GocryptfsShowVaultMasterkey(path string, password string) (string, error) {
	var err error
	var masterkey string

	vaultConfigPath := filepath.Join(path, "gocryptfs.conf")
	xrayProc := exec.Command(s.xrayCmd, "-dumpmasterkey", vaultConfigPath)
	// Password is piped through STDIN
	stdIn, err := xrayProc.StdinPipe()
	var errorOutput, stdOutput bytes.Buffer
	xrayProc.Stderr = &errorOutput
	xrayProc.Stdout = &stdOutput
	if err != nil {
		logger.Error().Err(err).
			Str("vaultPath", path).
			Msg("Failed to create STDIN pipe when revealing masterkey for vault")
		return masterkey, err
	}

	go func() {
		defer stdIn.Close()
		if _, err := io.WriteString(stdIn, password); err != nil {
			logger.Error().Err(err).
				Str("vaultPath", path).
				Msg("Failed to pipe passwords to gocryptfs when revealing masterkey for vault")
		}
	}()

	// Vault created, add to vault repository
	if err := xrayProc.Run(); err != nil { // Failed to init vault, inspect error and respond to UI
		rc := xrayProc.ProcessState.ExitCode()
		errString := errorOutput.String()
		outString := stdOutput.String()
		errlog := logger.With().Err(err).
			Int("RC", rc).
			Str("vaultPath", path).
			Str("stdErr", errString).
			Str("stdOut", outString).
			Logger()
		switch rc {
		case 12:
			errlog.Error().Msg("Password incorrect")
			return masterkey, ErrWrongPassword
		case 23:
			errlog.Error().Msg("Gocryptfs could not open gocryptfs.conf for reading")
			return masterkey, ErrCantOpenVaultConf
		case 24:
			errlog.Error().Msg("Gocryptfs could not write the updated gocryptfs.conf")
			return masterkey, ErrVaultUpdateConfFailed
		default:
			errlog.Error().Msg("Unknown error when changing password for vault")
			if strings.TrimSpace(errString) == "" {
				errString = outString
			}
			return masterkey, ErrUnknown.Reformat(errString)
		}
	}
	masterkey = strings.TrimSpace(stdOutput.String())
	return masterkey, nil
}
