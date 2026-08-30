package detector

import (
	"path/filepath"
	"strings"
)

type ToolboxDetector struct{}

func init() { Register(&ToolboxDetector{}) }

func (d *ToolboxDetector) Name() Manager { return ManagerToolbox }

func (d *ToolboxDetector) IsAvailable() bool {
	return HasCommand("toolbox")
}

func (d *ToolboxDetector) Detect(execPath string) *Result {
	// toolbox binaries live inside a container, exposed via specific run/overlay
	// locations on the host. Avoid claiming any binary whose path merely contains
	// the word "toolbox" (e.g. /home/u/toolbox-scripts/foo).
	if !strings.HasPrefix(execPath, "/run/toolbox/") &&
		!strings.HasPrefix(execPath, "/var/lib/toolbox/") {
		return nil
	}
	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerToolbox,
		Package:        filepath.Base(execPath),
		InstallCmd:     "toolbox run -- <package-manager> install <package>",
		UpgradeCmd:     "toolbox run -- <package-manager> upgrade <package>",
		RemoveCmd:      "toolbox rm <container>",
	}
}

func (d *ToolboxDetector) Score(execPath string) (Confidence, []Reason) {
	if strings.HasPrefix(execPath, "/run/toolbox/") || strings.HasPrefix(execPath, "/var/lib/toolbox/") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under a toolbox container root", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
