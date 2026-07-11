package detector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager string

const (
	ManagerPacman    Manager = "pacman"
	ManagerParu      Manager = "paru"
	ManagerYay       Manager = "yay"
	ManagerBun       Manager = "bun"
	ManagerNpm       Manager = "npm"
	ManagerPnpm      Manager = "pnpm"
	ManagerYarn      Manager = "yarn"
	ManagerPip       Manager = "pip"
	ManagerPipx      Manager = "pipx"
	ManagerUv        Manager = "uv"
	ManagerCargo     Manager = "cargo"
	ManagerGo        Manager = "go"
	ManagerGem       Manager = "gem"
	ManagerComposer  Manager = "composer"
	ManagerFlatpak   Manager = "flatpak"
	ManagerSnap      Manager = "snap"
	ManagerAppImage  Manager = "appimage"
	ManagerBrew      Manager = "homebrew"
	ManagerNix       Manager = "nix"
	ManagerDocker    Manager = "docker"
	ManagerPodman    Manager = "podman"
	ManagerDistrobox Manager = "distrobox"
	ManagerToolbox   Manager = "toolbox"
	ManagerManual    Manager = "manual"
)

type RuntimeType string

const (
	RuntimeNone    RuntimeType = ""
	RuntimeFnm     RuntimeType = "fnm"
	RuntimeNvm     RuntimeType = "nvm"
	RuntimeVolta   RuntimeType = "volta"
	RuntimeAsdf    RuntimeType = "asdf"
	RuntimeMise    RuntimeType = "mise"
	RuntimePyenv   RuntimeType = "pyenv"
	RuntimeRye     RuntimeType = "rye"
	RuntimeConda   RuntimeType = "conda"
	RuntimeMamba   RuntimeType = "mamba"
	RuntimeMicromamba RuntimeType = "micromamba"
	RuntimePixi    RuntimeType = "pixi"
	RuntimePoetry  RuntimeType = "poetry"
	RuntimeRbenv   RuntimeType = "rbenv"
	RuntimeRVM     RuntimeType = "rvm"
	RuntimeSdkman  RuntimeType = "sdkman"
	RuntimeJenv    RuntimeType = "jenv"
	RuntimeGvm     RuntimeType = "gvm"
	RuntimeGoenv   RuntimeType = "goenv"
	RuntimeRustup  RuntimeType = "rustup"
	RuntimePhpenv  RuntimeType = "phpenv"
	RuntimeKerl    RuntimeType = "kerl"
	RuntimeBun     RuntimeType = "bun" // bun can be both runtime and package manager
)

type Result struct {
	Found           bool        `json:"found" yaml:"found"`
	Executable      string      `json:"executable" yaml:"executable"`
	ExecutablePath  string      `json:"executable_path" yaml:"executable_path"`
	Manager         Manager     `json:"manager" yaml:"manager"`
	Package         string      `json:"package" yaml:"package"`
	Version         string      `json:"version,omitempty" yaml:"version,omitempty"`
	Repository      string      `json:"repository,omitempty" yaml:"repository,omitempty"`
	IsAUR           bool        `json:"is_aur,omitempty" yaml:"is_aur,omitempty"`
	Global          *bool       `json:"global,omitempty" yaml:"global,omitempty"`
	PythonVersion   string      `json:"python_version,omitempty" yaml:"python_version,omitempty"`
	VirtualEnv      string      `json:"virtual_env,omitempty" yaml:"virtual_env,omitempty"`
	Remote          string      `json:"remote,omitempty" yaml:"remote,omitempty"`
	InstallCmd      string      `json:"install_command" yaml:"install_command"`
	UpgradeCmd      string      `json:"upgrade_command" yaml:"upgrade_command"`
	RemoveCmd       string      `json:"remove_command" yaml:"remove_command"`
	Explain         string      `json:"explain,omitempty" yaml:"explain,omitempty"`
	SimilarPackages []string    `json:"similar_packages,omitempty" yaml:"similar_packages,omitempty"`
	Runtime         RuntimeType `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	RuntimeVersion  string      `json:"runtime_version,omitempty" yaml:"runtime_version,omitempty"`
	RuntimeManager  string      `json:"runtime_manager,omitempty" yaml:"runtime_manager,omitempty"`
	Confidence      int         `json:"confidence" yaml:"confidence"`
}

func (r *Result) ManagerName() string {
	return string(r.Manager)
}

type Detector interface {
	Name() Manager
	Detect(execPath string) *Result
	IsAvailable() bool
}

type RuntimeDetector interface {
	Name() RuntimeType
	Detect(execPath string) (runtime RuntimeType, version string, manager string, confidence int)
	IsAvailable() bool
}

func ResolveExecPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("executable %q not found in PATH", name)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	realPath, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return realPath, nil
	}
	return abs, nil
}

func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func RunCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func RunCmdStdout(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func BoolPtr(b bool) *bool {
	return &b
}

func HomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func PathHasPrefix(path, prefix string) bool {
	return strings.HasPrefix(path, prefix)
}

func ExtractVersionFromPath(execPath string, markers ...string) string {
	for _, marker := range markers {
		idx := strings.Index(execPath, marker)
		if idx == -1 {
			continue
		}
		rest := execPath[idx+len(marker):]
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) > 0 {
			v := parts[0]
			v = strings.TrimPrefix(v, "v")
			return v
		}
	}
	return ""
}
