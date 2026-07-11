package detector

import (
	"path/filepath"
	"strings"
)

type NpmDetector struct{}

func init() { Register(&NpmDetector{}) }

func (d *NpmDetector) Name() Manager { return ManagerNpm }

func (d *NpmDetector) IsAvailable() bool {
	return HasCommand("npm")
}

func (d *NpmDetector) Detect(execPath string) *Result {
	prefix, _ := RunCmd("npm", "prefix", "-g")
	if prefix == "" {
		return nil
	}
	binDir := filepath.Join(prefix, "bin")

	if !strings.HasPrefix(execPath, binDir) && execPath != filepath.Join(prefix, "bin", filepath.Base(execPath)) {
		altBin := filepath.Join(prefix, "node_modules", ".bin")
		if !strings.HasPrefix(execPath, altBin) {
			return nil
		}
	}

	globalList, _ := RunCmd("npm", "list", "-g", "--depth=0", "--json")
	base := filepath.Base(execPath)
	pkgName := extractNpmPkgName(globalList, base)
	if pkgName == "" {
		pkgName = base
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerNpm,
		Package:        pkgName,
		Global:         BoolPtr(true),
		InstallCmd:     "npm install -g " + pkgName,
		UpgradeCmd:     "npm update -g " + pkgName,
		RemoveCmd:      "npm uninstall -g " + pkgName,
	}
}

func extractNpmPkgName(jsonOutput, binName string) string {
	for _, line := range strings.Split(jsonOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "\""+binName+"\"") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 0 {
				name := strings.Trim(parts[0], " \"")
				return name
			}
		}
	}
	return ""
}
