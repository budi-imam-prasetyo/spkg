package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type BunDetector struct{}

func init() { Register(&BunDetector{}) }

func (d *BunDetector) Name() Manager { return ManagerBun }

func (d *BunDetector) IsAvailable() bool {
	return HasCommand("bun")
}

func (d *BunDetector) Detect(execPath string) *Result {
	home := HomeDir()
	bunBin := filepath.Join(home, ".bun", "bin") + string(filepath.Separator)

	if !strings.HasPrefix(execPath, bunBin) {
		return nil
	}

	base := filepath.Base(execPath)

	// If detecting bun itself, show how to install/remove bun
	if base == "bun" {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerBun,
			Package:        "bun",
			Global:         BoolPtr(true),
			InstallCmd:     "curl -fsSL https://bun.sh/install | bash",
			UpgradeCmd:     "bun upgrade",
			RemoveCmd:      "rm -rf ~/.bun",
		}
	}

	globalDir := filepath.Join(home, ".bun", "install", "global", "node_modules")
	_, err := bunGlobalList()
	if err != nil {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerBun,
			Package:        base,
			Global:         BoolPtr(true),
			InstallCmd:     "bun add -g " + base,
			UpgradeCmd:     "bun update -g " + base,
			RemoveCmd:      "bun remove -g " + base,
		}
	}

	pkgName := detectGlobalPkgFromDir(globalDir, execPath)
	if pkgName == "" {
		pkgName = base
	}

	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerBun,
		Package:        pkgName,
		Global:         BoolPtr(true),
		InstallCmd:     "bun add -g " + pkgName,
		UpgradeCmd:     "bun update -g " + pkgName,
		RemoveCmd:      "bun remove -g " + pkgName,
		ConfirmReasons: []Reason{{
			Signal: "manager_db",
			Detail: "bun global dir owns this path",
			Weight: int(ConfidenceHigh),
		}},
	}
}

func detectGlobalPkgFromDir(globalDir, execPath string) string {
	base := filepath.Base(execPath)
	entries, err := os.ReadDir(globalDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		binDir := filepath.Join(globalDir, e.Name(), "bin")
		if _, err := os.Stat(filepath.Join(binDir, base)); err == nil {
			return e.Name()
		}
		binDir2 := filepath.Join(globalDir, e.Name(), ".bin")
		if _, err := os.Stat(filepath.Join(binDir2, base)); err == nil {
			return e.Name()
		}
	}
	return ""
}

func (d *BunDetector) Score(execPath string) (Confidence, []Reason) {
	home := HomeDir()
	bunBin := filepath.Join(home, ".bun", "bin")
	if strings.HasPrefix(execPath, bunBin) {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is in bun's bin dir", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
