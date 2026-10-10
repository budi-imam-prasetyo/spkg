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
	globalDir := composerGlobalHome()
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
			InstallCmd:     "# Install Composer from https://getcomposer.org/download/",
			UpgradeCmd:     "composer self-update",
			RemoveCmd:      "composer self-update --clean-backups",
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
		ConfirmReasons: []Reason{{
			Signal: "manager_db",
			Detail: "composer global vendor bin owns this path",
			Weight: int(ConfidenceHigh),
		}},
	}
}

func (d *ComposerDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	vendor := filepath.Join(home, ".config", "composer", "vendor")
	if strings.HasPrefix(execPath, vendor) || strings.HasPrefix(execPath, "/usr/local/bin") {
		return ConfidenceLow, []Reason{{Signal: "ambiguous_path", Detail: "global composer bin dir overlaps with system bin", Weight: int(ConfidenceLow)}}
	}
	return ConfidenceUnknown, nil
}
