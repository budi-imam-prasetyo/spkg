package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type AppImageDetector struct{}

func init() { Register(&AppImageDetector{}) }

func (d *AppImageDetector) Name() Manager { return ManagerAppImage }

func (d *AppImageDetector) IsAvailable() bool {
	return true
}

func (d *AppImageDetector) Detect(execPath string) *Result {
	if strings.HasSuffix(execPath, ".AppImage") {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerAppImage,
			Package:        filepath.Base(execPath),
			InstallCmd:     "# Download .AppImage file",
			UpgradeCmd:     "# Download latest .AppImage file",
			RemoveCmd:      "rm " + execPath,
		}
	}

	home, _ := os.UserHomeDir()
	appImageDirs := []string{
		filepath.Join(home, "Applications"),
		"/opt",
		"/usr/local/bin",
		"/var/lib/flatpak",
		"/var/lib/snapd",
	}
	for _, dir := range appImageDirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(dir, "**", "*.desktop"))
		for _, m := range matches {
			content, _ := os.ReadFile(m)
			if strings.Contains(string(content), filepath.Base(execPath)) {
				return &Result{
					Found:          true,
					ExecutablePath: execPath,
					Manager:        ManagerAppImage,
					Package:        filepath.Base(execPath),
					InstallCmd:     "# Download .AppImage file",
					UpgradeCmd:     "# Download latest .AppImage file",
					RemoveCmd:      "rm " + execPath,
				}
			}
		}
	}

	return nil
}
