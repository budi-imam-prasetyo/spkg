package detector

import (
	"path/filepath"
	"strings"
)

type GoDetector struct{}

func init() { Register(&GoDetector{}) }

func (d *GoDetector) Name() Manager { return ManagerGo }

func (d *GoDetector) IsAvailable() bool {
	return HasCommand("go")
}

func (d *GoDetector) Detect(execPath string) *Result {
	gopath, _ := RunCmd("go", "env", "GOPATH")
	if gopath == "" {
		gopath = filepath.Join(HomeDir(), "go")
	}

	binDir := filepath.Join(gopath, "bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)

	// If detecting go itself, show how to install/remove Go
	if base == "go" || base == "gofmt" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerGo,
			Package:        "go",
			InstallCmd:     "sudo pacman -S go  # or: sudo snap install go --classic",
			UpgradeCmd:     "sudo pacman -Syu go",
			RemoveCmd:      "sudo pacman -Rns go  # or: sudo snap remove go",
		}
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerGo,
		Package:        base,
		InstallCmd:     "go install " + base + "@latest",
		UpgradeCmd:     "go install " + base + "@latest",
		RemoveCmd:      "rm " + execPath,
	}
}
