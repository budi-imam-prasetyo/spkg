// detector.go provides core interfaces for package manager detection.
package detector

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// System interface for mockable system calls.
type System interface {
	LookPath(file string) (string, error)
	CmdContext(ctx context.Context, name string, args ...string) *exec.Cmd
}

type realSystem struct{}

func (realSystem) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (realSystem) CmdContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// sysMu guards the pluggable System implementation. Detection runs concurrently
// (DetectBatch), so swapping the implementation must not race with readers.
var (
	sysMu sync.RWMutex
	sys   System = realSystem{}
)

// SetSystem replaces the System implementation used for external command
// lookups and execution. Passing nil restores the real OS-backed one. Intended
// for tests that need deterministic behaviour without host tooling installed.
func SetSystem(s System) {
	sysMu.Lock()
	defer sysMu.Unlock()
	if s == nil {
		s = realSystem{}
	}
	sys = s
	clearMemo()
}

func currentSystem() System {
	sysMu.RLock()
	defer sysMu.RUnlock()
	return sys
}

var diagnostics = struct {
	sync.RWMutex
	enabled bool
	writer  io.Writer
}{writer: os.Stderr}

func SetDiagnostics(enabled bool, writer io.Writer) {
	diagnostics.Lock()
	defer diagnostics.Unlock()
	diagnostics.enabled = enabled
	if writer == nil {
		writer = os.Stderr
	}
	diagnostics.writer = writer
}

func reportCommandError(name string, args []string, err error) {
	diagnostics.RLock()
	defer diagnostics.RUnlock()
	if diagnostics.enabled {
		_, _ = fmt.Fprintf(diagnostics.writer, "detector command failed: %s %s: %v\n", name, strings.Join(args, " "), err)
	}
}

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

	// Runtime/version managers that can be the *owner* of a binary they placed
	// in PATH (e.g. fnm owns the node it symlinks into a multishell bin dir).
	ManagerFnm    Manager = "fnm"
	ManagerNvm    Manager = "nvm"
	ManagerVolta  Manager = "volta"
	ManagerRustup Manager = "rustup"
)

type RuntimeType string

const (
	RuntimeNone       RuntimeType = ""
	RuntimeFnm        RuntimeType = "fnm"
	RuntimeNvm        RuntimeType = "nvm"
	RuntimeVolta      RuntimeType = "volta"
	RuntimeAsdf       RuntimeType = "asdf"
	RuntimeMise       RuntimeType = "mise"
	RuntimePyenv      RuntimeType = "pyenv"
	RuntimeRye        RuntimeType = "rye"
	RuntimeConda      RuntimeType = "conda"
	RuntimeMamba      RuntimeType = "mamba"
	RuntimeMicromamba RuntimeType = "micromamba"
	RuntimePixi       RuntimeType = "pixi"
	RuntimePoetry     RuntimeType = "poetry"
	RuntimeRbenv      RuntimeType = "rbenv"
	RuntimeRVM        RuntimeType = "rvm"
	RuntimeSdkman     RuntimeType = "sdkman"
	RuntimeJenv       RuntimeType = "jenv"
	RuntimeGvm        RuntimeType = "gvm"
	RuntimeGoenv      RuntimeType = "goenv"
	RuntimeRustup     RuntimeType = "rustup"
	RuntimePhpenv     RuntimeType = "phpenv"
	RuntimeKerl       RuntimeType = "kerl"
	RuntimeBun        RuntimeType = "bun" // bun can be both runtime and package manager
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
	// AurHelper indicates which AUR helper installed this package if detected
	// (e.g., "paru", "yay"). Empty if not detected or not an AUR package.
	AurHelper string `json:"aur_helper,omitempty" yaml:"aur_helper,omitempty"`
	// Confidence is our certainty (0-100, see ConfidenceLevel) that the reported
	// manager genuinely owns the executable. It is computed structurally by
	// CombineConfidence (see confidence.go), combining the evidence each detector
	// returned with the strength of the ownership mechanism that matched.
	Confidence int `json:"confidence,omitempty" yaml:"confidence,omitempty"`
	// Reasons records the individual evidence signals that contributed to
	// Confidence, so users and tests can see *why* a score was assigned.
	Reasons []string `json:"reasons,omitempty" yaml:"reasons,omitempty"`
	// ConfirmReasons are command-level proof reasons attached by a detector during
	// Detect (e.g. "pacman -Qo: owned by git"). The engine folds these into the
	// final Confidence when confirmed.
	ConfirmReasons []Reason `json:"-" yaml:"-"`
}

// Confidence is a 0-100 certainty score. The band constants below are the
// canonical thresholds; individual detectors emit a band + reasons, and the
// engine combines them into a final score on this scale.
type Confidence int

const (
	// ConfidenceUnknown: no real evidence — manual fallback only.
	ConfidenceUnknown Confidence = 0
	// ConfidenceLow: weak, indirect evidence (name match, ambiguous dir).
	ConfidenceLow Confidence = 40
	// ConfidenceMedium: solid path/ownership evidence but no command confirmation.
	ConfidenceMedium Confidence = 70
	// ConfidenceHigh: strong evidence (package-manager DB / runtime owns the path).
	ConfidenceHigh Confidence = 90
	// ConfidenceCertain: cryptographic/DB-backed ownership proof (pacman -Qo, etc).
	ConfidenceCertain Confidence = 100
)

// Reason is a single evidence signal that contributed to a confidence score.
type Reason struct {
	// Signal identifies the kind of evidence, e.g. "path_match",
	// "manager_db", "runtime_owns_path", "name_match".
	Signal string
	// Detail is a human-readable explanation.
	Detail string
	// Weight is the points this signal contributed (before clamping).
	Weight int
}

// Band returns the ConfidenceLevel band for a score.
func (c Confidence) Band() ConfidenceLevel { return BandFor(int(c)) }

func (c Confidence) Label() string { return LabelFor(int(c)) }

// ConfidenceLevel groups scores into a coarse category for output/UX.
type ConfidenceLevel string

const (
	LevelUnknown ConfidenceLevel = "unknown" // 0
	LevelLow     ConfidenceLevel = "low"     // 1-49
	LevelMedium  ConfidenceLevel = "medium"  // 50-89
	LevelHigh    ConfidenceLevel = "high"    // 90-99
	LevelCertain ConfidenceLevel = "certain" // 100
)

func BandFor(score int) ConfidenceLevel {
	switch {
	case score <= 0:
		return LevelUnknown
	case score < 50:
		return LevelLow
	case score < 90:
		return LevelMedium
	case score < 100:
		return LevelHigh
	default:
		return LevelCertain
	}
}

func LabelFor(score int) string { return string(BandFor(score)) }

func (r *Result) ManagerName() string {
	return string(r.Manager)
}

type Detector interface {
	Name() Manager
	Detect(execPath string) *Result
	IsAvailable() bool
	// Score reports how confident this detector is that `execPath` is owned by
	// its manager, WITHOUT needing a successful full Detect (which may spawn
	// subprocesses). It is the cheap, structural signal used by the confidence
	// engine to rank and combine candidates.
	Score(execPath string) (band Confidence, reasons []Reason)
}

type RuntimeDetector interface {
	Name() RuntimeType
	Detect(execPath string) (runtime RuntimeType, version string, manager string, confidence int)
	IsAvailable() bool
}

type ResolvedPath struct {
	Original string
	Real     string
}

func (rp ResolvedPath) Both() []string {
	if rp.Original == rp.Real {
		return []string{rp.Real}
	}
	return []string{rp.Original, rp.Real}
}

func ResolveExecPath(name string) (ResolvedPath, error) {
	path, err := sys.LookPath(name)
	if err != nil {
		return ResolvedPath{}, &NotFoundError{name: name}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	realPath, err := filepath.EvalSymlinks(abs)
	if err != nil {
		realPath = abs
	}
	// Normalize paths for consistent comparison
	abs = filepath.Clean(abs)
	realPath = filepath.Clean(realPath)
	return ResolvedPath{Original: abs, Real: realPath}, nil
}

// NotFoundError is returned by ResolveExecPath when the named executable is not
// in PATH. It lets the CLI exit non-zero on a miss so scripts can distinguish
// "found" from "not found".
type NotFoundError struct {
	name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("executable %q not found in PATH", e.name)
}

// IsNotFound reports whether err is a NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// cmdTimeout bounds how long a single detector's external command may run.
// Without this, a hung package-manager daemon (flatpak, snap, ...) would hang
// the whole CLI with no way to abort.
const cmdTimeout = 15 * time.Second

func HasCommand(name string) bool {
	return memoizedHasCommand(name)
}

// RunCmd runs name with args, returning combined stdout/stderr trimmed of
// surrounding whitespace. It is bounded by cmdTimeout; if the command is killed
// the returned error carries the context deadline. The underlying error is
// returned (not swallowed) so callers can surface failures in verbose mode.
func RunCmd(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cmd := currentSystem().CmdContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		reportCommandError(name, args, err)
	}
	return strings.TrimSpace(string(out)), err
}

func RunCmdStdout(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	cmd := currentSystem().CmdContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		reportCommandError(name, args, err)
	}
	return strings.TrimSpace(string(out)), err
}

func BoolPtr(b bool) *bool {
	return &b
}

func HomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
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
