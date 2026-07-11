package detector

import (
	"os"
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

	show, err := RunCmd(pipCmd, "show", filepath.Base(execPath))
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

func detectPipExecInSite(execPath string) bool {
	home, _ := os.UserHomeDir()
	sites := []string{
		filepath.Join(home, ".local", "lib"),
		"/usr/lib",
		"/usr/local/lib",
	}
	for _, site := range sites {
		matches, _ := filepath.Glob(filepath.Join(site, "python*", "site-packages"))
		for _, m := range matches {
			binDir := filepath.Join(filepath.Dir(m), "bin")
			if _, err := os.Stat(filepath.Join(binDir, filepath.Base(execPath))); err == nil {
				return true
			}
		}
	}
	return false
}
