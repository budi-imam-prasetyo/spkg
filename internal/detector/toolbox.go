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
	if !strings.Contains(execPath, "toolbox") {
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
