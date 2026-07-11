package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type UvDetector struct{}

func init() { Register(&UvDetector{}) }

func (d *UvDetector) Name() Manager { return ManagerUv }

func (d *UvDetector) IsAvailable() bool {
	return HasCommand("uv")
}

func (d *UvDetector) Detect(execPath string) *Result {
	home, _ := os.UserHomeDir()
	uvBin := filepath.Join(home, ".local", "bin")
	if !strings.HasPrefix(execPath, uvBin) {
		return nil
	}

	show, err := RunCmd("uv", "tool", "list")
	if err != nil {
		return nil
	}

	base := filepath.Base(execPath)
	for _, line := range strings.Split(show, "\n") {
		if strings.Contains(line, base) {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				pkgName := parts[0]
				return &Result{
					Found:          true,
					ExecutablePath: execPath,
					Manager:        ManagerUv,
					Package:        pkgName,
					InstallCmd:     "uv tool install " + pkgName,
					UpgradeCmd:     "uv tool upgrade " + pkgName,
					RemoveCmd:      "uv tool uninstall " + pkgName,
				}
			}
		}
	}

	return nil
}
