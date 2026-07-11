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

	home := HomeDir()
	appImageDirs := []string{
		filepath.Join(home, "Applications"),
		"/opt",
		"/usr/local/bin",
	}
	base := filepath.Base(execPath)
	for _, dir := range appImageDirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		found := false
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if found || err != nil || info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".desktop") {
				content, _ := os.ReadFile(path)
				if strings.Contains(string(content), base) {
					found = true
				}
			}
			return nil
		})
		if found {
			return &Result{
				Found:          true,
				ExecutablePath: execPath,
				Manager:        ManagerAppImage,
				Package:        base,
				InstallCmd:     "# Download .AppImage file",
				UpgradeCmd:     "# Download latest .AppImage file",
				RemoveCmd:      "rm " + execPath,
			}
		}
	}

	return nil
}
