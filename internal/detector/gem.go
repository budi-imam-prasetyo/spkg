package detector

import (
	"strings"
)

type GemDetector struct{}

func init() { Register(&GemDetector{}) }

func (d *GemDetector) Name() Manager { return ManagerGem }

func (d *GemDetector) IsAvailable() bool {
	return HasCommand("gem")
}

func (d *GemDetector) Detect(execPath string) *Result {
	list, err := RunCmd("gem", "list", "-d")
	if err != nil {
		return nil
	}

	base := strings.TrimSuffix(strings.TrimPrefix(execPath, "/usr/bin/"), "")
	for _, block := range strings.Split(list, "\n\n") {
		lines := strings.Split(block, "\n")
		if len(lines) < 1 {
			continue
		}
		header := strings.TrimSpace(lines[0])
		parts := strings.Fields(header)
		if len(parts) < 1 {
			continue
		}
		pkgName := parts[0]
		version := ""
		if len(parts) > 1 {
			version = strings.Trim(parts[1], "()")
		}
		for _, line := range lines[1:] {
			if strings.Contains(line, execPath) || strings.Contains(line, base) {
				return &Result{
					Found:          true,
					ExecutablePath: execPath,
					Manager:        ManagerGem,
					Package:        pkgName,
					Version:        version,
					InstallCmd:     "gem install " + pkgName,
					UpgradeCmd:     "gem install " + pkgName,
					RemoveCmd:      "gem uninstall " + pkgName,
				}
			}
		}
	}

	return nil
}
