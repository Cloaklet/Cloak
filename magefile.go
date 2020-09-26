// +build mage

package main

import (
	"context"
	"fmt"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
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

func buildForTarget(c context.Context) (output string, err error) {
	os.RemoveAll(`rsrc.syso`)

	env := map[string]string{
		"GOOS":   c.Value(osKey).(string),
		"GOARCH": c.Value(archKey).(string),
	}

	// Read version string from version/VERSION
	var versionString string
	if version, err := ioutil.ReadFile(filepath.Join("version", "VERSION")); err != nil {
		return "", err
	} else {
		versionString = string(version)
	}
	// Get commit ID from git
	commitString := "unknown"
	if commit, err := sh.Output(`git`, `rev-parse`, `--short`, `HEAD`); err != nil {
		return "", err
	} else {
		commitString = commit
	}
	currentTimeString := time.Now().Format(`2006-01-02 15:04:05 MST`)

	executable := "cloak"
	buildCmd := []string{
		`go`, `build`,
		`-ldflags`, strings.Join([]string{
			`-X 'main.ReleaseMode=true'`,
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

		// Here's a list of files to be bundled
		files := map[string]string{
			executable:       filepath.Join(executableDir, executable),
			"Info.plist":     filepath.Join(executableDir, `..`, `Info.plist`),
			"gocryptfs":      filepath.Join(executableDir, "gocryptfs"),
			"gocryptfs-xray": filepath.Join(executableDir, "gocryptfs-xray"),
		}
		for filename, _ := range files {
			if err := sh.Copy(files[filename], filename); err != nil {
				return output, err
			}
		}
		output = `Cloak.app`
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
		for filename, _ := range files {
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
			`--icon-file`, `Cloak.svg`,
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

	fmt.Printf("Building for OS=%s ARCH=%s... ", runtime.GOOS, runtime.GOARCH)
	if output, err := buildForTarget(ctx); err != nil {
		fmt.Print(output)
		return err
	} else {
		fmt.Printf("Bundle created: %s\n", output)
		return nil
	}
}

// PackAssets packs static files using `statik` tool
func PackAssets(_ context.Context) error {
	npmBuild := exec.Command(`npm`, `run`, `build`)
	npmBuild.Dir = "frontend"
	output, err := npmBuild.CombinedOutput()
	fmt.Printf("%s\n", output)
	if err != nil {
		return err
	}

	return sh.Run(`statik`, `-src`, `frontend/dist`, `-dest`, `.`, `-f`)
}

// InstallDeps installs extra tools required for building
func InstallDeps(_ context.Context) error {
	fmt.Println("Installing Deps...")
	for toolBinary, toolPkg := range map[string]string{
		"statik": "github.com/rakyll/statik",
	} {
		if toolPath, err := exec.LookPath(toolBinary); err != nil {
			fmt.Printf("> %s not found, install from %s\n", toolBinary, toolPkg)
			if err = sh.Run(`go`, `install`, toolPkg); err != nil {
				return err
			}
		} else {
			fmt.Printf("> Found %s: %s\n", toolBinary, toolPath)
		}
	}
	//return sh.Run(`go`, `get`, `github.com/akavel/rsrc`)
	return nil
}

// Clean remove build artifacts from last build
func Clean(c context.Context) error {
	fmt.Println("Cleaning...")
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
	//os.RemoveAll(`rsrc.syso`)
	return nil
}

// downloadExecutable downloads given URL into an executable file named by `name` in current directory.
func downloadExecutable(url string, name string) error {
	fmt.Printf("  > Downloading %s from %s\n", name, url)
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
		"386": "i386",
		"amd64": "x86_64",
	}[goArch]
	if !ok {
		panic(fmt.Errorf("Unsupported architecture: %s", goArch))
	}
	return archString
}

// Download static build binary of gocryptfs
func DownloadExternalTools(c context.Context) error {
	cloakVersion := "0.0.1"
	gocryptfsVersion := "1.8.0"
	goOs := c.Value(osKey).(string)

	// Here's a list of external tools to be downloaded, they are going to be bundled
	tools := map[string]string{
		"gocryptfs":      "https://github.com/Cloaklet/resources/releases/download/%s/gocryptfs-%s-%s",
		"gocryptfs-xray": "https://github.com/Cloaklet/resources/releases/download/%s/gocryptfs-xray-%s-%s",
	}
	switch goOs {
	case "darwin", "linux":
		for name, _ := range tools {
			toolUrl := fmt.Sprintf(tools[name], cloakVersion, gocryptfsVersion, goOs)
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
