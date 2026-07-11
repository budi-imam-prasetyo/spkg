package detector

import (
	"os"
	"path/filepath"
	"strings"
)

// ─── JavaScript Runtime Detectors ────────────────────────────────────────────

type FnmDetector struct{}

func init() { RegisterRuntime(&FnmDetector{}) }

func (d *FnmDetector) Name() RuntimeType { return RuntimeFnm }
func (d *FnmDetector) IsAvailable() bool { return HasCommand("fnm") }
func (d *FnmDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	fnmDir := filepath.Join(home, ".local", "share", "fnm", "node-versions")
	if !strings.HasPrefix(execPath, fnmDir) {
		return RuntimeNone, "", "", 0
	}
	version := ExtractVersionFromPath(execPath, "node-versions/v", "node-versions/")
	if version == "" {
		return RuntimeNone, "", "", 0
	}
	return RuntimeFnm, "Node.js " + version, "fnm", 95
}

type NvmDetector struct{}

func init() { RegisterRuntime(&NvmDetector{}) }

func (d *NvmDetector) Name() RuntimeType { return RuntimeNvm }
func (d *NvmDetector) IsAvailable() bool { return HasCommand("nvm") || os.Getenv("NVM_DIR") != "" }
func (d *NvmDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		nvmDir = filepath.Join(home, ".nvm")
	}
	nvmNode := filepath.Join(nvmDir, "versions", "node")
	if !strings.HasPrefix(execPath, nvmNode) {
		return RuntimeNone, "", "", 0
	}
	version := ExtractVersionFromPath(execPath, "versions/node/v")
	if version == "" {
		return RuntimeNone, "", "", 0
	}
	return RuntimeNvm, "Node.js " + version, "nvm", 95
}

type VoltaDetector struct{}

func init() { RegisterRuntime(&VoltaDetector{}) }

func (d *VoltaDetector) Name() RuntimeType { return RuntimeVolta }
func (d *VoltaDetector) IsAvailable() bool { return HasCommand("volta") }
func (d *VoltaDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	voltaDir := filepath.Join(home, ".volta", "tools")
	if !strings.HasPrefix(execPath, voltaDir) {
		return RuntimeNone, "", "", 0
	}
	// ~/.volta/tools/image/node/{version}/bin/node
	if strings.Contains(execPath, voltaDir+"/image/node/") {
		version := ExtractVersionFromPath(execPath, "image/node/")
		if version != "" {
			return RuntimeVolta, "Node.js " + version, "volta", 95
		}
	}
	// ~/.volta/tools/image/packages/{org}/{pkg}/{version}/...
	if strings.Contains(execPath, voltaDir+"/image/packages/") {
		return RuntimeVolta, "", "volta", 80
	}
	return RuntimeNone, "", "", 0
}

type BunRuntimeDetector struct{}

func init() { RegisterRuntime(&BunRuntimeDetector{}) }

func (d *BunRuntimeDetector) Name() RuntimeType { return RuntimeBun }
func (d *BunRuntimeDetector) IsAvailable() bool { return HasCommand("bun") }
func (d *BunRuntimeDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	bunBin := filepath.Join(home, ".bun", "bin")
	if !strings.HasPrefix(execPath, bunBin) {
		return RuntimeNone, "", "", 0
	}
	base := filepath.Base(execPath)
	if base != "bun" {
		return RuntimeNone, "", "", 0
	}
	out, _ := RunCmd("bun", "--version")
	return RuntimeBun, "Bun " + out, "bun", 90
}

// ─── Python Runtime Detectors ────────────────────────────────────────────────

type PyenvDetector struct{}

func init() { RegisterRuntime(&PyenvDetector{}) }

func (d *PyenvDetector) Name() RuntimeType { return RuntimePyenv }
func (d *PyenvDetector) IsAvailable() bool { return HasCommand("pyenv") }
func (d *PyenvDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	pyenvDir := filepath.Join(home, ".pyenv", "versions")
	if !strings.HasPrefix(execPath, pyenvDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, pyenvDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 {
		return RuntimeNone, "", "", 0
	}
	version := parts[0]
	return RuntimePyenv, "Python " + version, "pyenv", 95
}

type CondaDetector struct{}

func init() { RegisterRuntime(&CondaDetector{}) }

func (d *CondaDetector) Name() RuntimeType { return RuntimeConda }
func (d *CondaDetector) IsAvailable() bool { return HasCommand("conda") }
func (d *CondaDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	condaEnvDir := filepath.Join(home, "miniconda3", "envs")
	condaEnvDir2 := filepath.Join(home, "anaconda3", "envs")
	condaEnvDir3 := filepath.Join(home, "miniforge3", "envs")
	condaEnvDir4 := filepath.Join(home, "mambaforge", "envs")

	for _, dir := range []string{condaEnvDir, condaEnvDir2, condaEnvDir3, condaEnvDir4} {
		if strings.HasPrefix(execPath, dir) {
			rest := strings.TrimPrefix(execPath, dir+"/")
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) > 0 {
				envName := parts[0]
				return RuntimeConda, "conda env:" + envName, "conda", 90
			}
		}
	}

	// Check for conda base env
	for _, prefix := range []string{
		filepath.Join(home, "miniconda3"),
		filepath.Join(home, "anaconda3"),
		filepath.Join(home, "miniforge3"),
		filepath.Join(home, "mambaforge"),
	} {
		if strings.HasPrefix(execPath, filepath.Join(prefix, "bin")) {
			return RuntimeConda, "conda base", "conda", 85
		}
	}

	return RuntimeNone, "", "", 0
}

type MambaDetector struct{}

func init() { RegisterRuntime(&MambaDetector{}) }

func (d *MambaDetector) Name() RuntimeType { return RuntimeMamba }
func (d *MambaDetector) IsAvailable() bool { return HasCommand("mamba") }
func (d *MambaDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	for _, prefix := range []string{
		filepath.Join(home, "mambaforge"),
		filepath.Join(home, "miniforge3"),
	} {
		if strings.HasPrefix(execPath, filepath.Join(prefix, "envs")) {
			rest := strings.TrimPrefix(execPath, filepath.Join(prefix, "envs")+"/")
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) > 0 {
				return RuntimeMamba, "mamba env:" + parts[0], "mamba", 90
			}
		}
		if strings.HasPrefix(execPath, filepath.Join(prefix, "bin")) {
			return RuntimeMamba, "mamba base", "mamba", 85
		}
	}
	return RuntimeNone, "", "", 0
}

type MicromambaDetector struct{}

func init() { RegisterRuntime(&MicromambaDetector{}) }

func (d *MicromambaDetector) Name() RuntimeType { return RuntimeMicromamba }
func (d *MicromambaDetector) IsAvailable() bool { return HasCommand("micromamba") }
func (d *MicromambaDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	mmDir := filepath.Join(home, ".local", "share", "micromamba")
	if strings.HasPrefix(execPath, filepath.Join(mmDir, "envs")) {
		rest := strings.TrimPrefix(execPath, filepath.Join(mmDir, "envs")+"/")
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) > 0 {
			return RuntimeMicromamba, "micromamba env:" + parts[0], "micromamba", 90
		}
	}
	return RuntimeNone, "", "", 0
}

type PixiDetector struct{}

func init() { RegisterRuntime(&PixiDetector{}) }

func (d *PixiDetector) Name() RuntimeType { return RuntimePixi }
func (d *PixiDetector) IsAvailable() bool { return HasCommand("pixi") }
func (d *PixiDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	// pixi uses project-local .pixi dirs
	if strings.Contains(execPath, ".pixi") {
		return RuntimePixi, "pixi env", "pixi", 85
	}
	return RuntimeNone, "", "", 0
}

type RyeDetector struct{}

func init() { RegisterRuntime(&RyeDetector{}) }

func (d *RyeDetector) Name() RuntimeType { return RuntimeRye }
func (d *RyeDetector) IsAvailable() bool { return HasCommand("rye") }
func (d *RyeDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	ryeDir := filepath.Join(home, ".rye")
	if strings.HasPrefix(execPath, ryeDir) {
		return RuntimeRye, "rye managed", "rye", 85
	}
	return RuntimeNone, "", "", 0
}

type PoetryDetector struct{}

func init() { RegisterRuntime(&PoetryDetector{}) }

func (d *PoetryDetector) Name() RuntimeType { return RuntimePoetry }
func (d *PoetryDetector) IsAvailable() bool { return HasCommand("poetry") }
func (d *PoetryDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	poetryDir := filepath.Join(home, "Library", "Caches", "pypoetry", "virtualenvs")
	poetryDir2 := filepath.Join(home, ".cache", "pypoetry", "virtualenvs")
	for _, dir := range []string{poetryDir, poetryDir2} {
		if strings.HasPrefix(execPath, dir) {
			return RuntimePoetry, "poetry venv", "poetry", 80
		}
	}
	return RuntimeNone, "", "", 0
}

// ─── Ruby Runtime Detectors ──────────────────────────────────────────────────

type RbenvDetector struct{}

func init() { RegisterRuntime(&RbenvDetector{}) }

func (d *RbenvDetector) Name() RuntimeType { return RuntimeRbenv }
func (d *RbenvDetector) IsAvailable() bool { return HasCommand("rbenv") }
func (d *RbenvDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	rbenvDir := filepath.Join(home, ".rbenv", "versions")
	if !strings.HasPrefix(execPath, rbenvDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, rbenvDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		return RuntimeRbenv, "Ruby " + parts[0], "rbenv", 95
	}
	return RuntimeNone, "", "", 0
}

type RVMDetector struct{}

func init() { RegisterRuntime(&RVMDetector{}) }

func (d *RVMDetector) Name() RuntimeType { return RuntimeRVM }
func (d *RVMDetector) IsAvailable() bool { return HasCommand("rvm") || os.Getenv("rvm_path") != "" }
func (d *RVMDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	rvmDir := filepath.Join(home, ".rvm", "rubies")
	if !strings.HasPrefix(execPath, rvmDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, rvmDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		rubyVer := strings.TrimPrefix(parts[0], "ruby-")
		return RuntimeRVM, "Ruby " + rubyVer, "rvm", 95
	}
	return RuntimeNone, "", "", 0
}

// ─── Java Runtime Detectors ──────────────────────────────────────────────────

type SdkmanDetector struct{}

func init() { RegisterRuntime(&SdkmanDetector{}) }

func (d *SdkmanDetector) Name() RuntimeType { return RuntimeSdkman }
func (d *SdkmanDetector) IsAvailable() bool {
	home := HomeDir()
	sdkmanDir := filepath.Join(home, ".sdkman")
	_, err := os.Stat(sdkmanDir)
	return err == nil
}

func (d *SdkmanDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	sdkmanDir := filepath.Join(home, ".sdkman", "candidates")
	if !strings.HasPrefix(execPath, sdkmanDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, sdkmanDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) >= 2 {
		tool := parts[0]
		version := parts[1]
		if version == "current" && len(parts) >= 3 {
			version = parts[2]
		}
		name := Capitalize(tool)
		return RuntimeSdkman, name + " " + version, "SDKMAN!", 95
	}
	return RuntimeNone, "", "", 0
}

type JenvDetector struct{}

func init() { RegisterRuntime(&JenvDetector{}) }

func (d *JenvDetector) Name() RuntimeType { return RuntimeJenv }
func (d *JenvDetector) IsAvailable() bool { return HasCommand("jenv") }
func (d *JenvDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	jenvDir := filepath.Join(home, ".jenv", "versions")
	if strings.HasPrefix(execPath, jenvDir) {
		rest := strings.TrimPrefix(execPath, jenvDir+"/")
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) > 0 {
			return RuntimeJenv, "Java " + parts[0], "jenv", 85
		}
	}
	return RuntimeNone, "", "", 0
}

// ─── Go Runtime Detectors ────────────────────────────────────────────────────

type GvmDetector struct{}

func init() { RegisterRuntime(&GvmDetector{}) }

func (d *GvmDetector) Name() RuntimeType { return RuntimeGvm }
func (d *GvmDetector) IsAvailable() bool {
	home := HomeDir()
	gvmDir := filepath.Join(home, ".gvm")
	_, err := os.Stat(gvmDir)
	return err == nil
}

func (d *GvmDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	gvmDir := filepath.Join(home, ".gvm", "gos")
	if !strings.HasPrefix(execPath, gvmDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, gvmDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		return RuntimeGvm, "Go " + parts[0], "gvm", 90
	}
	return RuntimeNone, "", "", 0
}

type GoenvDetector struct{}

func init() { RegisterRuntime(&GoenvDetector{}) }

func (d *GoenvDetector) Name() RuntimeType { return RuntimeGoenv }
func (d *GoenvDetector) IsAvailable() bool { return HasCommand("goenv") }
func (d *GoenvDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	goenvDir := filepath.Join(home, ".goenv", "versions")
	if !strings.HasPrefix(execPath, goenvDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, goenvDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		return RuntimeGoenv, "Go " + parts[0], "goenv", 90
	}
	return RuntimeNone, "", "", 0
}

// ─── Rust Runtime Detector ──────────────────────────────────────────────────

type RustupDetector struct{}

func init() { RegisterRuntime(&RustupDetector{}) }

func (d *RustupDetector) Name() RuntimeType { return RuntimeRustup }
func (d *RustupDetector) IsAvailable() bool { return HasCommand("rustup") }
func (d *RustupDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	rustupDir := filepath.Join(home, ".rustup", "toolchains")
	if !strings.HasPrefix(execPath, rustupDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, rustupDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		toolchain := parts[0]
		version := toolchain
		if idx := strings.Index(toolchain, "-"); idx > 0 {
			version = toolchain[:idx]
		}
		return RuntimeRustup, "Rust " + version, "rustup", 95
	}
	return RuntimeNone, "", "", 0
}

// ─── PHP Runtime Detector ───────────────────────────────────────────────────

type PhpenvDetector struct{}

func init() { RegisterRuntime(&PhpenvDetector{}) }

func (d *PhpenvDetector) Name() RuntimeType { return RuntimePhpenv }
func (d *PhpenvDetector) IsAvailable() bool { return HasCommand("phpenv") }
func (d *PhpenvDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	phpenvDir := filepath.Join(home, ".phpenv", "versions")
	if !strings.HasPrefix(execPath, phpenvDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, phpenvDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		return RuntimePhpenv, "PHP " + parts[0], "phpenv", 90
	}
	return RuntimeNone, "", "", 0
}

// ─── Multi-language Version Managers ─────────────────────────────────────────

type AsdfDetector struct{}

func init() { RegisterRuntime(&AsdfDetector{}) }

func (d *AsdfDetector) Name() RuntimeType { return RuntimeAsdf }
func (d *AsdfDetector) IsAvailable() bool {
	home := HomeDir()
	asdfDir := filepath.Join(home, ".asdf")
	_, err := os.Stat(asdfDir)
	return err == nil
}

func (d *AsdfDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	asdfInstalls := filepath.Join(home, ".asdf", "installs")
	if !strings.HasPrefix(execPath, asdfInstalls) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, asdfInstalls+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) >= 2 {
		plugin := parts[0]
		version := parts[1]
		name := Capitalize(plugin)
		return RuntimeAsdf, name + " " + version, "asdf", 90
	}
	return RuntimeNone, "", "", 0
}

type MiseDetector struct{}

func init() { RegisterRuntime(&MiseDetector{}) }

func (d *MiseDetector) Name() RuntimeType { return RuntimeMise }
func (d *MiseDetector) IsAvailable() bool { return HasCommand("mise") }
func (d *MiseDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	miseInstalls := filepath.Join(home, ".local", "share", "mise", "installs")
	if !strings.HasPrefix(execPath, miseInstalls) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, miseInstalls+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) >= 2 {
		plugin := parts[0]
		version := parts[1]
		name := Capitalize(plugin)
		return RuntimeMise, name + " " + version, "mise", 90
	}
	return RuntimeNone, "", "", 0
}

type KerlDetector struct{}

func init() { RegisterRuntime(&KerlDetector{}) }

func (d *KerlDetector) Name() RuntimeType { return RuntimeKerl }
func (d *KerlDetector) IsAvailable() bool {
	home := HomeDir()
	kerlDir := filepath.Join(home, ".kerl", "installs")
	_, err := os.Stat(kerlDir)
	return err == nil
}

func (d *KerlDetector) Detect(execPath string) (RuntimeType, string, string, int) {
	home := HomeDir()
	kerlDir := filepath.Join(home, ".kerl", "installs")
	if !strings.HasPrefix(execPath, kerlDir) {
		return RuntimeNone, "", "", 0
	}
	rest := strings.TrimPrefix(execPath, kerlDir+"/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) > 0 {
		return RuntimeKerl, "Erlang " + parts[0], "kerl", 90
	}
	return RuntimeNone, "", "", 0
}
