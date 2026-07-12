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
			InstallCmd:     "sudo pacman -S python-pip  # or: sudo apt install python3-pip",
			UpgradeCmd:     "sudo pacman -Syu python-pip",
			RemoveCmd:      "sudo pacman -Rns python-pip  # or: sudo apt remove python3-pip",
		}
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
	}
}
