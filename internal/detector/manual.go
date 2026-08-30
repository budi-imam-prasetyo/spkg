package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type ManualDetector struct{}

func (d *ManualDetector) Name() Manager { return ManagerManual }

// NewManualResult builds the "manual" fallback result used when no detector
// claims an executable. (ManualDetector is not registered; it exists only to
// produce this fallback.)
func NewManualResult(rp ResolvedPath, name string) *Result {
	execPath := rp.Real
	return &Result{
		Found:          true,
		Executable:     name,
		ExecutablePath: execPath,
		Manager:        ManagerManual,
		Package:        name,
		InstallCmd:     "# Manual installation",
		UpgradeCmd:     "# Manual upgrade required",
		RemoveCmd:      buildRemoveCmd(execPath),
	}
}

func buildRemoveCmd(execPath string) string {
	home := HomeDir()

	// Check if it's a bun global package symlink
	// ~/.bun/bin/foo -> ../install/global/node_modules/{pkg}/...
	if strings.Contains(execPath, filepath.Join(home, ".bun", "install", "global", "node_modules")) {
		pkgName := extractBunPackageName(execPath)
		if pkgName != "" {
			return "bun remove -g " + pkgName
		}
	}

	// Check if it's an npm global package
	// ~/.local/share/fnm/node-versions/v22.18.0/installation/lib/node_modules/{pkg}/...
	if strings.Contains(execPath, "node_modules") && strings.Contains(execPath, "bin") {
		pkgName := extractNpmPackageName(execPath)
		if pkgName != "" {
			return "npm uninstall -g " + pkgName
		}
	}

	// Check if it's in a Python venv
	if strings.Contains(execPath, "venvs") || strings.Contains(execPath, "virtualenvs") {
		return "# Remove the virtual environment"
	}

	// Check if it's a cargo-installed binary
	if strings.HasPrefix(execPath, filepath.Join(home, ".cargo", "bin")) {
		pkgName := filepath.Base(execPath)
		return "cargo uninstall " + pkgName
	}

	// Check if it's a go-installed binary
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(home, "go")
	}
	if strings.HasPrefix(execPath, filepath.Join(gopath, "bin")) {
		return "# Run 'go install' to reinstall or manually remove"
	}

	// Check if it's a flatpak app
	if strings.Contains(execPath, "flatpak") || strings.Contains(execPath, ".local/share/flatpak") {
		return "# flatpak uninstall <app-id>"
	}

	// Check if it's a snap
	if strings.HasPrefix(execPath, "/snap/") {
		return "# snap remove <package>"
	}

	// Check if it's an AppImage
	if strings.HasSuffix(execPath, ".AppImage") {
		return "rm " + execPath
	}

	// For files in /usr/local/bin or similar, just remove the file
	if strings.HasPrefix(execPath, "/usr/local/bin") || strings.HasPrefix(execPath, "/usr/bin") {
		return "sudo rm " + execPath
	}

	// Default: remove the file (but warn about dependencies)
	return "rm " + execPath + " # Note: this only removes the executable, not dependencies"
}

func extractBunPackageName(execPath string) string {
	home := HomeDir()
	marker := filepath.Join(home, ".bun", "install", "global", "node_modules") + string(os.PathSeparator)
	if !strings.HasPrefix(execPath, marker) {
		return ""
	}
	rest := strings.TrimPrefix(execPath, marker)
	// Handle scoped packages like @scope/pkg
	if strings.HasPrefix(rest, "@") {
		parts := strings.SplitN(rest, string(os.PathSeparator), 3)
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
	}
	parts := strings.SplitN(rest, string(os.PathSeparator), 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func extractNpmPackageName(execPath string) string {
	// {prefix}/lib/node_modules/{pkg}/...
	idx := strings.Index(execPath, "node_modules"+string(os.PathSeparator))
	if idx == -1 {
		return ""
	}
	rest := execPath[idx+len("node_modules")+1:]
	parts := strings.SplitN(rest, string(os.PathSeparator), 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
