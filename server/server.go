package server

import (
	"Cloak/extension"
	"Cloak/models"
	_ "Cloak/statik"
	"Cloak/version"
	"context"
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/xattr"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logger zerolog.Logger

func init() {
	logger = extension.GetLogger("server")
}

type ApiServer struct {
	repo          *models.VaultRepo   // database repository
	echo          *echo.Echo          // the actual HTTP server
	cmd           string              // `gocryptfs` binary path
	xrayCmd       string              // `gocryptfs-xray` binary path
	fuseAvailable bool                // whether FUSE is available
	processes     map[int64]*exec.Cmd // vaultID: process
	mountPoints   map[int64]string    // vaultID: mountPoint
	lock          sync.Mutex          // lock on `processes` and `mountPoints`
}

// Start starts the server
func (s *ApiServer) Start(address string) error {
	return s.echo.Start(address)
}

// Stop stops the server
// All vaults will be locked before the server stops.
func (s *ApiServer) Stop() error {
	logger.Debug().Msg("Requestd to stop API server")

	// Lock internal maps
	s.lock.Lock()
	defer s.lock.Unlock()

	// Lock all unlocked vaults
	for vaultId, mountPoint := range s.mountPoints {
		logger.Debug().
			Int64("vaultId", vaultId).
			Str("mountPoint", mountPoint).
			Msg("Locking vault")

		// Stop corresponding gocryptfs process to lock this vault
		if err := s.processes[vaultId].Process.Signal(os.Interrupt); err != nil {
			logger.Error().Err(err).
				Int64("vaultId", vaultId).
				Str("mountPoint", mountPoint).
				Msg("Failed to stop gocryptfs process with SIGINT")
		}
	}

	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	return s.echo.Shutdown(ctx)
}

// NewApiServer creates a new ApiServer instance
// - repo passes in the vault repository to persist vault list data
func NewApiServer(repo *models.VaultRepo, releaseMode bool) *ApiServer {
	// Create server
	server := ApiServer{
		repo:          repo,
		echo:          echo.New(),
		cmd:           "",
		xrayCmd:       "",
		fuseAvailable: extension.IsFuseAvailable(),
		processes:     map[int64]*exec.Cmd{},
		mountPoints:   map[int64]string{},
	}

	// Detect external runtime dependencies
	var err error
	if server.cmd, err = extension.LocateBinary("gocryptfs"); err != nil {
		logger.Error().Err(err).
			Msg("Failed to locate gocryptfs binary, nothing will work")
	} else {
		logger.Debug().Str("gocryptfs", server.cmd).Msg("Gocryptfs binary located")
	}
	if server.xrayCmd, err = extension.LocateBinary("gocryptfs-xray"); err != nil {
		logger.Error().Err(err).
			Msg("Failed to locate gocryptfs-xray binary, masterkey revealing will not work")
	} else {
		logger.Debug().Str("gocryptfs-xray", server.xrayCmd).Msg("gocryptfs-xray binary located")
	}
	logger.Debug().Bool("fuseAvailable", server.fuseAvailable).Msg("FUSE detection finished")

	// Setup HTTP server
	server.echo.HideBanner = true
	server.echo.HidePort = true

	// Load files from disk when we're not built for release
	if !releaseMode {
		logger.Info().Msg("Running in DEV mode")
		server.echo.Static("/", "./frontend/dist")
	} else { // Load files from embedded FS when in release mode
		logger.Debug().Msg("Running in RELEASE mode")
		embedFs, err := fs.New()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to init FS from embedded assets")
		}
		server.echo.GET("/*", echo.WrapHandler(http.FileServer(embedFs)))
	}

	// Use a custom error handler to produce unified JSON responses.
	server.echo.HTTPErrorHandler = func(err error, c echo.Context) {
		switch typedErr := err.(type) {
		case *ApiError, *DataContainer:
			c.JSON(http.StatusOK, typedErr)
			return
		default:
			c.JSON(http.StatusInternalServerError, ErrUnknown.Reformat(err))
		}
	}
	/**
	APIs are located at /api:
	- GET /vaults: get a list of all known vaults
	- POST /vaults: create or add a vault
	  - op=create: create a new vault
	  - op=add: add an existing vault to Cloak app
	- POST /vault/N: operate on a vault
	  - op=update: update vault information
	  - op=unlock: unlock a vault, pass password with `pw`
	  - op=lock: lock a vault
	  - op=reveal: reveal mountpoint in file manager, only available if vault is unlocked
	- DELETE /vault/N: delete a vault from Cloak. Files are reserved on disk.
	*/
	if !releaseMode {
		logger.Warn().
			Bool("releaseMode", releaseMode).
			Msg("Running in DEV mode, CORS enabled")
		server.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{
				http.MethodOptions,
				http.MethodGet,
				http.MethodHead,
				http.MethodPut,
				http.MethodPatch,
				http.MethodPost,
				http.MethodDelete,
			},
		}))
	}
	apis := server.echo.Group("/api", server.CheckRuntimeDeps)

	{
		apis.GET("/vaults", server.ListVaults)
		apis.DELETE("/vault/:id", server.RemoveVault)
		apis.POST("/vaults", server.AddOrCreateVault)
		// Unlock a vault / Lock a vault / reveal mountpoint for an unlocked vault
		apis.POST("/vault/:id", server.OperateOnVault)
		// Update vault options (autoreveal / readonly)
		apis.POST("/vault/:id/options", server.UpdateVaultOptions)
		// Change vault password
		apis.POST("/vault/:id/password", server.ChangeVaultPassword)
		// Reveal vault masterkey
		apis.POST("/vault/:id/masterkey", server.RevealVaultMasterkey)
		// List local disk content
		apis.POST("/subpaths", server.ListSubPaths)
		apis.GET("/options", server.GetOptions)
	}
	// We use `rand` to generate random mountpoint name, so be sure to seed it upon start up
	rand.Seed(time.Now().UTC().UnixNano())
	return &server
}

// VaultInfo represents a single vault along with its current state
type VaultInfo struct {
	models.Vault
	State string `json:"state"` // Legal values: locked/unlocked
}

// ListVaults returns a list of all known vaults
func (s *ApiServer) ListVaults(_ echo.Context) error {
	if vaults, err := s.repo.List(nil); err != nil {
		return ErrListFailed
	} else {
		vaultList := make([]VaultInfo, len(vaults))
		for i, v := range vaults {
			vaultList[i] = VaultInfo{Vault: v, State: "locked"}
			// Detect vault state
			if _, ok := s.mountPoints[v.ID]; ok {
				vaultList[i].State = "unlocked"
			}
		}
		return ErrOk.WrapList(vaultList)
	}
}

// OperateOnVault performs operations on a single vault, including lock/unlock.
// - `op` identifies the operation to perform
func (s *ApiServer) OperateOnVault(c echo.Context) error {
	// Pre-check on ID
	vaultId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrMalformedInput
	}

	var form struct {
		Op       string `json:"op"`       // lock/unlock/reveal/update
		Password string `json:"password"` // for `unlock` op only
	}
	if err := c.Bind(&form); err != nil {
		return ErrMalformedInput
	}

	if form.Op == "unlock" {
		// Check current state
		if _, ok := s.mountPoints[vaultId]; ok {
			return ErrVaultAlreadyUnlocked
		}

		// Lock internal maps
		s.lock.Lock()
		defer s.lock.Unlock()
		// Locate vault in repository
		vault, err := s.repo.Get(vaultId, nil)
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
		// Start a gocryptfs process to unlock this vault
		args := []string{"-fg"}
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
		s.processes[vaultId] = exec.Command(s.cmd, args...)
		s.mountPoints[vaultId] = vault.MountPoint

		// Password is piped through STDIN
		stdIn, err := s.processes[vaultId].StdinPipe()
		if err != nil {
			logger.Error().Err(err).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Msg("Failed to create STDIN pipe")
			defer delete(s.processes, vaultId)
			defer delete(s.mountPoints, vaultId)
			return err
		}

		go func() {
			defer stdIn.Close()
			if _, err := io.WriteString(stdIn, form.Password); err != nil {
				logger.Error().Err(err).
					Str("vaultPath", vault.Path).
					Str("mountPoint", vault.MountPoint).
					Msg("Failed to pipe vault password to gocryptfs")
			}
		}()
		if err := s.processes[vaultId].Start(); err != nil {
			logger.Error().Err(err).
				Int64("vaultId", vaultId).
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Str("gocryptfs", s.cmd).
				Msg("Failed to start gocryptfs process")

			// Cleanup immediately
			defer delete(s.processes, vaultId)
			defer delete(s.mountPoints, vaultId)

			return err
		}

		// Need to wait for this process to exit, otherwise it becomes zombie after exiting.
		go func() {
			proc := s.processes[vaultId]
			err := proc.Wait()
			s.lock.Lock()
			defer s.lock.Unlock()
			defer delete(s.processes, vaultId)
			defer delete(s.mountPoints, vaultId)

			if err != nil {
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
			} else {
				logger.Debug().
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", vault.MountPoint).
					Msg("Gocryptfs exited without error, the mountpoint was probably unmounted manually")
			}
		}()

		// TODO How to improve this?
		// Wait for a little time, then check if gocryptfs process is alive
		// If it exited, there's something wrong, respond to the UI
		time.Sleep(time.Second * 1)
		if s.processes[vaultId].ProcessState != nil {
			defer delete(s.processes, vaultId)
			defer delete(s.mountPoints, vaultId)

			rc := s.processes[vaultId].ProcessState.ExitCode()
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
		}

		if vault.AutoReveal {
			logger.Debug().
				Str("vaultPath", vault.Path).
				Str("mountPoint", vault.MountPoint).
				Bool("autoReveal", vault.AutoReveal).
				Msg("Auto revealing mountpoint")
			go extension.OpenPath(vault.MountPoint)
		}
		// Respond
		logger.Debug().
			Int64("vaultId", vaultId).
			Str("vaultPath", vault.Path).
			Str("mountPoint", vault.MountPoint).
			Msg("Vault unlocked")
		return ErrOk.WrapState("unlocked")
	} else if form.Op == "lock" {
		// Check current state
		if _, ok := s.mountPoints[vaultId]; !ok {
			return ErrVaultAlreadyLocked.WrapState("locked")
		}

		// Lock internal maps
		s.lock.Lock()
		defer s.lock.Unlock()
		// Stop corresponding gocryptfs process to lock this vault
		if err := s.processes[vaultId].Process.Signal(os.Interrupt); err != nil {
			return err
		}
		// We have a pairing gorountine to wait for gocryptfs process to exit and do the cleanup,
		// so no need to cleaning `s.processes` and `s.mountPoints` here.
		// Respond
		return ErrOk.WrapState("locked")
	} else if form.Op == "reveal" {
		var mountPoint string
		var ok bool
		// Check current state
		if mountPoint, ok = s.mountPoints[vaultId]; !ok {
			return ErrVaultAlreadyLocked
		}

		// Check mountpoint path existence
		if pathInfo, err := os.Stat(mountPoint); err != nil || !pathInfo.IsDir() {
			return ErrPathNotExist
		}

		extension.OpenPath(mountPoint)
		return ErrOk
	} else {
		// Currently not supported
		return ErrUnsupportedOperation
	}
}

// UpdateVaultOptions updates options for given vault
func (s *ApiServer) UpdateVaultOptions(c echo.Context) error {
	// Pre-check on ID
	vaultId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrMalformedInput
	}

	var form struct {
		AutoReveal bool   `json:"autoreveal"`
		ReadOnly   bool   `json:"readonly"`
		Mountpoint string `json:"mountpoint"`
	}
	if err := c.Bind(&form); err != nil {
		return ErrMalformedInput
	}

	// Lock internal maps
	s.lock.Lock()
	defer s.lock.Unlock()

	// Locate vault in repository
	vault, err := s.repo.Get(vaultId, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrVaultNotExist
		}
		return err
	}

	// Vault must be locked to update its settings
	if _, ok := s.mountPoints[vaultId]; ok {
		return ErrVaultAlreadyUnlocked.WrapItem(VaultInfo{
			Vault: vault,
			State: "unlocked",
		})
	}

	if err := s.repo.WithTransaction(func(tx models.Transactional) error {
		vault.AutoReveal = form.AutoReveal
		vault.ReadOnly = form.ReadOnly
		vault.MountPoint = strings.TrimSpace(form.Mountpoint)
		return s.repo.Update(&vault, tx)
	}); err != nil {
		return err
	}
	return ErrOk.WrapItem(VaultInfo{State: "locked", Vault: vault})
}

// ChangeVaultPassword changes password for given vault
func (s *ApiServer) ChangeVaultPassword(c echo.Context) error {
	// Pre-check on ID
	vaultId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrMalformedInput
	}

	var form struct {
		Password    string `json:"password"`
		NewPassword string `json:"newpassword"`
	}
	if err := c.Bind(&form); err != nil {
		return ErrMalformedInput
	}

	// Lock internal maps
	s.lock.Lock()
	defer s.lock.Unlock()

	// Locate vault in repository
	vault, err := s.repo.Get(vaultId, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrVaultNotExist
		}
		return err
	}

	// Vault must be locked to update its settings
	if _, ok := s.mountPoints[vaultId]; ok {
		return ErrVaultAlreadyUnlocked.WrapItem(VaultInfo{
			Vault: vault,
			State: "unlocked",
		})
	}

	// Start a gocryptfs process to change password
	if err = s.GocryptfsChangeVaultPassword(vault.Path, form.Password, form.NewPassword); err != nil {
		return err
	}
	return ErrOk
}

// RemoveVault removes vault specified by ID from database repository,
// the corresponding directory remains on the disk.
func (s *ApiServer) RemoveVault(c echo.Context) error {
	// Pre-check on ID
	vaultId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrMalformedInput
	}

	// Lock internal maps
	s.lock.Lock()
	defer s.lock.Unlock()
	// Lock vault if necessary
	// Stop corresponding gocryptfs process to lock this vault
	if _, ok := s.mountPoints[vaultId]; ok {
		if err := s.processes[vaultId].Process.Signal(os.Interrupt); err != nil {
			return err
		}
	}

	var vault models.Vault
	err = s.repo.WithTransaction(func(tx models.Transactional) error {
		vault, err = s.repo.Get(vaultId, tx)
		if err != nil {
			// Vault ID not found
			if err == sql.ErrNoRows {
				return ErrVaultNotExist
			}
			return err
		}
		return s.repo.Delete(&vault, tx)
	})
	if err != nil {
		return err
	}
	// Deleted
	logger.Debug().
		Int64("vaultId", vaultId).
		Str("vaultPath", vault.Path).
		Msg("Vault removed. It is no longer managed by Cloak.")
	return ErrOk
}

// AddOrCreateVault adds an existing gocryptfs vault to the repository,
// Or it creates a new vault at specified location (not currently supported).
// When adding an existing vault `path` will be the absolute path of `gocryptfs.conf`;
// When creating a new vault `path` will be the parent directory of the new vault.
func (s *ApiServer) AddOrCreateVault(c echo.Context) error {
	var form struct {
		Op       string `json:"op"` // add/create
		Path     string `json:"path"`
		Name     string `json:"name"`     // optional, only when op=create
		Password string `json:"password"` // optional, only when op=create
	}
	if err := c.Bind(&form); err != nil {
		return ErrMalformedInput
	}

	if form.Op == "add" {
		// Check path existence
		if pathInfo, err := os.Stat(form.Path); err != nil || pathInfo.IsDir() {
			return ErrPathNotExist
		}
		vaultPath := filepath.Dir(form.Path)

		var err error
		var vault models.Vault
		err = s.repo.WithTransaction(func(tx models.Transactional) error {
			vault, err = s.repo.Create(echo.Map{"path": vaultPath}, tx)
			if err != nil {
				logger.Error().Err(err).
					Str("vaultPath", vaultPath).
					Msg("Failed to add existing vault")
			}
			return err
		})
		if err != nil {
			return err
		}
		logger.Debug().
			Str("vaultPath", vaultPath).
			Int64("vaultId", vault.ID).
			Msg("Added existing vault")
		return ErrOk.WrapItem(VaultInfo{
			Vault: vault,
			State: "locked",
		})

	} else if form.Op == "create" {
		// Check path existence
		if pathInfo, err := os.Stat(form.Path); err != nil || !pathInfo.IsDir() {
			return ErrPathNotExist
		}
		vaultPath := filepath.Join(form.Path, form.Name)
		if err := os.Mkdir(vaultPath, 0700); err != nil {
			logger.Error().Err(err).
				Str("vaultDirectroy", form.Path).
				Str("vaultName", form.Name).
				Msg("Failed to create vault directory")
			return ErrVaultMkdirFailed.Reformat(err)
		}

		err := s.GocryptfsCreateVault(vaultPath, form.Password)
		if err != nil {
			return err
		}

		// Vault created, add to vault repository
		var vault models.Vault
		if err := s.repo.WithTransaction(func(tx models.Transactional) error {
			vault, err = s.repo.Create(echo.Map{"path": vaultPath}, tx)
			if err != nil {
				logger.Error().Err(err).
					Str("vaultPath", vaultPath).
					Msg("Failed to add newly created vault")
			}
			return err
		}); err != nil {
			return err
		}
		logger.Debug().
			Str("vaultPath", vaultPath).
			Int64("vaultId", vault.ID).
			Msg("Added newly created vault")
		return ErrOk.WrapItem(VaultInfo{Vault: vault, State: "locked"})

	} else {
		// Currently not supported
		return ErrUnsupportedOperation
	}
}

// ListSubPaths lists items in given path.
// - `pwd` identifies the path to use, the special value `$HOME` translates to home directory of current user.
func (s *ApiServer) ListSubPaths(c echo.Context) error {
	var form struct {
		Pwd string `json:"pwd"`
	}
	if err := c.Bind(&form); err != nil {
		return ErrMalformedInput
	}

	// Locate user home directory
	if form.Pwd == `$HOME` {
		currentUser, err := user.Current()
		if err != nil {
			return err
		}
		form.Pwd = currentUser.HomeDir
	}
	// Normalize path
	form.Pwd = filepath.Clean(form.Pwd)

	// List items
	items, err := ioutil.ReadDir(form.Pwd)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrPathNotExist
		}
		return err
	}

	type Item struct {
		Name string `json:"name"`
		Type string `json:"type"` // file/directory
	}
	subPathItems := make([]Item, 0, len(items))
	for _, item := range items {
		var subItem Item
		subItem.Name = item.Name()
		// Skip hidden items
		if subItem.Name[0] == '.' {
			continue
		}

		// Skip symbolic links
		if item.Mode()&os.ModeSymlink != 0 {
			continue
		}

		// Skip items that aren't visible in Finder app
		if runtime.GOOS == "darwin" {
			xattrs, err := xattr.Get(filepath.Join(form.Pwd, item.Name()), "com.apple.FinderInfo")
			if err != nil {
				// No attribute is ok, other errors need to be logged
				if errno, ok := err.(*xattr.Error); !ok || errno.Err != xattr.ENOATTR {
					logger.Warn().Err(err).
						Str("pwd", form.Pwd).
						Str("fileName", subItem.Name).
						Msg("Failed to get extended attributes")
				}
			}
			// I have no idea how to match this piece of data, so this is based on some samples I observed.
			// Some references:
			// https://discussions.apple.com/thread/5846108
			// http://dubeiko.com/development/FileSystems/HFSPLUS/tn1150.html#FinderInfo
			// https://apple.stackexchange.com/a/174571
			if len(xattrs) == 32 && xattrs[8] > 40 {
				logger.Debug().
					Bytes("com.apple.FinderInfo", xattrs).
					Str("pwd", form.Pwd).
					Str("fileName", subItem.Name).
					Msg("Item skipped")
				continue
			}
		}

		if item.IsDir() {
			subItem.Type = "directory"
		} else {
			// We are not interested in files other than `gocryptfs.conf`
			if subItem.Name != "gocryptfs.conf" {
				continue
			}
			subItem.Type = "file"
		}
		subPathItems = append(subPathItems, subItem)
	}

	// Respond
	// TODO Improve
	return c.JSON(http.StatusOK, echo.Map{
		"code":  ErrOk.Code,
		"msg":   ErrOk.Message,
		"sep":   string(filepath.Separator),
		"pwd":   form.Pwd,
		"items": subPathItems,
	})
}

// GetOptions returns app options.
// Currently it only returns version info.
func (s *ApiServer) GetOptions(c echo.Context) error {
	return ErrOk.WrapItem(echo.Map{
		"version": echo.Map{
			"version":   version.Version,
			"buildTime": version.BuildTime,
			"gitCommit": version.GitCommit,
		},
	})
}

// RevealVaultMasterkey returns masterkey for given vault.
func (s *ApiServer) RevealVaultMasterkey(c echo.Context) error {
	if s.xrayCmd == "" {
		return ErrMissingGocryptfsXrayBinary
	}

	// Pre-check on ID
	vaultId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrMalformedInput
	}

	var form struct {
		Password string `json:"password"`
	}
	if err := c.Bind(&form); err != nil {
		return ErrMalformedInput
	} else if form.Password == "" {
		return ErrMalformedInput
	}

	vault, err := s.repo.Get(vaultId, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrVaultNotExist
		}
		return err
	}

	masterKey, err := s.GocryptfsShowVaultMasterkey(vault.Path, form.Password)
	if err != nil {
		return err
	}

	return ErrOk.WrapItem(masterKey)
}
