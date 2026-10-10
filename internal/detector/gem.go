package detector

import (
	"path/filepath"
	"strings"
)

type GemDetector struct{}

func init() { Register(&GemDetector{}) }

func (d *GemDetector) Name() Manager { return ManagerGem }

func (d *GemDetector) IsAvailable() bool {
	return HasCommand("gem")
}

func (d *GemDetector) Detect(execPath string) *Result {
	base := filepath.Base(execPath)

	// If detecting gem itself, show how to install/remove Ruby/gem
	if base == "gem" || base == "ruby" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerGem,
			Package:        "ruby",
			InstallCmd:     "gem update --system",
			UpgradeCmd:     "gem update --system",
			RemoveCmd:      "# RubyGems is normally installed with Ruby",
		}
	}

	list, err := gemListDetail()
	if err != nil {
		return nil
	}

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
			// Exact match on the Executables list only: split "Executables: a, b"
			// and compare tokens. Matching arbitrary substrings of the install
			// path produced false positives.
			if strings.Contains(line, "Executables:") && matchExecutablesLine(line, base) {
				return &Result{
					Found:          true,
					ExecutablePath: execPath,
					Manager:        ManagerGem,
					Package:        pkgName,
					Version:        version,
					InstallCmd:     "gem install " + pkgName,
					UpgradeCmd:     "gem install " + pkgName,
					RemoveCmd:      "gem uninstall " + pkgName,
					ConfirmReasons: []Reason{{
						Signal: "manager_db",
						Detail: "gem list confirms executable " + base + " in package " + pkgName,
						Weight: int(ConfidenceHigh),
					}},
				}
			}
		}
	}

	return nil
}

func (d *GemDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	gemDir := filepath.Join(home, ".gem")
	if strings.HasPrefix(execPath, gemDir) || strings.HasPrefix(execPath, "/usr/lib/ruby/") || strings.Contains(execPath, "gems") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is under a ruby gem directory", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
