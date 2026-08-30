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
	// Bound the search to user-owned application directories. We deliberately do
	// NOT walk system dirs like /opt (potentially huge, slow, and reads unrelated
	// files). A reasonable file-count cap guards against runaway walks.
	appImageDirs := []string{
		filepath.Join(home, "Applications"),
		filepath.Join(home, ".local", "share", "applications"),
	}
	base := filepath.Base(execPath)
	for _, dir := range appImageDirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		found := false
		scanned := 0
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if found || err != nil || info.IsDir() {
				return nil
			}
			// Cap the number of files we inspect to keep the walk bounded.
			scanned++
			if scanned > 2000 {
				return filepath.SkipDir
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

func (d *AppImageDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	dirs := []string{filepath.Join(home, "Applications"), filepath.Join(home, ".local", "share", "applications")}
	for _, dir := range dirs {
		if strings.HasPrefix(execPath, dir) {
			return ConfidenceLow, []Reason{{Signal: "path_match", Detail: "path is in an AppImage application dir", Weight: int(ConfidenceLow)}}
		}
	}
	return ConfidenceUnknown, nil
}
