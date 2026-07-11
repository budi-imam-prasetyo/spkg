package detector

import (
	"path/filepath"
	"strings"
)

type PnpmDetector struct{}

func init() { Register(&PnpmDetector{}) }

func (d *PnpmDetector) Name() Manager { return ManagerPnpm }

func (d *PnpmDetector) IsAvailable() bool {
	return HasCommand("pnpm")
}

func (d *PnpmDetector) Detect(execPath string) *Result {
	prefix, _ := RunCmd("pnpm", "prefix", "-g")
	if prefix == "" {
		return nil
	}
	binDir := filepath.Join(prefix, "node_modules", ".bin")
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)
	list, _ := RunCmd("pnpm", "list", "-g", "--depth=0")
	pkgName := extractPnpmPkgName(list, base)
	if pkgName == "" {
		pkgName = base
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerPnpm,
		Package:        pkgName,
		Global:         BoolPtr(true),
		InstallCmd:     "pnpm add -g " + pkgName,
		UpgradeCmd:     "pnpm update -g " + pkgName,
		RemoveCmd:      "pnpm remove -g " + pkgName,
	}
}

func extractPnpmPkgName(listOutput, binName string) string {
	for _, line := range strings.Split(listOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, binName) {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				name := strings.Trim(parts[0], "@")
				if idx := strings.LastIndex(name, "@"); idx > 0 {
					name = name[:idx]
				}
				return name
			}
		}
	}
	return ""
}
