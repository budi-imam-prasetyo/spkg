package detector

import (
	"path/filepath"
	"strings"
)

type FlatpakDetector struct{}

func init() { Register(&FlatpakDetector{}) }

func (d *FlatpakDetector) Name() Manager { return ManagerFlatpak }

func (d *FlatpakDetector) IsAvailable() bool {
	_, err := flatpakBin()
	return err == nil
}

func flatpakBin() (string, error) {
	if HasCommand("flatpak") {
		return "flatpak", nil
	}
	return "", errNotFound
}

var errNotFound = &flatpakNotFoundError{}

type flatpakNotFoundError struct{}

func (e *flatpakNotFoundError) Error() string { return "flatpak not found" }

func (d *FlatpakDetector) Detect(execPath string) *Result {
	list, err := RunCmd("flatpak", "list", "--app", "--columns=application,origin,installation")
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(list, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		appID := parts[0]
		origin := ""
		if len(parts) > 1 {
			origin = parts[1]
		}

		info, _ := RunCmd("flatpak", "info", appID)
		if strings.Contains(info, execPath) || strings.Contains(info, filepath.Base(execPath)) {
			version := extractFlatpakVersion(info)
			return &Result{
				Found:          true,
				ExecutablePath: execPath,
				Manager:        ManagerFlatpak,
				Package:        appID,
				Version:        version,
				Remote:         origin,
				InstallCmd:     "flatpak install " + origin + " " + appID,
				UpgradeCmd:     "flatpak update " + appID,
				RemoveCmd:      "flatpak uninstall " + appID,
			}
		}
	}

	return nil
}

func extractFlatpakVersion(info string) string {
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "Version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		}
	}
	return ""
}
