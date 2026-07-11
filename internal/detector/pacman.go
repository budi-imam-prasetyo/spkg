package detector

import (
	"os"
	"strings"
)

type PacmanDetector struct{}

func init() { Register(&PacmanDetector{}) }

func (d *PacmanDetector) Name() Manager { return ManagerPacman }

func (d *PacmanDetector) IsAvailable() bool {
	_, err := os.Stat("/usr/bin/pacman")
	return err == nil
}

func (d *PacmanDetector) Detect(execPath string) *Result {
	out, err := RunCmd("pacman", "-Qo", execPath)
	if err != nil {
		return nil
	}
	// Output: /usr/bin/git is owned by git 2.55.0-1.1
	lines := strings.SplitN(out, "\n", 2)
	if len(lines) < 1 {
		return nil
	}
	parts := strings.Fields(lines[0])
	// Need at least: path "is" "owned" "by" name version
	if len(parts) < 6 {
		return nil
	}
	pkgName := parts[len(parts)-2]
	version := parts[len(parts)-1]

	qi, _ := RunCmd("pacman", "-Qi", pkgName)
	repo := ""
	isAUR := false
	for _, line := range strings.Split(qi, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version") && strings.Contains(line, ":") {
			version = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
		if strings.HasPrefix(line, "Repository") && strings.Contains(line, ":") {
			repo = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}

	if repo == "aur" || repo == "" {
		aurCheck, _ := RunCmd("pacman", "-Qm", pkgName)
		if aurCheck != "" && !strings.HasPrefix(aurCheck, "error") {
			isAUR = true
			if repo == "" {
				repo = "aur"
			}
		}
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerPacman,
		Package:        pkgName,
		Version:        version,
		Repository:     repo,
		IsAUR:          isAUR,
		InstallCmd:     "sudo pacman -S " + pkgName,
		UpgradeCmd:     "sudo pacman -Syu " + pkgName,
		RemoveCmd:      "sudo pacman -Rns " + pkgName,
	}
}
