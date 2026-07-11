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
	home, _ := os.UserHomeDir()
	bunBin := filepath.Join(home, ".bun", "bin")

	if !strings.HasPrefix(execPath, bunBin) {
		return nil
	}

	globalDir := filepath.Join(home, ".bun", "install", "global", "node_modules")
	_, err := RunCmd("bun", "pm", "ls", "--global")
	if err != nil {
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerBun,
			Package:        filepath.Base(execPath),
			Global:         BoolPtr(true),
			InstallCmd:     "bun add -g " + filepath.Base(execPath),
			UpgradeCmd:     "bun update -g " + filepath.Base(execPath),
			RemoveCmd:      "bun remove -g " + filepath.Base(execPath),
		}
	}

	pkgName := detectGlobalPkgFromDir(globalDir, execPath)
	if pkgName == "" {
		pkgName = filepath.Base(execPath)
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
