package detector

import (
	"io"
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
	// Renamed AppImages retain the ELF AppImage marker at offset 8.
	// A desktop launcher mentioning a binary is not ownership evidence.
	if !strings.HasSuffix(execPath, ".AppImage") && !hasAppImageHeader(execPath) {
		return nil
	}
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

func hasAppImageHeader(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	var header [11]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return false
	}
	return string(header[:4]) == "\x7fELF" && header[8] == 'A' && header[9] == 'I' && (header[10] == 1 || header[10] == 2)
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
