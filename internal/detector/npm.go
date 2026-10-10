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
	prefix := npmPrefix()
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
			InstallCmd:     "npm install -g npm",
			UpgradeCmd:     "npm install -g npm@latest",
			RemoveCmd:      "# npm is normally installed with Node.js",
		}
	}

	globalList := npmGlobalList()
	pkgName := extractNpmPkgName(globalList, base)
	// The resolved path may point at the real script inside the package
	// (e.g. .../node_modules/9remote/dist/cli.cjs), so `base` is a filename
	// like "cli.cjs", not the package name. Recover the real package name
	// from the node_modules path segment when possible.
	if fromPath := pkgFromNodeModulesPath(execPath); fromPath != "" {
		pkgName = fromPath
	}
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
		ConfirmReasons: []Reason{{
			Signal: "manager_db",
			Detail: "npm global prefix owns this path (" + prefix + ")",
			Weight: int(ConfidenceHigh),
		}},
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

func (d *NpmDetector) Score(execPath string) (Confidence, []Reason) {
	if strings.Contains(execPath, "node_modules") || strings.Contains(execPath, ".nvm") || strings.Contains(execPath, "fnm") || strings.Contains(execPath, "/.volta/") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under a node_modules or node runtime dir", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}

// pkgFromNodeModulesPath extracts the package name from a path that contains a
// node_modules segment, handling scoped packages. It returns the directory name
// immediately after the LAST "node_modules" component:
//
//	.../lib/node_modules/9remote/dist/cli.cjs  -> "9remote"
//	.../node_modules/@scope/pkg/bin/x.js       -> "@scope/pkg"
//
// Returns "" when the path has no node_modules segment.
func pkgFromNodeModulesPath(execPath string) string {
	parts := strings.Split(filepath.ToSlash(execPath), "/")
	idx := -1
	for i, p := range parts {
		if p == "node_modules" {
			idx = i
		}
	}
	if idx == -1 || idx+1 >= len(parts) {
		return ""
	}
	name := parts[idx+1]
	if name == "" {
		return ""
	}
	// Scoped package: @scope/pkg
	if strings.HasPrefix(name, "@") && idx+2 < len(parts) && parts[idx+2] != "" {
		return name + "/" + parts[idx+2]
	}
	return name
}
