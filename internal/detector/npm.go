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
	binDir := filepath.Join(prefix, "bin") + string(filepath.Separator)
	libModules := filepath.Join(prefix, "lib", "node_modules") + string(filepath.Separator)

	// Check if path is in npm's bin dir or lib/node_modules
	inBin := strings.HasPrefix(execPath, binDir)
	inLib := strings.HasPrefix(execPath, libModules)
	if !inBin && !inLib {
		altBin := filepath.Join(prefix, "node_modules", ".bin") + string(filepath.Separator)
		if !strings.HasPrefix(execPath, altBin) {
			return nil
		}
	}

	base := filepath.Base(execPath)

	// If detecting npm itself, show how to install/remove npm
	if base == "npm" || base == "npx" || base == "npm-cli.js" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerNpm,
			Package:        "npm",
			Global:         BoolPtr(true),
			InstallCmd:     "sudo pacman -S npm  # or: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo bash - && sudo apt install -y nodejs",
			UpgradeCmd:     "sudo pacman -Syu npm",
			RemoveCmd:      "sudo pacman -Rns npm  # npm comes with nodejs",
		}
	}

	globalList, _ := RunCmd("npm", "list", "-g", "--depth=0", "--json")
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
