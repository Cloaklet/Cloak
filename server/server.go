package server

import (
	"Cloak/extension"
	"Cloak/i18n"
	"Cloak/models"
	"Cloak/version"
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/gommon/random"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/xattr"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	logger = extension.GetLogger("server")
}

// ApiServer represents a type which:
// - Communicates with frontend (the UI);
// - Controls and reacts to gocryptfs processes;
// - Maintains the vault database;
type ApiServer struct {
	VaultManager
	echo  *echo.Echo // the actual HTTP server
	token string
}

// NewApiServer creates a new ApiServer instance
// - repo passes in the vault repository to persist vault list data
func NewApiServer(repo *models.VaultRepo, releaseMode bool, configCh chan map[string]string) *ApiServer {
	// Create server
	server := ApiServer{
		echo: echo.New(),
		VaultManager: VaultManager{
			repo:          repo,
			fuseAvailable: extension.IsFuseAvailable(),
			processes:     make(map[int64]*exec.Cmd),
			mountPoints:   make(map[int64]string),
			configCh:      configCh,
		},
		// Generate a random token on startup, for API access
		token: random.String(64),
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
		embeddedFs, err := fs.Sub(frontend, "dist")
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to mount subdirectory of embedded FS to HTTP router")
		}
		server.echo.GET("/*", echo.WrapHandler(http.FileServer(http.FS(embeddedFs))))
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
	apis.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check token
			if c.Request().Header.Get("Authorization") != fmt.Sprintf("Bearer %s", server.token) {
				return ErrUnauthorized
			}
			return next(c)
		}
	})
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
		apis.POST("/options", server.SetOptions)
	}

	return &server
}

func (s *ApiServer) GetAccessUrl() string {
	return fmt.Sprintf("http://127.0.0.1:9763/#token=%s", s.token)
}

// Start starts the server
func (s *ApiServer) Start(address string) error {
	return s.echo.Start(address)
}

// Stop stops the server
// All vaults will be locked before the server stops.
func (s *ApiServer) Stop() error {
	logger.Debug().Msg("Requested to stop API server")

	// Lock internal maps
	s.lock.Lock()
	defer s.lock.Unlock()

	// Lock all unlocked vaults
	for vaultId, mountPoint := range s.mountPoints {
		logger.Debug().
			Int64("vaultId", vaultId).
			Str("mountPoint", mountPoint).
			Msg("Locking vault")

		// stop corresponding gocryptfs process to lock this vault
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

// VaultInfo represents a single vault along with its current state
type VaultInfo struct {
	models.Vault
	State string `json:"state"` // Legal values: locked/unlocked
}

// ListVaults returns a list of all known vaults
func (s *ApiServer) ListVaults(_ echo.Context) error {
	var (
		vaults []models.Vault
		err    error
	)
	if vaults, err = s.repo.List(nil); err != nil {
		return ErrListFailed
	}

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

	switch form.Op {
	case "unlock":
		if err := s.GocryptfsUnlockVault(vaultId, form.Password); err != nil {
			return err
		}
		return ErrOk.WrapState("unlocked")
	case "lock":
		// Check current state
		if _, ok := s.mountPoints[vaultId]; !ok {
			return ErrVaultAlreadyLocked.WrapState("locked")
		}

		// Lock internal maps
		s.lock.Lock()
		defer s.lock.Unlock()
		// stop corresponding gocryptfs process to lock this vault
		if err := s.processes[vaultId].Process.Signal(os.Interrupt); err != nil {
			return err
		}
		// We have a pairing gorountine to wait for gocryptfs process to exit and do the cleanup,
		// so no need to cleaning `s.processes` and `s.mountPoints` here.
		// Respond
		return ErrOk.WrapState("locked")
	case "reveal_mountpoint":
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
	case "reveal_vault":
		if vault, err := s.repo.Get(vaultId, nil); err == nil {
			if _, err := os.Stat(vault.Path); err != nil {
				return ErrPathNotExist
			} else {
				extension.OpenPath(vault.Path)
				return ErrOk
			}
		} else {
			return err
		}
	default:
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
		Password    string `json:"password"`  // optional, either `password` or `masterkey` will do
		MasterKey   string `json:"masterkey"` // optional, either `password` or `masterkey` will do
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
	if form.Password != "" {
		if err = s.GocryptfsChangeVaultPassword(vault.Path, form.Password, form.NewPassword); err != nil {
			return err
		}
	} else if form.MasterKey != "" {
		if err = s.GocryptfsResetVaultPassword(vault.Path, form.MasterKey, form.NewPassword); err != nil {
			return err
		}
	} else {
		return ErrMalformedInput
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
	// stop corresponding gocryptfs process to lock this vault
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

	switch form.Op {
	case "add":
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
	case "create":
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
	default:
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
			xAttrs, err := xattr.Get(filepath.Join(form.Pwd, item.Name()), "com.apple.FinderInfo")
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
			if len(xAttrs) == 32 && xAttrs[8] > 40 {
				logger.Debug().
					Bytes("com.apple.FinderInfo", xAttrs).
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
		"options": echo.Map{
			"locale":   i18n.GetLocalizer().GetCurrentLocale(),
			"loglevel": strings.ToUpper(zerolog.GlobalLevel().String()),
		},
	})
}

// SetOptions persists application options.
// Currently the only available option is `locale`.
func (s *ApiServer) SetOptions(c echo.Context) error {
	var appOption struct {
		Locale   string `json:"locale"`
		LogLevel string `json:"loglevel"`
	}
	if err := c.Bind(&appOption); err != nil {
		return ErrMalformedInput
	}

	if appOption.Locale != "" && appOption.Locale != i18n.GetLocalizer().GetCurrentLocale() {
		s.configCh <- map[string]string{"locale": appOption.Locale}
	}

	if appOption.LogLevel != "" && appOption.LogLevel != zerolog.GlobalLevel().String() {
		s.configCh <- map[string]string{"loglevel": appOption.LogLevel}
	}
	return ErrOk
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
