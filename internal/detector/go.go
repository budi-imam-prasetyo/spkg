package detector

import (
	"os"
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
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}

	binDir := filepath.Join(gopath, "bin")
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)
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
