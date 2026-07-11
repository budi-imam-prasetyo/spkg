package detector

import (
	"strings"
)

type SnapDetector struct{}

func init() { Register(&SnapDetector{}) }

func (d *SnapDetector) Name() Manager { return ManagerSnap }

func (d *SnapDetector) IsAvailable() bool {
	return HasCommand("snap")
}

func (d *SnapDetector) Detect(execPath string) *Result {
	list, err := RunCmd("snap", "list")
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

		info, _ := RunCmd("snap", "info", name)
		if strings.Contains(info, execPath) || strings.Contains(info, "/snap/"+name) {
			return &Result{
				Found:          true,
				ExecutablePath: execPath,
				Manager:        ManagerSnap,
				Package:        name,
				Version:        version,
				InstallCmd:     "sudo snap install " + name,
				UpgradeCmd:     "sudo snap refresh " + name,
				RemoveCmd:      "sudo snap remove " + name,
			}
		}
	}

	return nil
}
