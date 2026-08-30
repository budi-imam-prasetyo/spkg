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
	binDir := filepath.Join(prefix, "node_modules", ".bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)

	// If detecting pnpm itself, show how to install/remove pnpm
	if base == "pnpm" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerPnpm,
			Package:        "pnpm",
			Global:         BoolPtr(true),
			InstallCmd:     "npm install -g pnpm  # or: curl -fsSL https://get.pnpm.io/install.sh | sh -",
			UpgradeCmd:     "npm update -g pnpm  # or: pnpm add -g pnpm",
			RemoveCmd:      "npm uninstall -g pnpm  # or: rm -rf ~/.local/share/pnpm",
		}
	}

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
				name := parts[0]
				// Strip trailing version: @scope/pkg@2.0.0 → @scope/pkg
				if idx := strings.LastIndex(name, "@"); idx > 0 {
					name = name[:idx]
				}
				return name
			}
		}
	}
	return ""
}

func (d *PnpmDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	store := filepath.Join(home, ".local", "share", "pnpm")
	global := filepath.Join(home, ".pnpm-store")
	if strings.HasPrefix(execPath, store) || strings.HasPrefix(execPath, global) || strings.HasPrefix(execPath, filepath.Join(home, ".local", "bin")) {
		return ConfidenceLow, []Reason{{Signal: "ambiguous_path", Detail: "path is under a pnpm store/shared bin dir", Weight: int(ConfidenceLow)}}
	}
	return ConfidenceUnknown, nil
}
