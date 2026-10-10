package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type SnapDetector struct{}

func init() { Register(&SnapDetector{}) }

func (d *SnapDetector) Name() Manager { return ManagerSnap }

func (d *SnapDetector) IsAvailable() bool {
	return HasCommand("snap")
}

func (d *SnapDetector) Detect(execPath string) *Result {
	list, err := snapList()
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(list, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		if name == "Name" {
			continue
		}
		version := ""
		if len(parts) > 1 {
			version = parts[1]
		}

		info := snapInfo(name)
		if info == "" {
			continue
		}
		if strings.Contains(info, execPath) || strings.Contains(info, "/snap/"+name) {
			return &Result{
				Found:          true,
				ExecutablePath: execPath,
				Manager:        ManagerSnap,
				Package:        name,
				Version:        version,
				InstallCmd:     "snap install " + name,
				UpgradeCmd:     "snap refresh " + name,
				RemoveCmd:      "snap remove " + name,
				ConfirmReasons: []Reason{{
					Signal: "manager_db",
					Detail: "snap info confirms package " + name,
					Weight: int(ConfidenceHigh),
				}},
			}
		}
	}

	return nil
}

func (d *SnapDetector) Score(execPath string) (Confidence, []Reason) {
	if _, err := os.Stat(filepath.Join("/snap", "bin")); err == nil {
		if strings.HasPrefix(execPath, "/snap/") {
			return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under /snap (snap-managed)", Weight: int(ConfidenceMedium)}}
		}
	}
	return ConfidenceUnknown, nil
}
