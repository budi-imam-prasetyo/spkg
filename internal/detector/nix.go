package detector

import (
	"path/filepath"
	"strings"
)

type NixDetector struct{}

func init() { Register(&NixDetector{}) }

func (d *NixDetector) Name() Manager { return ManagerNix }

func (d *NixDetector) IsAvailable() bool {
	return HasCommand("nix")
}

func (d *NixDetector) Detect(execPath string) *Result {
	list, err := RunCmd("nix", "profile", "list")
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(list, "\n") {
		if strings.Contains(line, execPath) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				pkgName := parts[1]
				return &Result{
					Found:          true,
					ExecutablePath: execPath,
					Manager:        ManagerNix,
					Package:        pkgName,
					InstallCmd:     "nix profile install nixpkgs#" + pkgName,
					UpgradeCmd:     "nix profile upgrade " + pkgName,
					RemoveCmd:      "nix profile remove " + pkgName,
				}
			}
		}
	}

	nixStore := "/nix/store"
	if strings.HasPrefix(execPath, nixStore) {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerNix,
			Package:        filepath.Base(execPath),
			InstallCmd:     "nix profile install nixpkgs#" + filepath.Base(execPath),
			UpgradeCmd:     "nix profile upgrade " + filepath.Base(execPath),
			RemoveCmd:      "nix profile remove " + filepath.Base(execPath),
		}
	}

	return nil
}
