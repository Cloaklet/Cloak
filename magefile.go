//go:build mage
// +build mage

package main

import (
	"Cloak/server"
	"bytes"
	"context"
	json2 "encoding/json"
	"fmt"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

const (
	osKey = iota
	archKey
)

var logger zerolog.Logger

func init() {
	logger = zlog.Logger
}

func buildForTarget(c context.Context) (output string, err error) {
	env := map[string]string{
		"GOOS":   c.Value(osKey).(string),
		"GOARCH": c.Value(archKey).(string),
	}

	var versionString string

	// Read version string from git
	gitOutput, err := sh.Output("git", "describe", "--tags", "--exact-match")
	// We're at an exact tag
	if err == nil {
		versionString = strings.TrimSpace(gitOutput)
		logger.Debug().Str("version", versionString).Msg("Determined exact Git tag")
	} else {
		// We've got commits after a tag
		gitOutput, err = sh.Output("git", "describe", "--tags")
		// Unable to locate git tag
		if err != nil {
			logger.Error().Err(err).Msg("Failed to determine version via Git")
			return "", err
		}
		versionString = strings.TrimSpace(gitOutput)
		logger.Warn().Str("version", versionString).Msg("Not at an exact Git tag")
	}

	// Get commit ID from git
	commitString := "unknown"
	if commit, err := sh.Output(`git`, `rev-parse`, `--short`, `HEAD`); err != nil {
		return "", err
	} else {
		commitString = commit
	}
	currentTimeString := time.Now().Format(`2006-01-02 15:04:05 MST`)

	if err := sh.RunV("go", "generate", "./..."); err != nil {
		return "", err
	}

	executable := "cloak"
	buildCmd := []string{
		`go`, `build`,
		`-ldflags`, strings.Join([]string{
			`-X 'Cloak/extension.ReleaseMode=true'`,
			fmt.Sprintf(`-X 'Cloak/version.Version=%s'`, versionString),
			fmt.Sprintf(`-X 'Cloak/version.BuildTime=%s'`, currentTimeString),
			fmt.Sprintf(`-X 'Cloak/version.GitCommit=%s'`, commitString),
		}, " "),
	}
	buildCmd = append(buildCmd, `-o`, executable)

	if err = sh.RunWith(env, buildCmd[0], buildCmd[1:]...); err != nil {
		return
	}

	var executableDir string
	switch env["GOOS"] {
	case "darwin":
		executableDir = filepath.Join(`Cloak.app`, `Contents`, `MacOS`)
		if err = os.MkdirAll(executableDir, 0755); err != nil {
			return
		}
		resourceDir := filepath.Join(`Cloak.app`, `Contents`, `Resources`)
		if err = os.MkdirAll(resourceDir, 0755); err != nil {
			return
		}

		// Here's a list of files to be bundled
		files := map[string]string{
			executable:       filepath.Join(executableDir, executable),
			"Info.plist":     filepath.Join(executableDir, `..`, `Info.plist`),
			"gocryptfs":      filepath.Join(executableDir, "gocryptfs"),
			"gocryptfs-xray": filepath.Join(executableDir, "gocryptfs-xray"),
			"Cloak.icns":     filepath.Join(resourceDir, "Cloak.icns"),
		}
		for filename := range files {
			if err := sh.Copy(files[filename], filename); err != nil {
				return output, err
			}
		}
		output = `Cloak.app`

		// Zip the app bundle
		sh.RunV(
			"ditto",
			"-c", "-k", "--sequesterRsrc", "--keepParent", "Cloak.app", "Cloak.app.zip",
		)
		return
	case "linux":
		executableDir := filepath.Join("AppDir", "usr", "bin")
		if err = os.MkdirAll(executableDir, 0755); err != nil {
			return
		}

		files := map[string]string{
			"gocryptfs":      filepath.Join(executableDir, "gocryptfs"),
			"gocryptfs-xray": filepath.Join(executableDir, "gocryptfs-xray"),
		}
		for filename := range files {
			if err := sh.Copy(files[filename], filename); err != nil {
				return output, err
			}
		}

		// Locate linuxdeploy tool
		var linuxDeploy string
		if linuxDeploy, err = exec.LookPath("linuxdeploy.AppImage"); err != nil {
			if _, err = os.Stat("./linuxdeploy.AppImage"); err != nil {
				return output, fmt.Errorf("Cannot locate the required tool 'linuxdeploy.AppImage'")
			} else {
				linuxDeploy = "./linuxdeploy.AppImage"
			}
		}

		// Pack AppImage binary
		err = sh.RunWithV(
			map[string]string{
				"VERSION": versionString,
			},
			linuxDeploy,
			`--executable`, executable,
			`--appdir`, `AppDir`,
			`--desktop-file`, `Cloak.desktop`,
			`--icon-file`, `Cloak.png`,
			`--output`, `appimage`,
		)
		output = fmt.Sprintf(`Cloak-%s-%s.AppImage`, versionString, linuxArch(env["GOARCH"]))
		return output, err
	default:
		err = fmt.Errorf("unsupported OS: %s", env["GOOS"])
		return
	}
}

// Build build source code files into OS-specific executable
func Build() error {
	ctx := context.WithValue(context.TODO(), osKey, runtime.GOOS)
	ctx = context.WithValue(ctx, archKey, runtime.GOARCH)
	mg.SerialCtxDeps(ctx, Clean, InstallDeps, DownloadExternalTools, PackAssets)

	logger.Debug().Str("OS", runtime.GOOS).Str("ARCH", runtime.GOARCH).Msg("Building...")
	if output, err := buildForTarget(ctx); err != nil {
		logger.Debug().Msg(output)
		return err
	} else {
		logger.Debug().Str("file", output).Msg("Bundle created")
		return nil
	}
}

// Ref: https://github.com/magefile/mage/pull/220/files#diff-cad56499684fcd6aa40dccc829a443b9d4895a7cf898f211cf910b5af9d7e15aR38
func CopyDir(dst, src string) error {
	fn := func(srcpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if srcpath == src {
			return nil
		}

		if strings.HasPrefix(filepath.Base(srcpath), ".") {
			return filepath.SkipDir
		}

		dstpath := filepath.Clean(strings.Replace(srcpath, src, dst, 1))

		if fi.IsDir() {
			mkerr := os.Mkdir(dstpath, fi.Mode())
			if os.IsExist(mkerr) {
				return os.Chmod(dstpath, fi.Mode())
			}

			return mkerr
		}

		return copyFile(dstpath, srcpath, fi)
	}

	if err := filepath.Walk(src, fn); err != nil {
		return fmt.Errorf("can't copy %s to %s: %v", src, dst, err)
	}

	return nil
}

func copyFile(dst, src string, sfi os.FileInfo) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}

	defer from.Close()

	to, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sfi.Mode())

	if err != nil {
		return err
	}

	defer to.Close()

	if _, err = io.Copy(to, from); err != nil {
		return err
	}

	return nil
}

// PackAssets packs static files
func PackAssets(_ context.Context) error {
	logger.Debug().Msg("Building frontend...")
	npmBuild := exec.Command(`npm`, `run`, `build`)
	npmBuild.Dir = "frontend"
	output, err := npmBuild.CombinedOutput()
	logger.Debug().Msg(string(output))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to build frontend project")
		return err
	}
	logger.Debug().Msg("Successfully built frontend project")

	// Remove `server/dist/.gitkeep`
	sh.Rm(filepath.Join("server", "dist", ".gitkeep"))
	sh.Rm(filepath.Join("server", "dist", ".DS_Store"))
	// Copy `frontend/dist` to `server/dist`
	return CopyDir(filepath.Join("server", "dist"), filepath.Join("frontend", "dist"))
}

// InstallDeps installs extra tools required for building
func InstallDeps(_ context.Context) error {
	goPath, err := sh.Output(`go`, `env`, `GOPATH`)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to determine GOPATH")
		return err
	}
	if goPath == "" {
		logger.Error().Msg("GOPATH is empty")
		return fmt.Errorf("failed to get GOPATH")
	}

	return nil
}

// Clean remove build artifacts from last build
func Clean(c context.Context) error {
	logger.Debug().Msg("Cleaning...")
	goOs := c.Value(osKey).(string)
	switch goOs {
	case "darwin":
		sh.Rm("Cloak.app")
		sh.Rm("cloak")
		sh.Rm("Cloak")
	case "linux":
		sh.Rm("cloak")
		sh.Rm("Cloak")
		sh.Rm("AppDir")
	default:
		return fmt.Errorf("Unsupported OS: %s", goOs)
	}
	sh.Rm("gocryptfs")
	sh.Rm("gocryptfs-xray")
	return nil
}

// downloadExecutable downloads given URL into an executable file named by `name` in current directory.
func downloadExecutable(url string, name string) error {
	logger.Debug().Str("url", url).Str("binary", name).Msg("Downloading binary executable")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	binFile, err := os.Create(name)
	if err != nil {
		return err
	}

	defer binFile.Close()
	if _, err := io.Copy(binFile, resp.Body); err != nil {
		return err
	}

	return os.Chmod(name, 0755)
}

func linuxArch(goArch string) string {
	archString, ok := map[string]string{
		"386":   "i386",
		"amd64": "x86_64",
	}[goArch]
	if !ok {
		panic(fmt.Errorf("Unsupported architecture: %s", goArch))
	}
	return archString
}

// Download static build binary of gocryptfs
func DownloadExternalTools(c context.Context) error {
	cloakResourceVersion := "0.0.4"
	gocryptfsVersion := "v2.4.0-33-gf06f27e"
	goOs := c.Value(osKey).(string)
	goArch := c.Value(archKey).(string)

	// Here's a list of external tools to be downloaded, they are going to be bundled
	tools := map[string]string{
		"gocryptfs":      "https://github.com/Cloaklet/resources/releases/download/%s/gocryptfs-%s-%s-%s",
		"gocryptfs-xray": "https://github.com/Cloaklet/resources/releases/download/%s/gocryptfs-xray-%s-%s-%s",
	}
	switch goOs {
	case "darwin", "linux":
		for name := range tools {
			toolUrl := fmt.Sprintf(tools[name], cloakResourceVersion, gocryptfsVersion, goOs, goArch)
			if err := downloadExecutable(toolUrl, name); err != nil {
				return err
			}
		}
		if goOs == "linux" {
			return downloadExecutable(
				fmt.Sprintf("https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-%s.AppImage", linuxArch(c.Value(archKey).(string))),
				"linuxdeploy.AppImage",
			)
		}
		return nil
	default:
		return fmt.Errorf("Unsupported OS: %s", goOs)
	}
}

// PackErrorsIntoLocales [for DEVs] inject all missing error codes into UI locales.
func PackErrorsIntoLocales() error {
	errors := []*server.ApiError{
		server.ErrOk,
		server.ErrListFailed,
		server.ErrMalformedInput,
		server.ErrUnknown,
		server.ErrPathNotExist,
		server.ErrUnsupportedOperation,
		server.ErrVaultNotExist,
		server.ErrVaultAlreadyUnlocked,
		server.ErrVaultAlreadyLocked,
		server.ErrMountpointNotEmpty,
		server.ErrWrongPassword,
		server.ErrCantOpenVaultConf,
		server.ErrMissingGocryptfsBinary,
		server.ErrMissingFuse,
		server.ErrVaultMkdirFailed,
		server.ErrVaultDirNotEmpty,
		server.ErrVaultPasswordEmpty,
		server.ErrVaultInitConfFailed,
		server.ErrVaultUpdateConfFailed,
		server.ErrMissingGocryptfsXrayBinary,
		server.ErrMountpointMkdirFailed,
	}
	localesDir := filepath.Join("frontend", "src", "locales")
	localeFiles, err := ioutil.ReadDir(localesDir)
	if err != nil {
		return err
	}

	for _, file := range localeFiles {
		logger.Debug().Str("file", file.Name()).Msg("Processing locale file")
		jsonBytes, err := ioutil.ReadFile(filepath.Join(localesDir, file.Name()))
		if err != nil {
			return err
		}
		json := string(jsonBytes)

		if errorSubKey := gjson.Get(json, "errors"); !errorSubKey.Exists() {
			// Initialize .errors object
			if json, err = sjson.Set(json, "errors", map[string]string{}); err != nil {
				return err
			}
		}
		for _, error := range errors {
			errorKey := fmt.Sprintf("errors.api_%d", error.Code)
			if errorValue := gjson.Get(json, errorKey); !errorValue.Exists() {
				errorString := ""
				if file.Name() == "en.json" {
					errorString = error.Message
				}
				if json, err = sjson.Set(json, errorKey, errorString); err != nil {
					return err
				}
				logger.Debug().Int("errorCode", error.Code).Send()
			}
		}
		var jsonOut bytes.Buffer
		if err = json2.Indent(&jsonOut, []byte(json), "", "  "); err != nil {
			return err
		}
		if err = ioutil.WriteFile(filepath.Join(localesDir, file.Name()), jsonOut.Bytes(), file.Mode()); err != nil {
			return err
		}
		logger.Debug().Msg("Done.")
	}
	return nil
}
