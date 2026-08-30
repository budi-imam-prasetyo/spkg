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

	// Detect which AUR helper was used to install this package
	aurHelper := ""
	installCmd := "sudo pacman -S " + pkgName
	upgradeCmd := "sudo pacman -Syu " + pkgName
	removeCmd := "sudo pacman -Rns " + pkgName

	if isAUR {
		// Try to detect AUR helper by checking if they're installed
		// Priority: paru > yay > pikaur > pakku
		for _, helper := range []string{"paru", "yay", "pikaur", "pakku"} {
			if _, err := RunCmd("which", helper); err == nil {
				aurHelper = helper
				installCmd = helper + " -S " + pkgName
				upgradeCmd = helper + " -Syu " + pkgName
				removeCmd = helper + " -Rns " + pkgName
				break
			}
		}
	}

	// An AUR package installed by a helper (paru/yay/...) is owned by that
	// helper, not by pacman itself. Report the helper as the manager so batch
	// (Source) and single (Install/Upgrade/Remove) views agree, instead of
	// showing "pacman" with "paru -S" style commands.
	manager := ManagerPacman
	if isAUR && aurHelper != "" {
		manager = Manager(aurHelper)
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        manager,
		Package:        pkgName,
		Version:        version,
		Repository:     repo,
		IsAUR:          isAUR,
		AurHelper:      aurHelper,
		InstallCmd:     installCmd,
		UpgradeCmd:     upgradeCmd,
		RemoveCmd:      removeCmd,
		ConfirmReasons: []Reason{{
			Signal: "manager_db",
			Detail: "pacman -Qo confirms ownership by package " + pkgName,
			Weight: int(ConfidenceHigh),
		}},
	}
}

func (d *PacmanDetector) Score(execPath string) (Confidence, []Reason) {
	// pacman owns anything under the system root that isn't a known user/alt
	// domain. We give a medium structural floor; confirmation comes from -Qo.
	if strings.HasPrefix(execPath, "/usr/bin/") || strings.HasPrefix(execPath, "/usr/lib/") || strings.HasPrefix(execPath, "/bin/") || strings.HasPrefix(execPath, "/opt/") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under a system install dir (pacman domain)", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
