package server

import (
	"Cloak/extension"
	"Cloak/models"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
	This file contains (hopefully) all the necessary utility functions to interact with gocryptfs tool.
	All the utility functions:
	  - return ApiError if possible;
	  - do nothing but running gocryptfs process and logging;
*/

// Vault represents a single vault, as available to the frontend store.
type Vault struct {
	ID         int64
	Path       string
	MountPoint string
	AutoReveal bool
	ReadOnly   bool
	Unlocked   bool
}

// VaultManager is the main server type exposed to Wails frontend, for managing all vaults.
type VaultManager struct {
	repo          *models.VaultRepo      // database repository
	cmd           string                 // `gocryptfs` binary path
	xrayCmd       string                 // `gocryptfs-xray` binary path
	fuseAvailable bool                   // whether FUSE is available
	processes     map[int64]*exec.Cmd    // vaultID: process
	mountPoints   map[int64]string       // vaultID: mountPoint
	lock          sync.Mutex             // lock on `processes` and `mountPoints`
	configCh      chan map[string]string // channel for notifying config change requests
}

// Init init current manager instance.
func (m *VaultManager) Init() error {
	// Detect external runtime dependencies
	var err error
	if m.cmd, err = extension.LocateBinary("gocryptfs"); err != nil {
		logger.Error().Err(err).
			Msg("Failed to locate gocryptfs binary, nothing will work")
		return err
	} else {
		logger.Debug().Str("gocryptfs", m.cmd).Msg("Gocryptfs binary located")
	}
	if m.xrayCmd, err = extension.LocateBinary("gocryptfs-xray"); err != nil {
		logger.Error().Err(err).
			Msg("Failed to locate gocryptfs-xray binary, masterkey revealing will not work")
		return err
	} else {
		logger.Debug().Str("gocryptfs-xray", m.xrayCmd).Msg("gocryptfs-xray binary located")
	}
	logger.Debug().Bool("fuseAvailable", m.fuseAvailable).Msg("FUSE detection finished")

	// We use `rand` to generate random mountpoint name, so be sure to seed it upon start up
	rand.Seed(time.Now().UTC().UnixNano())

	return nil
}

func NewVaultManager(repo *models.VaultRepo, releaseMode bool, configCh chan map[string]string) *VaultManager {
	// Create manager
	return &VaultManager{
		repo:          repo,
		fuseAvailable: extension.IsFuseAvailable(),
		processes:     make(map[int64]*exec.Cmd),
		mountPoints:   make(map[int64]string),
		configCh:      configCh,
	}
}

// GocryptfsCreateVault creates a new vault at `path` with `password`.
func (m *VaultManager) GocryptfsCreateVault(path string, password string) error {
	// Start a gocryptfs process to init this vault
	initProc := exec.Command(m.cmd, "-init", "--", path)
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
		errLog := logger.With().
			Err(err).
			Int("RC", rc).
			Str("vaultPath", path).
			Str("stdErr", errString).
			Logger()
		switch rc {
		case 6:
			errLog.Error().Msg("New vault directory (CIPHERDIR) is not empty")
			return ErrVaultDirNotEmpty
		case 22:
			errLog.Error().Msg("Password for new vault is empty")
			return ErrVaultPasswordEmpty
		case 24:
			errLog.Error().Msg("Gocryptfs could not create gocryptfs.conf")
			return ErrVaultInitConfFailed
		default:
			errLog.Error().Msg("Unknown error when initializing new vault")
			return ErrUnknown.Reformat(errString)
		}
	}
	return err
}

// GocryptfsChangeVaultPassword changes password for vault identified by `path` directory.
func (m *VaultManager) GocryptfsChangeVaultPassword(path string, password string, newPassword string) error {
	chPwProc := exec.Command(m.cmd, "-passwd", "--", path)
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
		errLog := logger.With().
			Err(err).
			Int("RC", rc).
			Str("vaultPath", path).
			Str("stdErr", errString).
			Logger()
		switch rc {
		case 12:
			errLog.Error().Msg("Password incorrect")
			return ErrWrongPassword
		case 23:
			errLog.Error().Msg("Gocryptfs could not open gocryptfs.conf for reading")
			return ErrCantOpenVaultConf
		case 24:
			errLog.Error().Msg("Gocryptfs could not write the updated gocryptfs.conf")
			return ErrVaultUpdateConfFailed
		default:
			errLog.Error().Msg("Unknown error when changing password for vault")
			return ErrUnknown.Reformat(errString)
		}
	}
	return nil
}

// GocryptfsShowVaultMasterkey reveals masterkey for vault identified by `path` directory.
// Returns (masterkey, error).
// Notice: `path` is the path to the vault directory.
func (m *VaultManager) GocryptfsShowVaultMasterkey(path string, password string) (string, error) {
	var err error
	var masterKey string

	vaultConfigPath := filepath.Join(path, "gocryptfs.conf")
	xrayProc := exec.Command(m.xrayCmd, "-dumpmasterkey", vaultConfigPath)
	// Password is piped through STDIN
	stdIn, err := xrayProc.StdinPipe()
	var errorOutput, stdOutput bytes.Buffer
	xrayProc.Stderr = &errorOutput
	xrayProc.Stdout = &stdOutput
	if err != nil {
		logger.Error().Err(err).
			Str("vaultPath", path).
			Msg("Failed to create STDIN pipe when revealing masterkey for vault")
		return masterKey, err
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
		errLog := logger.With().
			Err(err).
			Int("RC", rc).
			Str("vaultPath", path).
			Str("stdErr", errString).
			Str("stdOut", outString).
			Logger()
		switch rc {
		case 12:
			errLog.Error().Msg("Password incorrect")
			return masterKey, ErrWrongPassword
		case 23:
			errLog.Error().Msg("Gocryptfs could not open gocryptfs.conf for reading")
			return masterKey, ErrCantOpenVaultConf
		case 24:
			errLog.Error().Msg("Gocryptfs could not write the updated gocryptfs.conf")
			return masterKey, ErrVaultUpdateConfFailed
		default:
			errLog.Error().Msg("Unknown error when changing password for vault")
			if strings.TrimSpace(errString) == "" {
				errString = outString
			}
			return masterKey, ErrUnknown.Reformat(errString)
		}
	}
	masterKey = strings.TrimSpace(stdOutput.String())
	return masterKey, nil
}

// GocryptfsResetVaultPassword reset password for vault using masterkey.
func (m *VaultManager) GocryptfsResetVaultPassword(path string, masterkey string, newPassword string) error {
	chPwProc := exec.Command(m.cmd, "-passwd", "-masterkey", masterkey, "--", path)
	// Password is piped through STDIN
	stdIn, err := chPwProc.StdinPipe()
	var errorOutput bytes.Buffer
	chPwProc.Stderr = &errorOutput
	if err != nil {
		logger.Error().Err(err).
			Str("vaultPath", path).
			Msg("Failed to create STDIN pipe when recovering password for vault")
		return err
	}

	go func() {
		defer stdIn.Close()
		passwords := strings.Join([]string{newPassword, newPassword}, "\n")
		if _, err := io.WriteString(stdIn, passwords); err != nil {
			logger.Error().Err(err).
				Str("vaultPath", path).
				Msg("Failed to pipe passwords to gocryptfs when recovering password for vault")
		}
	}()

	if err := chPwProc.Run(); err != nil {
		rc := chPwProc.ProcessState.ExitCode()
		errString := errorOutput.String()
		errLog := logger.With().
			Err(err).
			Int("RC", rc).
			Str("vaultPath", path).
			Str("stdErr", errString).
			Logger()
		switch rc {
		case 12:
			errLog.Error().Msg("Password incorrect")
			return ErrWrongPassword
		case 23:
			errLog.Error().Msg("Gocryptfs could not open gocryptfs.conf for reading")
			return ErrCantOpenVaultConf
		case 24:
			errLog.Error().Msg("Gocryptfs could not write the updated gocryptfs.conf")
			return ErrVaultUpdateConfFailed
		default:
			errLog.Error().Msg("Unknown error when recovering password for vault")
			return ErrUnknown.Reformat(errString)
		}
	}

	// Rename backup of original vault config file
	// so the next time user do a password resetting there is no file conflicting
	vaultConfBackup := filepath.Join(path, "gocryptfs.conf.bak")
	vaultConfBackupWithTime := filepath.Join(
		path,
		fmt.Sprintf(
			"gocryptfs.conf.bak.%s", time.Now().UTC().Format("2006-01-02T15:04:05.000 MST"),
		),
	)
	if err := os.Rename(vaultConfBackup, vaultConfBackupWithTime); err != nil && !os.IsNotExist(err) {
		logger.Warn().Err(err).
			Str("vaultPath", path).
			Str("from", vaultConfBackup).
			Str("to", vaultConfBackupWithTime).
			Msg("Failed to rename backup of original gocryptfs.conf file")
	}
	return nil
}

// GocryptfsUnlockVault unlocks the vault identified by `vaultId` using given `password`.
func (m *VaultManager) GocryptfsUnlockVault(vaultId int64, password string) error {
	// Check current state
	if _, ok := m.mountPoints[vaultId]; ok {
		return ErrVaultAlreadyUnlocked
	}

	// Lock internal maps
	m.lock.Lock()
	defer m.lock.Unlock()

	// Locate vault in repository
	vault, err := m.repo.Get(vaultId, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrVaultNotExist
		}
		return err
	}
	// Locate a mountpoint for this vault
	if strings.TrimSpace(vault.MountPoint) == "" {
		var mountPointBase string
		if runtime.GOOS == "darwin" {
			mountPointBase = "/Volumes"
		} else {
			mountPointBase = os.TempDir()
		}
		vault.MountPoint = filepath.Join(mountPointBase, strconv.FormatInt(int64(rand.Int31()), 16))
	}
	// OSXFUSE will create mountpoint for us if it's located in `/Volumes`,
	// but for Linux we'll have to do the mkdir ourselves.
	shouldRemoveMountpoint := false
	if runtime.GOOS != "darwin" {
		if err := os.MkdirAll(vault.MountPoint, 0700); err != nil {
			logger.Error().Err(err).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Failed to create mountpoint directory")
			return ErrMountpointMkdirFailed
		}
		shouldRemoveMountpoint = true
	}
	// Start a gocryptfs process to unlock this vault
	args := []string{"-fg"}
	if runtime.GOOS == "darwin" {
		args = append(
			args,
			"-ko",
			fmt.Sprintf("volname=%s,local,auto_xattr,noappledouble", filepath.Base(vault.Path)),
		)
	}
	// Readonly mode
	if vault.ReadOnly {
		args = append(args, "-ro")
		logger.Debug().
			Str("vaultPath", vault.Path).
			Str("mountPoint", vault.MountPoint).
			Bool("readOnly", vault.ReadOnly).
			Msg("Vault is set to mount Read-Only")
	}
	args = append(args, "--", vault.Path, vault.MountPoint)
	m.processes[vaultId] = exec.Command(m.cmd, args...)
	m.mountPoints[vaultId] = vault.MountPoint

	// Password is piped through STDIN
	stdIn, err := m.processes[vaultId].StdinPipe()
	if err != nil {
		logger.Error().Err(err).
			Str("vaultPath", vault.Path).
			Str("mountPoint", vault.MountPoint).
			Msg("Failed to create STDIN pipe")
		defer delete(m.processes, vaultId)
		defer delete(m.mountPoints, vaultId)
		return err
	}

	rcPipe := make(chan int)

	go func() {
		defer stdIn.Close()
		if _, err := io.WriteString(stdIn, password); err != nil {
			logger.Error().Err(err).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Failed to pipe vault password to gocryptfs")
		}
	}()
	if err := m.processes[vaultId].Start(); err != nil {
		logger.Error().Err(err).
			Int64("vaultId", vaultId).
			Str("vaultPath", vault.Path).
			Str("mountPoint", vault.MountPoint).
			Str("gocryptfs", m.cmd).
			Msg("Failed to start gocryptfs process")

		// Cleanup immediately
		defer delete(m.processes, vaultId)
		defer delete(m.mountPoints, vaultId)
		defer close(rcPipe)

		return err
	}

	// Need to wait for this process to exit, otherwise it becomes zombie after exiting.
	go func() {
		proc := m.processes[vaultId]
		if err := proc.Wait(); err != nil {
			rc := proc.ProcessState.ExitCode()
			switch rc {
			case 10, 12, 23: // These are known errors meant to be reported directly to the UI
				break
			case 15: // gocryptfs interrupted by SIGINT, a.k.a. we locked this vault
				logger.Info().
					Int("RC", rc).
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", vault.MountPoint).
					Msg("Vault locked")
				break
			default:
				logger.Error().Err(err).
					Int("RC", rc).
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", vault.MountPoint).
					Msg("Gocryptfs exited unexpectedly")
			}
			rcPipe <- rc
		} else {
			logger.Debug().
				Int64("vaultId", vaultId).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Gocryptfs exited without error, the mountpoint was probably unmounted manually")
			rcPipe <- 0
		}

		// Cleanup
		m.lock.Lock()
		defer m.lock.Unlock()
		defer delete(m.processes, vaultId)
		defer delete(m.mountPoints, vaultId)
		defer close(rcPipe)

		// Remove mountpoint directory if we created it
		if shouldRemoveMountpoint {
			if err := os.Remove(vault.MountPoint); err != nil {
				logger.Error().Err(err).
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", vault.MountPoint).
					Msg("Failed to remove mountpoint directory")
			}
		}
	}()

	// Wait for a little time, then check if gocryptfs process is alive
	// If it exited, there's something wrong, respond to the UI
	timer := time.NewTimer(time.Second)
	select {
	case rc := <-rcPipe:
		switch rc {
		case 10: // Mountpoint not empty
			logger.Error().Err(err).
				Int("RC", rc).
				Int64("vaultId", vaultId).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Mountpoint not empty")
			return ErrMountpointNotEmpty.WrapState("locked")
		case 12: // Incorrect password
			logger.Error().Err(err).
				Int("RC", rc).
				Int64("vaultId", vaultId).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Vault password incorrect")
			return ErrWrongPassword.WrapState("locked")
		case 23: // gocryptfs.conf IO error
			logger.Error().Err(err).
				Int("RC", rc).
				Int64("vaultId", vaultId).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Gocryptfs cannot open gocryptfs.conf")
			return ErrCantOpenVaultConf.WrapState("locked")
		default:
			logger.Error().Err(err).
				Int("RC", rc).
				Int64("vaultId", vaultId).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Gocryptfs exited unexpectedly")
			return ErrUnknown.Reformat(rc).WrapState("locked")
		}
	case <-timer.C:
		logger.Debug().
			Int64("vaultId", vaultId).
			Str("vaultPath", vault.Path).
			Str("mountPoint", vault.MountPoint).
			Msg("Vault unlocked")
		// Read from rcPipe, otherwise the `Wait` goroutine will block after gocryptfs exited
		go func() {
			<-rcPipe
		}()
	}

	if vault.AutoReveal {
		go func() {
			start := time.Now()
			ticker := time.NewTicker(time.Millisecond * 100)
			timeout := time.NewTimer(time.Second * 5)
			defer ticker.Stop()
			defer timeout.Stop()

			for {
				select {
				case <-ticker.C:
					info, err := os.Stat(vault.MountPoint)
					if err == nil && info.IsDir() {
						logger.Debug().
							Str("vaultPath", vault.Path).
							Str("mountPoint", vault.MountPoint).
							Bool("autoReveal", vault.AutoReveal).
							Dur("waited", time.Since(start)).
							Msg("Auto revealing mountpoint")
						extension.OpenPath(vault.MountPoint)
						return
					}
					logger.Debug().
						Str("vaultPath", vault.Path).
						Str("mountPoint", vault.MountPoint).
						Bool("autoReveal", vault.AutoReveal).
						Dur("waited", time.Since(start)).
						Msg("Mountpoint not ready, still waiting")
				case <-timeout.C:
					logger.Warn().
						Str("vaultPath", vault.Path).
						Str("mountPoint", vault.MountPoint).
						Bool("autoReveal", vault.AutoReveal).
						Dur("waited", time.Since(start)).
						Msg("Auto revealing timed out")
					return
				}
			}
		}()

		logger.Debug().
			Str("vaultPath", vault.Path).
			Str("mountPoint", vault.MountPoint).
			Bool("autoReveal", vault.AutoReveal).
			Msg("Waiting for mountpoint to appear before auto revealing")
	}

	return nil
}
