package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type BrewDetector struct{}

func init() { Register(&BrewDetector{}) }

func (d *BrewDetector) Name() Manager { return ManagerBrew }

func (d *BrewDetector) IsAvailable() bool {
	return HasCommand("brew")
}

func (d *BrewDetector) Detect(execPath string) *Result {
	prefix, _ := RunCmd("brew", "--prefix")
	if prefix == "" {
		return nil
	}
	binDir := filepath.Join(prefix, "bin")
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)
	info, err := RunCmd("brew", "info", "--json=v2", base)
	if err != nil {
		return nil
	}

	version := ""
	_ = info
	if idx := strings.Index(info, "\"stable\""); idx != -0 {
		verIdx := strings.Index(info[idx:], "\"version\"")
		if verIdx != -1 {
			sub := info[idx+verIdx:]
			lines := strings.SplitN(sub, "\n", 3)
			if len(lines) > 0 {
				parts := strings.SplitN(lines[0], ":", 2)
				if len(parts) > 1 {
					version = strings.Trim(parts[1], " \",")
				}
			}
		}
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerBrew,
		Package:        base,
		Version:        version,
		InstallCmd:     "brew install " + base,
		UpgradeCmd:     "brew upgrade " + base,
		RemoveCmd:      "brew uninstall " + base,
	}
}

func detectBrewExec(execPath string) bool {
	prefix, _ := RunCmd("brew", "--prefix")
	if prefix == "" {
		return false
	}
	return strings.HasPrefix(execPath, filepath.Join(prefix, "bin"))
}

func isBrewAvailable() bool {
	home, _ := os.UserHomeDir()
	linuxbrew := filepath.Join(home, ".linuxbrew", "bin", "brew")
	if _, err := os.Stat(linuxbrew); err == nil {
		return true
	}
	return HasCommand("brew")
}
