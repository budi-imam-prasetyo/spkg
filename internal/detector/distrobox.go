package detector

import (
	"path/filepath"
	"strings"
)

type DistroboxDetector struct{}

func init() { Register(&DistroboxDetector{}) }

func (d *DistroboxDetector) Name() Manager { return ManagerDistrobox }

func (d *DistroboxDetector) IsAvailable() bool {
	return HasCommand("distrobox")
}

func (d *DistroboxDetector) Detect(execPath string) *Result {
	if !strings.Contains(execPath, "distrobox") {
		return nil
	}
	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerDistrobox,
		Package:        filepath.Base(execPath),
		InstallCmd:     "distrobox-enter -- <package-manager> install <package>",
		UpgradeCmd:     "distrobox-enter -- <package-manager> upgrade <package>",
		RemoveCmd:      "distrobox-rm <container>",
	}
}
