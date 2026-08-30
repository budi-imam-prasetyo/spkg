package detector

import (
	"path/filepath"
	"strings"
)

type CargoDetector struct{}

func init() { Register(&CargoDetector{}) }

func (d *CargoDetector) Name() Manager { return ManagerCargo }

func (d *CargoDetector) IsAvailable() bool {
	return HasCommand("cargo")
}

func (d *CargoDetector) Detect(execPath string) *Result {
	home := HomeDir()
	cargoBin := filepath.Join(home, ".cargo", "bin") + string(filepath.Separator)
	if !strings.HasPrefix(execPath, cargoBin) {
		return nil
	}

	base := filepath.Base(execPath)

	// If detecting cargo/rustc itself, show how to install/remove Rust
	if base == "cargo" || base == "rustc" || base == "rustup" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerCargo,
			Package:        "rustup",
			Version:        detectRustVersion(),
			InstallCmd:     "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
			UpgradeCmd:     "rustup update",
			RemoveCmd:      "rustup self uninstall  # then: rm -rf ~/.cargo ~/.rustup",
		}
	}

	list, err := RunCmd("cargo", "install", "--list")
	if err != nil {
		return nil
	}

	lines := strings.Split(list, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ":") {
			pkgName := strings.TrimSuffix(line, ":")
			for j := i + 1; j < len(lines) && j < i+5; j++ {
				if strings.Contains(lines[j], filepath.Join(cargoBin, base)) {
					parts := strings.Fields(pkgName)
					name := parts[0]
					version := ""
					if len(parts) > 1 {
						version = strings.Trim(parts[1], "v")
					}
					return &Result{
						Found:          true,
						ExecutablePath: execPath,
						Manager:        ManagerCargo,
						Package:        name,
						Version:        version,
						InstallCmd:     "cargo install " + name,
						UpgradeCmd:     "cargo install " + name,
						RemoveCmd:      "cargo uninstall " + name,
					}
				}
			}
		}
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerCargo,
		Package:        base,
		InstallCmd:     "cargo install " + base,
		UpgradeCmd:     "cargo install " + base,
		RemoveCmd:      "cargo uninstall " + base,
	}
}

func detectRustVersion() string {
	out, _ := RunCmd("rustc", "--version")
	return out
}

func (d *CargoDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	binDir := filepath.Join(home, ".cargo", "bin")
	if strings.HasPrefix(execPath, binDir) || strings.HasPrefix(execPath, "/usr/local/cargo/bin") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is in cargo bin dir", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
