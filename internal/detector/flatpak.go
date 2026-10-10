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
	list, err := flatpakAppList()
	if err != nil {
		return nil
	}

	base := filepath.Base(execPath)
	for _, line := range strings.Split(list, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		appID := strings.TrimSpace(parts[0])
		if appID == "" {
			continue
		}
		origin := ""
		if len(parts) > 1 {
			origin = strings.TrimSpace(parts[1])
		}

		info := flatpakInfo(appID)
		if info == "" {
			continue
		}
		if !containsWord(info, execPath) && !containsWord(info, base) {
			continue
		}
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
			ConfirmReasons: []Reason{{
				Signal: "manager_db",
				Detail: "flatpak info confirms app " + appID,
				Weight: int(ConfidenceHigh),
			}},
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

func (d *FlatpakDetector) Score(execPath string) (Confidence, []Reason) {
	if strings.HasPrefix(execPath, "/var/lib/flatpak/") || strings.HasPrefix(execPath, filepath.Join(HomeDir(), ".local", "share", "flatpak")) {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under a flatpak installation root", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
