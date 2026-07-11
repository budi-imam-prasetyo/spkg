package detector

import (
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
		globalDir = filepath.Join(HomeDir(), ".composer")
	}

	binDir := filepath.Join(globalDir, "vendor", "bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)

	// If detecting composer itself, show how to install/remove Composer
	if base == "composer" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerComposer,
			Package:        "composer",
			InstallCmd:     "sudo pacman -S composer  # or: curl -sS https://getcomposer.org/installer | php",
			UpgradeCmd:     "sudo pacman -Syu composer  # or: composer self-update",
			RemoveCmd:      "sudo pacman -Rns composer  # or: rm /usr/local/bin/composer",
		}
	}

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
