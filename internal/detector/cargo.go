package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type CargoDetector struct{}

func init() { Register(&CargoDetector{}) }

func (d *CargoDetector) Name() Manager { return ManagerCargo }

func (d *CargoDetector) IsAvailable() bool {
	return HasCommand("cargo")
}

func (d *CargoDetector) Detect(execPath string) *Result {
	home, _ := os.UserHomeDir()
	cargoBin := filepath.Join(home, ".cargo", "bin")
	if !strings.HasPrefix(execPath, cargoBin) {
		return nil
	}

	list, err := RunCmd("cargo", "install", "--list")
	if err != nil {
		return nil
	}

	base := filepath.Base(execPath)
	lines := strings.Split(list, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ":") {
			pkgName := strings.TrimSuffix(line, ":")
			for j := i + 1; j < len(lines) && j < i+5; j++ {
				if strings.Contains(lines[j], filepath.Join(cargoBin, base)) {
					parts := strings.Fields(pkgName)
					name := parts[0]
					version := ""
					if len(parts) > 1 {
						version = strings.Trim(parts[1], "v")
					}
					return &Result{
						Found:          true,
						ExecutablePath: execPath,
						Manager:        ManagerCargo,
						Package:        name,
						Version:        version,
						InstallCmd:     "cargo install " + name,
						UpgradeCmd:     "cargo install " + name,
						RemoveCmd:      "cargo uninstall " + name,
					}
				}
			}
		}
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerCargo,
		Package:        base,
		InstallCmd:     "cargo install " + base,
		UpgradeCmd:     "cargo install " + base,
		RemoveCmd:      "cargo uninstall " + base,
	}
}
