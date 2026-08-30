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
	// distrobox/toolbox binaris only exist inside a container's mount
	// namespace; from the host they're exposed via specific overlay/run
	// locations. Avoid claiming any binary whose path merely contains the word
	// "distrobox" (e.g. /home/u/distrobox-scripts/foo).
	if !strings.HasPrefix(execPath, "/run/distrobox/") &&
		!strings.HasPrefix(execPath, "/var/lib/distrobox/") {
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

func (d *DistroboxDetector) Score(execPath string) (Confidence, []Reason) {
	if strings.HasPrefix(execPath, "/run/distrobox/") || strings.HasPrefix(execPath, "/var/lib/distrobox/") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under a distrobox container root", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
