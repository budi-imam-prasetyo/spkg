package detector

import (
	"path/filepath"
	"strings"
)

type UvDetector struct{}

func init() { Register(&UvDetector{}) }

func (d *UvDetector) Name() Manager { return ManagerUv }

func (d *UvDetector) IsAvailable() bool {
	return HasCommand("uv")
}

func (d *UvDetector) Detect(execPath string) *Result {
	home := HomeDir()
	uvBin := filepath.Join(home, ".local", "bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, uvBin) {
		return nil
	}

	show, err := RunCmd("uv", "tool", "list")
	if err != nil {
		return nil
	}

	base := filepath.Base(execPath)
	for _, line := range strings.Split(show, "\n") {
		if strings.Contains(line, base) {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				pkgName := parts[0]
				// Require the binary to live in uv's tool bin dir
				// (~/.local/bin, shared with pipx) before trusting the match.
				uvToolBin := filepath.Join(home, ".local", "bin")
				if !strings.HasPrefix(execPath, uvToolBin) {
					continue
				}
				return &Result{
					Found:          true,
					ExecutablePath: execPath,
					Manager:        ManagerUv,
					Package:        pkgName,
					InstallCmd:     "uv tool install " + pkgName,
					UpgradeCmd:     "uv tool upgrade " + pkgName,
					RemoveCmd:      "uv tool uninstall " + pkgName,
				}
			}
		}
	}

	return nil
}

func (d *UvDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	uvToolBin := filepath.Join(home, ".local", "bin")
	if strings.HasPrefix(execPath, uvToolBin) {
		return ConfidenceLow, []Reason{{Signal: "ambiguous_path", Detail: "path is in uv's shared tool bin dir", Weight: int(ConfidenceLow)}}
	}
	return ConfidenceUnknown, nil
}
