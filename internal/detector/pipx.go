package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type PipxDetector struct{}

func init() { Register(&PipxDetector{}) }

func (d *PipxDetector) Name() Manager { return ManagerPipx }

func (d *PipxDetector) IsAvailable() bool {
	return HasCommand("pipx")
}

func (d *PipxDetector) Detect(execPath string) *Result {
	home := HomeDir()
	localBin := filepath.Join(home, ".local", "bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, localBin) {
		return nil
	}

	list, err := RunCmd("pipx", "list", "--short")
	if err != nil {
		return nil
	}

	base := filepath.Base(execPath)
	for _, line := range strings.Split(list, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 1 {
			continue
		}
		pkgName := strings.TrimRight(parts[0], ":")
		venv := filepath.Join(home, ".local", "pipx", "venvs", pkgName)
		if _, err := os.Stat(venv); err == nil {
			binDir := filepath.Join(venv, "bin")
			if _, err := os.Stat(filepath.Join(binDir, base)); err == nil {
				return &Result{
					Found:          true,
					ExecutablePath: execPath,
					Manager:        ManagerPipx,
					Package:        pkgName,
					VirtualEnv:     venv,
					InstallCmd:     "pipx install " + pkgName,
					UpgradeCmd:     "pipx upgrade " + pkgName,
					RemoveCmd:      "pipx uninstall " + pkgName,
				}
			}
		}
	}

	pipxDetect, _ := RunCmd("pipx", "list")
	for _, line := range strings.Split(pipxDetect, "\n") {
		if strings.Contains(line, base) {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				pkgName := strings.TrimRight(parts[0], ":")
				venv := filepath.Join(home, ".local", "pipx", "venvs", pkgName)
				// Only trust this match if the binary actually lives inside a
				// pipx-managed venv (pipx/uv/brew share ~/.local/bin).
				if _, err := os.Stat(filepath.Join(venv, "bin", base)); err == nil {
					return &Result{
						Found:          true,
						ExecutablePath: execPath,
						Manager:        ManagerPipx,
						Package:        pkgName,
						VirtualEnv:     venv,
						InstallCmd:     "pipx install " + pkgName,
						UpgradeCmd:     "pipx upgrade " + pkgName,
						RemoveCmd:      "pipx uninstall " + pkgName,
					}
				}
			}
		}
	}

	return nil
}

func (d *PipxDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	venvBin := filepath.Join(home, ".local", "pipx", "venvs")
	if strings.HasPrefix(execPath, venvBin) {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is inside a pipx venv", Weight: int(ConfidenceMedium)}}
	}
	localBin := filepath.Join(home, ".local", "bin")
	if strings.HasPrefix(execPath, localBin) {
		return ConfidenceLow, []Reason{{Signal: "ambiguous_path", Detail: "path is in pipx's shared bin dir", Weight: int(ConfidenceLow)}}
	}
	return ConfidenceUnknown, nil
}
