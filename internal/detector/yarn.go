package detector

import (
	"path/filepath"
	"strings"
)

type YarnDetector struct{}

func init() { Register(&YarnDetector{}) }

func (d *YarnDetector) Name() Manager { return ManagerYarn }

func (d *YarnDetector) IsAvailable() bool {
	return HasCommand("yarn")
}

func (d *YarnDetector) Detect(execPath string) *Result {
	globalDir := yarnGlobalDir()
	if globalDir == "" || strings.Contains(strings.ToLower(globalDir), "error") {
		return nil
	}
	binDir := filepath.Join(globalDir, "node_modules", ".bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)

	// If detecting yarn itself, show how to install/remove Yarn
	if base == "yarn" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerYarn,
			Package:        "yarn",
			Global:         BoolPtr(true),
			InstallCmd:     "npm install -g yarn  # or: corepack enable",
			UpgradeCmd:     "npm update -g yarn  # or: corepack prepare yarn@stable --activate",
			RemoveCmd:      "npm uninstall -g yarn",
		}
	}

	list := yarnGlobalList()
	pkgName := extractYarnPkgName(list, base)
	if pkgName == "" {
		pkgName = base
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerYarn,
		Package:        pkgName,
		Global:         BoolPtr(true),
		InstallCmd:     "yarn global add " + pkgName,
		UpgradeCmd:     "yarn global upgrade " + pkgName,
		RemoveCmd:      "yarn global remove " + pkgName,
		ConfirmReasons: []Reason{{
			Signal: "manager_db",
			Detail: "yarn global dir owns this path (" + globalDir + ")",
			Weight: int(ConfidenceHigh),
		}},
	}
}

func extractYarnPkgName(listOutput, binName string) string {
	for _, line := range strings.Split(listOutput, "\n") {
		if !containsWord(line, binName) {
			continue
		}
		parts := strings.Fields(line)
		for _, p := range parts {
			if strings.Contains(p, "@") {
				name := strings.Trim(p, ",")
				if idx := strings.LastIndex(name, "@"); idx > 0 {
					return name[:idx]
				}
			}
		}
	}
	return ""
}

func (d *YarnDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	global := filepath.Join(home, ".yarn")
	if strings.HasPrefix(execPath, global) || strings.HasPrefix(execPath, filepath.Join(home, ".local", "bin")) {
		return ConfidenceLow, []Reason{{Signal: "ambiguous_path", Detail: "path is under a yarn global/shared bin dir", Weight: int(ConfidenceLow)}}
	}
	return ConfidenceUnknown, nil
}
