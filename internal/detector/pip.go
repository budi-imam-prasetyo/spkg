package detector

import (
	"path/filepath"
	"strings"
)

type PipDetector struct{}

func init() { Register(&PipDetector{}) }

func (d *PipDetector) Name() Manager { return ManagerPip }

func (d *PipDetector) IsAvailable() bool {
	return HasCommand("pip") || HasCommand("pip3")
}

func (d *PipDetector) Detect(execPath string) *Result {
	pipCmd := "pip"
	if !HasCommand("pip") {
		pipCmd = "pip3"
	}

	base := filepath.Base(execPath)

	// If detecting pip itself, show how to install/remove pip
	if base == "pip" || base == "pip3" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerPip,
			Package:        "python-pip",
			InstallCmd:     "python -m ensurepip --upgrade",
			UpgradeCmd:     "python -m pip install --upgrade pip",
			RemoveCmd:      "python -m pip uninstall pip",
		}
	}

	// Only claim ownership if the executable lives in a directory pip actually
	// manages (a venv / virtualenv / user scheme bin). Without this guard, a
	// binary whose *name* matches a pip package (but is installed by pipx, uv,
	// brew, conda, ...) would be falsely attributed to pip.
	if !pipOwnsPath(execPath) {
		return nil
	}

	show, err := RunCmd(pipCmd, "show", base)
	if err != nil {
		return nil
	}

	name := ""
	version := ""
	for _, line := range strings.Split(show, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Name":
			name = val
		case "Version":
			version = val
		}
	}

	if name == "" {
		return nil
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerPip,
		Package:        name,
		Version:        version,
		InstallCmd:     pipCmd + " install " + name,
		UpgradeCmd:     pipCmd + " install --upgrade " + name,
		RemoveCmd:      pipCmd + " uninstall " + name,
		ConfirmReasons: []Reason{{
			Signal: "manager_db",
			Detail: "pip show confirms package " + name,
			Weight: int(ConfidenceHigh),
		}},
	}
}

// pipOwnsPath reports whether execPath sits under a directory controlled by a
// Python environment (venv, virtualenv, or a user scheme bin). pip/pipx/uv share
// the same ~/.local/bin, so callers must check this before attributing a binary
// to pip by name alone.
func pipOwnsPath(execPath string) bool {
	home := HomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".local", "bin", ""),
		filepath.Join(home, ".local", "share", "python-wheels"),
		"/usr/local/bin",
		"/usr/bin",
	}
	for _, c := range candidates {
		if c != "" && strings.HasPrefix(execPath, c) {
			return true
		}
	}
	// Any .../venv/bin or .../site-packages or .../bin under a python env.
	if strings.Contains(execPath, "site-packages") || strings.Contains(execPath, "bin/python") {
		return true
	}
	return false
}

func (d *PipDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "bin"),
		"/usr/local/bin",
		"/usr/bin",
	}
	for _, c := range candidates {
		if strings.HasPrefix(execPath, c) {
			return ConfidenceLow, []Reason{{Signal: "ambiguous_path", Detail: "path is in a dir pip may manage (shared with pipx/uv/brew)", Weight: int(ConfidenceLow)}}
		}
	}
	return ConfidenceUnknown, nil
}
