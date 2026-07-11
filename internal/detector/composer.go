package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type ComposerDetector struct{}

func init() { Register(&ComposerDetector{}) }

func (d *ComposerDetector) Name() Manager { return ManagerComposer }

func (d *ComposerDetector) IsAvailable() bool {
	return HasCommand("composer")
}

func (d *ComposerDetector) Detect(execPath string) *Result {
	globalDir, _ := RunCmd("composer", "global", "config", "home")
	if globalDir == "" {
		home, _ := os.UserHomeDir()
		globalDir = filepath.Join(home, ".composer")
	}

	binDir := filepath.Join(globalDir, "vendor", "bin")
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)
	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerComposer,
		Package:        base,
		Global:         BoolPtr(true),
		InstallCmd:     "composer global require " + base,
		UpgradeCmd:     "composer global update " + base,
		RemoveCmd:      "composer global remove " + base,
	}
}
