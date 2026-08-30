package detector

import (
	"path/filepath"
	"strings"
)

type BrewDetector struct{}

func init() { Register(&BrewDetector{}) }

func (d *BrewDetector) Name() Manager { return ManagerBrew }

func (d *BrewDetector) IsAvailable() bool {
	return HasCommand("brew")
}

func (d *BrewDetector) Detect(execPath string) *Result {
	prefix, _ := RunCmd("brew", "--prefix")
	if prefix == "" {
		return nil
	}
	binDir := filepath.Join(prefix, "bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, binDir) {
		return nil
	}

	base := filepath.Base(execPath)
	info, err := RunCmd("brew", "info", "--json=v2", base)
	if err != nil {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerBrew,
			Package:        base,
			InstallCmd:     "brew install " + base,
			UpgradeCmd:     "brew upgrade " + base,
			RemoveCmd:      "brew uninstall " + base,
		}
	}

	version := extractBrewVersion(info)

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerBrew,
		Package:        base,
		Version:        version,
		InstallCmd:     "brew install " + base,
		UpgradeCmd:     "brew upgrade " + base,
		RemoveCmd:      "brew uninstall " + base,
	}
}

func extractBrewVersion(info string) string {
	idx := strings.Index(info, "\"stable\"")
	if idx == -1 {
		return ""
	}
	verIdx := strings.Index(info[idx:], "\"version\"")
	if verIdx == -1 {
		return ""
	}
	sub := info[idx+verIdx:]
	lines := strings.SplitN(sub, "\n", 3)
	if len(lines) == 0 {
		return ""
	}
	parts := strings.SplitN(lines[0], ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.Trim(parts[1], " \",")
}

func (d *BrewDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	prefixes := []string{"/usr/local", "/opt/homebrew", filepath.Join(home, "homebrew")}
	for _, p := range prefixes {
		if strings.HasPrefix(execPath, filepath.Join(p, "bin")) || strings.HasPrefix(execPath, filepath.Join(p, "Cellar")) || strings.HasPrefix(execPath, filepath.Join(p, "opt")) {
			return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under a homebrew prefix", Weight: int(ConfidenceMedium)}}
		}
	}
	return ConfidenceUnknown, nil
}
