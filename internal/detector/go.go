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
	gopath := goGopath()
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
			InstallCmd:     "# Install Go from https://go.dev/dl/",
			UpgradeCmd:     "# Install the latest Go release from https://go.dev/dl/",
			RemoveCmd:      "# Remove the Go installation directory",
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

func (d *GoDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	binDir := filepath.Join(home, "go", "bin")
	if strings.HasPrefix(execPath, binDir) || strings.HasPrefix(execPath, "/usr/local/go/bin") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is in GOPATH bin or go install dir", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
