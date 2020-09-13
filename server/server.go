package server

import (
	"Cloak/extension"
	"Cloak/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/xattr"
	"github.com/rs/zerolog"
	"html/template"
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

const (
	ERR_CODE_OK                       = 0
	ERR_MSG_OK                        = "Ok"
	ERR_CODE_LIST_FAILED              = 1
	ERR_MSG_LIST_FAILED               = "Failed to list vaults"
	ERR_CODE_MALFORMED_INPUT          = 2
	ERR_MSG_MALFORMED_INPUT           = "Malformed input data"
	ERR_CODE_UNKNOWN                  = 3
	ERR_MSG_UNKNOWN                   = "Error: %v"
	ERR_CODE_PATH_NOT_EXISTS          = 4
	ERR_MSG_PATH_NOT_EXISTS           = "Given path does not exist"
	ERR_CODE_UNSUPPORTED_OPERATION    = 5
	ERR_MSG_UNSUPPORTED_OPERATION     = "Unsupported operation"
	ERR_CODE_VAULT_NOT_EXISTS         = 6
	ERR_MSG_VAULT_NOT_EXISTS          = "Given vault ID does not exist"
	ERR_CODE_ALREADY_UNLOCKED         = 7
	ERR_MSG_ALREADY_UNLOCKED          = "This vault is already unlocked"
	ERR_CODE_ALREADY_LOCKED           = 8
	ERR_MSG_ALREADY_LOCKED            = "This vault is already locked"
	ERR_CODE_MOUNTPOINT_NOT_EMPTY     = 9
	ERR_MSG_MOUNTPOINT_NOT_EMPTY      = "Mountpoint is not empty"
	ERR_CODE_WRONG_PASSWORD           = 10
	ERR_MSG_WRONG_PASSWORD            = "Password incorrect"
	ERR_CODE_CANT_OPEN_VAULT_CONF     = 11
	ERR_MSG_CANT_OPEN_VAULT_CONF      = "gocryptfs.conf could not be opened"
	ERR_CODE_MISSING_GOCRYPTFS_BINARY = 12
	ERR_MSG_MISSING_GOCRYPTFS_BINARY  = "Cannot locate gocryptfs binary"
	ERR_CODE_MISSING_FUSE             = 13
	ERR_MSG_MISSING_FUSE              = "FUSE is not available on this computer"
)

var logger zerolog.Logger

func init() {
	logger = extension.GetLogger("server")
}

type ApiServer struct {
	repo          *models.VaultRepo   // database repository
	echo          *echo.Echo          // the actual HTTP server
	cmd           string              // `gocryptfs` binary path
	fuseAvailable bool                // whether FUSE is available
	processes     map[int64]*exec.Cmd // vaultID: process
	mountPoints   map[int64]string    // vaultID: mountPoint
	lock          sync.Mutex          // lock on `processes` and `mountPoints`
}

// Template is a simple template renderer for labstack/echo
type Template struct {
	template *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.template.ExecuteTemplate(w, name, data)
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
func NewApiServer(repo *models.VaultRepo) *ApiServer {
	// Create server
	server := ApiServer{
		repo:          repo,
		echo:          echo.New(),
		cmd:           "",
		fuseAvailable: extension.IsFuseAvailable(),
		processes:     map[int64]*exec.Cmd{},
		mountPoints:   map[int64]string{},
	}

	// Detect external runtime dependencies
	var err error
	if server.cmd, err = extension.LocateGocryptfsBinary(); err != nil {
		logger.Error().Err(err).
			Msg("Failed to locate gocryptfs binary, nothing will work")
	} else {
		logger.Debug().Str("gocryptfs", server.cmd).Msg("Gocryptfs binary located")
	}
	logger.Debug().Bool("fuseAvailable", server.fuseAvailable).Msg("FUSE detection finished")

	// Setup HTTP server
	server.echo.Renderer = &Template{
		template: template.Must(template.ParseGlob("web/templates/*.html")),
	}
	server.echo.HideBanner = true
	server.echo.HidePort = true
	server.echo.Static("/static", "web") // FIXME
	server.echo.GET("/", func(c echo.Context) error {
		// Render app page
		return c.Render(200, "app.html", echo.Map{})
	})
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
	apis := server.echo.Group("/api", server.CheckRuntimeDeps)
	{
		apis.GET("/vaults", server.ListVaults)
		apis.DELETE("/vault/:id", server.RemoveVault)
		apis.POST("/vaults", server.AddOrCreateVault)
		apis.POST("/vault/:id", server.OperateOnVault)
		apis.POST("/subpaths", server.ListSubPaths)
	}
	return &server
}

// VaultInfo represents a single vault along with its current state
type VaultInfo struct {
	models.Vault
	State string `json:"state"` // Legal values: locked/unlocked
}

// ListVaults returns a list of all known vaults
func (s *ApiServer) ListVaults(c echo.Context) error {
	if vaults, err := s.repo.List(nil); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"code": ERR_CODE_LIST_FAILED,
			"msg":  ERR_MSG_LIST_FAILED,
		})
	} else {
		vaultList := make([]VaultInfo, len(vaults))
		for i, v := range vaults {
			vaultList[i] = VaultInfo{Vault: v, State: "locked"}
			// Detect vault state
			if _, ok := s.mountPoints[v.ID]; ok {
				vaultList[i].State = "unlocked"
			}
		}
		return c.JSON(http.StatusOK, echo.Map{
			"code":  ERR_CODE_OK,
			"msg":   ERR_MSG_OK,
			"items": vaultList,
		})
	}
}

// OperateOnVault performs operations on a single vault, including lock/unlock.
// - `op` identifies the operation to perform
func (s *ApiServer) OperateOnVault(c echo.Context) error {
	// Pre-check on ID
	vaultId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"code": ERR_CODE_MALFORMED_INPUT,
			"msg":  ERR_MSG_MALFORMED_INPUT,
		})
	}

	var form struct {
		Op       string `json:"op"`       // lock/unlock
		Password string `json:"password"` // for `lock` op only
	}
	if err := c.Bind(&form); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"code": ERR_CODE_MALFORMED_INPUT,
			"msg":  ERR_MSG_MALFORMED_INPUT,
		})
	}

	if form.Op == "unlock" {
		// Check current state
		if _, ok := s.mountPoints[vaultId]; ok {
			return c.JSON(http.StatusOK, echo.Map{
				"code":  ERR_CODE_ALREADY_UNLOCKED,
				"msg":   ERR_MSG_ALREADY_UNLOCKED,
				"state": "unlocked",
			})
		}

		// Lock internal maps
		s.lock.Lock()
		defer s.lock.Unlock()
		// Locate vault in repository
		vault, err := s.repo.Get(vaultId, nil)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusOK, echo.Map{
					"code": ERR_CODE_VAULT_NOT_EXISTS,
					"msg":  ERR_MSG_VAULT_NOT_EXISTS,
				})
			}
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
		}
		// Locate a mountpoint for this vault
		if vault.MountPoint == nil || strings.TrimSpace(*vault.MountPoint) == "" {
			var mountPointBase string
			if runtime.GOOS == "darwin" {
				mountPointBase = "/Volumes"
			} else {
				mountPointBase = os.TempDir()
			}
			randomMountPoint := filepath.Join(mountPointBase, strconv.FormatInt(int64(rand.Int31()), 16))
			vault.MountPoint = &randomMountPoint
		}
		// Start a gocryptfs process to unlock this vault
		args := []string{"-fg", "--", vault.Path, *vault.MountPoint}
		s.processes[vaultId] = exec.Command(s.cmd, args...)
		s.mountPoints[vaultId] = *vault.MountPoint

		// Password is piped through STDIN
		stdIn, err := s.processes[vaultId].StdinPipe()
		if err != nil {
			logger.Error().Err(err).
				Str("vaultPath", vault.Path).
				Str("mountPoint", *vault.MountPoint).
				Msg("Failed to create STDIN pipe")
			defer delete(s.processes, vaultId)
			defer delete(s.mountPoints, vaultId)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
		}

		go func() {
			defer stdIn.Close()
			if _, err := io.WriteString(stdIn, form.Password); err != nil {
				logger.Error().Err(err).
					Str("vaultPath", vault.Path).
					Str("mountPoint", *vault.MountPoint).
					Msg("Failed to pipe vault password to gocryptfs")
			}
		}()
		if err := s.processes[vaultId].Start(); err != nil {
			logger.Error().Err(err).
				Int64("vaultId", vaultId).
				Str("vaultPath", vault.Path).
				Str("mountPoint", *vault.MountPoint).
				Str("gocryptfs", s.cmd).
				Msg("Failed to start gocryptfs process")

			// Cleanup immediately
			defer delete(s.processes, vaultId)
			defer delete(s.mountPoints, vaultId)

			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
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
						Str("mountPoint", *vault.MountPoint).
						Msg("Vault locked")
					break
				default:
					logger.Error().Err(err).
						Int("RC", rc).
						Int64("vaultId", vaultId).
						Str("vaultPath", vault.Path).
						Str("mountPoint", *vault.MountPoint).
						Msg("Gocryptfs exited unexpectedly")
				}
			} else {
				logger.Debug().
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", *vault.MountPoint).
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
					Str("mountPoint", *vault.MountPoint).
					Msg("Mountpoint not empty")
				return c.JSON(http.StatusOK, echo.Map{
					"code":  ERR_CODE_MOUNTPOINT_NOT_EMPTY,
					"msg":   ERR_MSG_MOUNTPOINT_NOT_EMPTY,
					"state": "locked",
				})
			case 12: // Incorrect password
				logger.Error().Err(err).
					Int("RC", rc).
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", *vault.MountPoint).
					Msg("Vault password incorrect")
				return c.JSON(http.StatusOK, echo.Map{
					"code":  ERR_CODE_WRONG_PASSWORD,
					"msg":   ERR_MSG_WRONG_PASSWORD,
					"state": "locked",
				})
			case 23: // gocryptfs.conf IO error
				logger.Error().Err(err).
					Int("RC", rc).
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", *vault.MountPoint).
					Msg("Gocryptfs cannot open gocryptfs.conf")
				return c.JSON(http.StatusOK, echo.Map{
					"code":  ERR_CODE_CANT_OPEN_VAULT_CONF,
					"msg":   ERR_MSG_CANT_OPEN_VAULT_CONF,
					"state": "locked",
				})
			default:
				logger.Error().Err(err).
					Int("RC", rc).
					Int64("vaultId", vaultId).
					Str("vaultPath", vault.Path).
					Str("mountPoint", *vault.MountPoint).
					Msg("Gocryptfs exited unexpectedly")
				return c.JSON(http.StatusOK, echo.Map{
					"code":  ERR_CODE_UNKNOWN,
					"msg":   fmt.Sprintf(ERR_MSG_UNKNOWN, rc),
					"state": "locked",
				})
			}
		}

		// Respond
		return c.JSON(http.StatusOK, echo.Map{
			"code":  ERR_CODE_OK,
			"msg":   ERR_MSG_OK,
			"state": "unlocked",
		})
	} else if form.Op == "lock" {
		// Check current state
		if _, ok := s.mountPoints[vaultId]; !ok {
			return c.JSON(http.StatusOK, echo.Map{
				"code":  ERR_CODE_ALREADY_LOCKED,
				"msg":   ERR_MSG_ALREADY_LOCKED,
				"state": "locked",
			})
		}

		// Lock internal maps
		s.lock.Lock()
		defer s.lock.Unlock()
		// Stop corresponding gocryptfs process to lock this vault
		if err := s.processes[vaultId].Process.Signal(os.Interrupt); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
		}
		// We have a pairing gorountine to wait for gocryptfs process to exit and do the cleanup,
		// so no need to cleaning `s.processes` and `s.mountPoints` here.
		// Respond
		return c.JSON(http.StatusOK, echo.Map{
			"code":  ERR_CODE_OK,
			"msg":   ERR_MSG_OK,
			"state": "locked",
		})
	} else if form.Op == "reveal" {
		var mountPoint string
		var ok bool
		// Check current state
		if mountPoint, ok = s.mountPoints[vaultId]; !ok {
			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_ALREADY_LOCKED,
				"msg":  ERR_MSG_ALREADY_LOCKED,
			})
		}

		// Check mountpoint path existence
		if pathInfo, err := os.Stat(mountPoint); err != nil || !pathInfo.IsDir() {
			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_PATH_NOT_EXISTS,
				"msg":  ERR_MSG_PATH_NOT_EXISTS,
			})
		}

		extension.RevealInFileManager(mountPoint)
		return c.JSON(http.StatusOK, echo.Map{
			"code": ERR_CODE_OK,
			"msg":  ERR_MSG_OK,
		})
	} else {
		// Currently not supported
		return c.JSON(http.StatusOK, echo.Map{
			"code": ERR_CODE_UNSUPPORTED_OPERATION,
			"msg":  ERR_MSG_UNSUPPORTED_OPERATION,
		})
	}
}

// RemoveVault removes vault specified by ID from database repository,
// the corresponding directory remains on the disk.
func (s *ApiServer) RemoveVault(c echo.Context) error {
	// Pre-check on ID
	vaultId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"code": ERR_CODE_MALFORMED_INPUT,
			"msg":  ERR_MSG_MALFORMED_INPUT,
		})
	}

	// Lock internal maps
	s.lock.Lock()
	defer s.lock.Unlock()
	// Lock vault if necessary
	// Stop corresponding gocryptfs process to lock this vault
	if _, ok := s.mountPoints[vaultId]; ok {
		if err := s.processes[vaultId].Process.Signal(os.Interrupt); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
		}
	}

	_ = s.repo.WithTransaction(func(tx models.Transactional) error {
		vault, err := s.repo.Get(vaultId, tx)
		if err != nil {
			// Vault ID not found
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusOK, echo.Map{
					"code": ERR_CODE_VAULT_NOT_EXISTS,
					"msg":  ERR_MSG_VAULT_NOT_EXISTS,
				})
			}
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
		}
		if err = s.repo.Delete(&vault, tx); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
		} else {
			// Deleted
			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_OK,
				"msg":  ERR_MSG_OK,
			})
		}
	})
	return nil
}

// AddOrCreateVault adds an existing gocryptfs vault to the repository,
// Or it creates a new vault at specified location (not currently supported).
func (s *ApiServer) AddOrCreateVault(c echo.Context) error {
	var form struct {
		Op   string `json:"op"` // add/create
		Path string `json:"path"`
	}
	if err := c.Bind(&form); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"code": ERR_CODE_MALFORMED_INPUT,
			"msg":  ERR_MSG_MALFORMED_INPUT,
		})
	}

	// Check path existence
	if pathInfo, err := os.Stat(form.Path); err != nil || !pathInfo.IsDir() {
		return c.JSON(http.StatusOK, echo.Map{
			"code": ERR_CODE_PATH_NOT_EXISTS,
			"msg":  ERR_MSG_PATH_NOT_EXISTS,
		})
	}

	if form.Op == "add" {
		_ = s.repo.WithTransaction(func(tx models.Transactional) error {
			vault, err := s.repo.Create(echo.Map{"path": form.Path}, tx)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"code": ERR_CODE_UNKNOWN,
					"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
				})
			}
			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_OK,
				"msg":  ERR_MSG_OK,
				"item": VaultInfo{
					Vault: vault,
					State: "locked",
				},
			})
		})
		return nil
	} else {
		// Currently not supported
		return c.JSON(http.StatusOK, echo.Map{
			"code": ERR_CODE_UNSUPPORTED_OPERATION,
			"msg":  ERR_MSG_UNSUPPORTED_OPERATION,
		})
	}
}

// ListSubPaths lists items in given path.
// - `pwd` identifies the path to use, the special value `$HOME` translates to home directory of current user.
func (s *ApiServer) ListSubPaths(c echo.Context) error {
	var form struct {
		Pwd string `json:"pwd"`
	}
	if err := c.Bind(&form); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"code": ERR_CODE_MALFORMED_INPUT,
			"msg":  ERR_MSG_MALFORMED_INPUT,
		})
	}

	// Locate user home directory
	if form.Pwd == `$HOME` {
		currentUser, err := user.Current()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"code": ERR_CODE_UNKNOWN,
				"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
			})
		}
		form.Pwd = currentUser.HomeDir
	}
	// Normalize path
	form.Pwd = filepath.Clean(form.Pwd)

	// List items
	items, err := ioutil.ReadDir(form.Pwd)
	if err != nil {
		if err == os.ErrNotExist {
			return c.JSON(http.StatusOK, echo.Map{
				"code": ERR_CODE_PATH_NOT_EXISTS,
				"msg":  ERR_MSG_PATH_NOT_EXISTS,
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"code": ERR_CODE_UNKNOWN,
			"msg":  fmt.Sprintf(ERR_MSG_UNKNOWN, err),
		})
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
	return c.JSON(http.StatusOK, echo.Map{
		"code":  ERR_CODE_OK,
		"msg":   ERR_MSG_OK,
		"sep":   string(filepath.Separator),
		"pwd":   form.Pwd,
		"items": subPathItems,
	})
}

//func (s *ApiServer) CreateVault(c echo.Context) error {}
