// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
	"os"
	"os/exec"
	"path/filepath"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

type target struct {
	goos   string
	goarch string
}

func buildForTarget(t target) (output string, err error) {
	os.RemoveAll(`rsrc.syso`)

	env := map[string]string{}
	if t.goos != "" && t.goarch != "" {
		env["GOOS"] = t.goos
		env["GOARCH"] = t.goarch
	}

	executable := "cloak"
	buildCmd := []string{
		`go`, `build`,
		`-ldflags`, `-X 'main.ReleaseMode=true' -X 'Cloak/extension.ReleaseMode=true'`,
	}
	if t.goos == "windows" {
		executable += ".exe"
		buildCmd = append(buildCmd, `-ldflags`, `"-H=windowsgui"`)
		if err = sh.Run(`rsrc`, `-manifest`, `manifest.xml`, `-o`, `rsrc.syso`); err != nil {
			return
		}
	}
	buildCmd = append(buildCmd, `-o`, executable)

	if err = sh.RunWith(env, buildCmd[0], buildCmd[1:]...); err != nil {
		return
	}

	var executableDir string
	if t.goos == "darwin" {
		executableDir = filepath.Join(`Cloak.app`, `Contents`, `MacOS`)
		if err = os.MkdirAll(executableDir, 0755); err != nil {
			return
		}
		if err = os.Rename(executable, filepath.Join(executableDir, executable)); err != nil {
			return
		}
		if err = sh.Copy(filepath.Join(executableDir, `..`, `Info.plist`), `Info.plist`); err != nil {
			return
		}
		output = `Cloak.app`
		return
	} else if t.goos == "windows" {
		output = executable
		return
	} else if t.goos == "linux" {
		// TODO
		return
	} else {
		err = fmt.Errorf("unsupported OS: %s", t.goos)
		return
	}
}

// Build build source code files into OS-specific executable
func Build() error {
	mg.Deps(InstallDeps, Clean, PackAssets)
	for _, t := range []target{
		{"darwin", "amd64"},
		//{"windows", "amd64"},
		//{"linux", "amd64"},
	} {
		fmt.Printf("Building for OS=%s ARCH=%s... ", t.goos, t.goarch)
		if output, err := buildForTarget(t); err != nil {
			return err
		} else {
			fmt.Printf("Bundle created: %s\n", output)
		}
	}
	return nil
}

// PackAssets packs static files using `statik` tool
func PackAssets() error {
	return sh.Run(`statik`, `-src`, `web`, `-dest`, `.`, `-f`)
}

// InstallDeps installs extra tools required for building
func InstallDeps() error {
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
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("cloak")
	os.RemoveAll("cloak.exe")
	os.RemoveAll(`Cloak.app`)
	//os.RemoveAll(`rsrc.syso`)
}
