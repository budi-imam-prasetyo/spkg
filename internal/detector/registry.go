package detector

import (
	"runtime"
	"sort"
	"strings"
	"sync"
)

var (
	registry   []Detector
	runtimeReg []RuntimeDetector
	mu         sync.RWMutex
	runtimeMu  sync.RWMutex

	// prioritySlice is a precomputed, stable ordering of managers by priority,
	// built once at init. DetectFirst uses it for O(1) lookups instead of
	// allocating a map on every invocation.
	prioritySlice []priorityEntry
)

type priorityEntry struct {
	m Manager
	p int
}

func init() {
	// Precompute a stable priority ordering once; lookups become a linear scan
	// over a tiny slice instead of re-evaluating the priorityOrder function and
	// allocating a map on every DetectFirst call.
	for _, m := range allManagers {
		prioritySlice = append(prioritySlice, priorityEntry{m, priorityOrder(m)})
	}
	sort.Slice(prioritySlice, func(i, j int) bool {
		return prioritySlice[i].p < prioritySlice[j].p
	})
}

// allManagers is the exhaustive set of Manager constants, used to build the
// precomputed priority slice.
var allManagers = []Manager{
	ManagerPacman, ManagerParu, ManagerYay, ManagerBun, ManagerNpm,
	ManagerPnpm, ManagerYarn, ManagerPip, ManagerPipx, ManagerUv,
	ManagerCargo, ManagerGo, ManagerGem, ManagerComposer, ManagerFlatpak,
	ManagerSnap, ManagerAppImage, ManagerBrew, ManagerNix, ManagerDocker,
	ManagerPodman, ManagerDistrobox, ManagerToolbox, ManagerManual,
	ManagerFnm, ManagerNvm, ManagerVolta, ManagerRustup,
}

// priorityOf returns the priority for a manager, using the precomputed slice.
func priorityOf(m Manager) int {
	for _, e := range prioritySlice {
		if e.m == m {
			return e.p
		}
	}
	return 50
}

func Register(d Detector) {
	mu.Lock()
	defer mu.Unlock()
	registry = append(registry, d)
}

func RegisterRuntime(r RuntimeDetector) {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()
	runtimeReg = append(runtimeReg, r)
}

func GetAll() []Detector {
	mu.RLock()
	defer mu.RUnlock()
	cp := make([]Detector, len(registry))
	copy(cp, registry)
	return cp
}

func GetAllRuntimes() []RuntimeDetector {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()
	cp := make([]RuntimeDetector, len(runtimeReg))
	copy(cp, runtimeReg)
	return cp
}

func DetectAll(execPath ResolvedPath) []*Result {
	detectors := GetAll()
	var results []*Result
	for _, d := range detectors {
		if !d.IsAvailable() {
			continue
		}
		for _, p := range execPath.Both() {
			r := d.Detect(p)
			if r != nil && r.Found {
				results = append(results, r)
				break
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return priorityOf(results[i].Manager) < priorityOf(results[j].Manager)
	})
	return results
}

// DetectBatch runs full detection (package manager + runtime) for many paths.
// It uses a single snapshot of the registry/runtime detectors and reuses the
// cache between entries, which is materially faster than calling
// DetectWithRuntime in a loop (e.g. `spkg all`).
func DetectBatch(paths []ResolvedPath, names []string) []*Result {
	results := make([]*Result, len(paths))
	if len(paths) == 0 {
		return results
	}
	workers := min(runtime.GOMAXPROCS(0), 8)
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				name := ""
				if i < len(names) {
					name = names[i]
				}
				results[i] = DetectWithRuntime(paths[i], name)
			}
		}()
	}
	for i := range paths {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

func DetectFirst(execPath ResolvedPath) *Result {
	detectors := GetAll()
	var best *Result
	bestPriority := 1 << 30
	// Collect every detector that claims this path, then return the one with the
	// highest priority (lowest priorityOrder). Registration order is lexical by
	// filename and must NOT decide the winner, otherwise e.g. npm (npm.go) can
	// claim a node binary that is actually managed by fnm/nvm/volta.
	for _, d := range detectors {
		if !d.IsAvailable() {
			continue
		}
		for _, p := range execPath.Both() {
			r := d.Detect(p)
			if r == nil || !r.Found {
				continue
			}
			// Prefer a detector that has actual path-based evidence (real
			// ownership) over one that only guessed from a name and has no
			// package/version detail.
			if best == nil || priorityOf(r.Manager) < bestPriority {
				best = r
				bestPriority = priorityOf(r.Manager)
			}
			break
		}
	}
	return best
}

func DetectRuntime(execPath ResolvedPath) (RuntimeType, string, string, int) {
	detectors := GetAllRuntimes()
	for _, d := range detectors {
		if !d.IsAvailable() {
			continue
		}
		rt, version, manager, confidence := d.Detect(execPath.Real)
		if rt != RuntimeNone && confidence > 0 {
			return rt, version, manager, confidence
		}
	}
	return RuntimeNone, "", "", 0
}

func DetectWithRuntime(execPath ResolvedPath, name string) *Result {
	r := DetectFirst(execPath)
	// pmDetected: a package-manager Detect() (npm/pnpm/yarn/bun/...) actually
	// claimed this path. Only detectors in the package-manager registry can
	// produce a non-nil result here (runtime/version managers live in a
	// separate registry), so this cleanly distinguishes "a package manager
	// owns a real package at this path" from the manual fallback.
	pmDetected := r != nil
	rt, version, manager, rtConf := DetectRuntime(execPath)
	runtimeOwns := rt != RuntimeNone

	if r == nil && !runtimeOwns {
		r = NewManualResult(execPath, name)
		score, manualReasons := ManualConfidence()
		r.Confidence = score
		r.Reasons = ReasonString(manualReasons)
		return r
	}
	if r == nil {
		r = NewManualResult(execPath, name)
	}

	// Structural band from the winning detector's cheap Score() signal.
	band, reasons := scoreFor(r.Manager, execPath)

	if runtimeOwns {
		// If a version/manager (fnm, nvm, volta, pyenv, rustup, ...) owns the
		// path, it — not the package manager — is the authority. A package
		// manager like npm/pip can still be the one that *installed* the
		// tooling, but reporting npm as the owner of an fnm-managed node is a
		// false positive, so we prefer the runtime as the primary source.
		r.Runtime = rt
		r.RuntimeVersion = version
		r.RuntimeManager = manager
		// EXCEPTION: a runtime manager's dir also contains global packages
		// installed by a package manager (e.g. `npm install -g 9remote` lands
		// under an fnm-managed node's lib/node_modules). The runtime is only
		// the authority for its OWN bundled tooling (node/npm/npx); a genuine
		// third-party global package stays owned by the package manager that
		// installed it, so `spkg 9remote` shows `npm uninstall -g 9remote`
		// instead of the nonsensical `fnm uninstall 24.20.0` (which would
		// remove the entire Node.js runtime). Keep the runtime as metadata.
		pmOwnsRealPackage := pmDetected && !isRuntimeBundledPackage(r.Package)
		if !pmOwnsRealPackage {
			applyRuntimeOwner(r, rt, runtimeVersionNumber(version))
		}
		// Re-evaluate the structural band for the (possibly new) manager.
		band, reasons = scoreFor(r.Manager, execPath)
		// A runtime owning the path is itself strong structural evidence; the
		// "runtime owns path" reason is added exactly once by CombineConfidence
		// via the RuntimeOwnsPath flag below (no manual append here).
		if band < ConfidenceMedium {
			band = ConfidenceMedium
		}
		_ = rtConf
	}

	// "Confirmed" means Detect() ran and produced a positive result (the manager's
	// own tool confirmed ownership). For most detectors that returned a *Result
	// here, that's true; we track it via a flag set by DetectFirst.
	return decorateConfidence(r, execPath, r.Manager, band, reasons, runtimeOwns)
}

// scoreFor returns the structural band + reasons for a manager on a given path,
// using the detector's side-effect-free Score().
func scoreFor(m Manager, execPath ResolvedPath) (Confidence, []Reason) {
	for _, d := range GetAll() {
		if d.Name() != m {
			continue
		}
		if !d.IsAvailable() {
			return ConfidenceUnknown, nil
		}
		return d.Score(execPath.Real)
	}
	return ConfidenceUnknown, nil
}

// decorateConfidence computes the final Confidence + Reasons via the engine and
// attaches them to the result. `confirmed` is true when the manager's own
// detection command succeeded (command-level proof).
func decorateConfidence(r *Result, execPath ResolvedPath, m Manager, band Confidence, reasons []Reason, runtimeOwns bool) *Result {
	// Exclusion count: how many *other* managers' domains do NOT match.
	excluded := 0
	for _, d := range GetAll() {
		if d.Name() == m {
			continue
		}
		if !d.IsAvailable() {
			continue
		}
		b, _ := d.Score(execPath.Real)
		if b == ConfidenceUnknown {
			excluded++
		}
	}

	// Command confirmation requires explicit evidence from the detector. A
	// positive Result alone may only mean a path/name heuristic matched.
	confirmed := r != nil && r.Found && r.Manager == m && len(r.ConfirmReasons) > 0

	score, finalReasons := CombineConfidence(combineInput{
		Manager:           m,
		StructuralBand:    band,
		StructuralReasons: reasons,
		Confirmed:         confirmed,
		ConfirmReasons:    r.ConfirmReasons,
		RuntimeOwnsPath:   runtimeOwns,
		ExcludedBy:        excluded,
	})

	r.Confidence = score
	r.Reasons = ReasonString(finalReasons)
	return r
}

func priorityOrder(m Manager) int {
	priorities := map[Manager]int{
		ManagerPacman:    1,
		ManagerParu:      2,
		ManagerYay:       3,
		ManagerFlatpak:   4,
		ManagerSnap:      5,
		ManagerBrew:      6,
		ManagerNix:       7,
		ManagerFnm:       8,
		ManagerNvm:       8,
		ManagerVolta:     8,
		ManagerRustup:    8,
		ManagerBun:       10,
		ManagerNpm:       11,
		ManagerPnpm:      12,
		ManagerYarn:      13,
		ManagerPip:       20,
		ManagerPipx:      21,
		ManagerUv:        22,
		ManagerCargo:     30,
		ManagerGo:        31,
		ManagerGem:       32,
		ManagerComposer:  33,
		ManagerDocker:    40,
		ManagerPodman:    41,
		ManagerDistrobox: 42,
		ManagerToolbox:   43,
		ManagerAppImage:  50,
		ManagerManual:    99,
	}
	if p, ok := priorities[m]; ok {
		return p
	}
	return 50
}

func runtimeVersionNumber(version string) string {
	parts := strings.Fields(version)
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return version
}

func applyRuntimeOwner(r *Result, rt RuntimeType, version string) {
	r.Manager = managerFromRuntime(rt)
	switch rt {
	case RuntimeFnm:
		r.Package = "node"
		r.InstallCmd = "fnm install " + version
		r.UpgradeCmd = "fnm install --latest"
		r.RemoveCmd = "fnm uninstall " + version
	case RuntimeNvm:
		r.Package = "node"
		r.InstallCmd = "nvm install " + version
		r.UpgradeCmd = "nvm install node"
		r.RemoveCmd = "nvm uninstall " + version
	case RuntimeVolta:
		r.Package = "node"
		r.InstallCmd = "volta install node@" + version
		r.UpgradeCmd = "volta install node@latest"
		r.RemoveCmd = "volta uninstall node"
	case RuntimeRustup:
		r.Package = "rust"
		r.InstallCmd = "rustup toolchain install " + version
		r.UpgradeCmd = "rustup update"
		r.RemoveCmd = "rustup toolchain uninstall " + version
	default:
		r.Package = string(rt)
		r.InstallCmd = "# Managed by " + r.RuntimeManager
		r.UpgradeCmd = "# Upgrade with " + r.RuntimeManager
		r.RemoveCmd = "# Remove with " + r.RuntimeManager
	}
}

// isRuntimeBundledPackage reports whether the package name is part of a runtime
// manager's own bundled toolchain (the interpreter/runtime itself and the
// package manager shipped with it), as opposed to a third-party global package
// a user installed into the runtime's dir. For bundled tooling the runtime
// manager is the correct owner (e.g. `fnm uninstall <ver>` for node/npm);
// for a real package it is not.
func isRuntimeBundledPackage(pkg string) bool {
	switch pkg {
	case "node", "npm", "npx", "corepack",
		"rust", "rustc", "cargo",
		"python", "python3", "pip", "pip3",
		"ruby", "gem",
		"go", "gofmt",
		"php",
		"bun":
		return true
	default:
		return false
	}
}

// managerFromRuntime maps a detected runtime/version manager to the Manager
// constant we report as the owner of the executable. When a runtime manager owns
// the path it placed, we report THAT as the owner — not a coincidentally-matching
// package manager (e.g. an fnm-managed node must not be reported as "npm").
// Where no dedicated owner constant exists, we fall back to the package manager
// that installs the toolchain (pip for pyenv-envs, cargo for rustup, ...).
func managerFromRuntime(rt RuntimeType) Manager {
	switch rt {
	case RuntimeFnm:
		return ManagerFnm
	case RuntimeNvm:
		return ManagerNvm
	case RuntimeVolta:
		return ManagerVolta
	case RuntimeRustup:
		return ManagerRustup
	case RuntimePyenv, RuntimeRye, RuntimePoetry, RuntimeConda, RuntimeMamba, RuntimeMicromamba, RuntimePixi:
		return ManagerPip
	case RuntimeGoenv, RuntimeGvm:
		return ManagerGo
	case RuntimeRbenv, RuntimeRVM:
		return ManagerGem
	case RuntimeSdkman, RuntimeJenv, RuntimeKerl, RuntimePhpenv, RuntimeAsdf, RuntimeMise, RuntimeBun:
		return ManagerManual
	default:
		return ManagerManual
	}
}
