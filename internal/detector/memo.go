package detector

import (
	"strings"
	"sync"
	"time"
)

// memo.go — per-process memoization for expensive, idempotent subprocess calls.
//
// WHY: `spkg all` resolves ~35 binaries. Without memoization every binary
// re-runs the same global queries (`npm prefix -g`, `npm list -g`,
// `cargo install --list`, `pipx list`, `uv tool list`, `brew --prefix`,
// `go env GOPATH`, `flatpak list`, `snap list`, ...). That turns a cold scan
// into 100+ subprocess spawns (~40s). The outputs are per-user/per-machine
// globals that don't change mid-run, so caching them for the life of the
// process (with a short TTL) is safe and cuts cold scans dramatically.
//
// THREAD SAFETY: DetectBatch runs detection concurrently (up to 8 workers).
// All maps below are sync.Map. SetSystem() clears both caches so mock-based
// tests stay deterministic.

var (
	availMemo sync.Map // string -> availEntry
	cmdMemo   sync.Map // string -> cmdEntry
)

type availEntry struct {
	ok      bool
	expires time.Time
}

type cmdEntry struct {
	out     string
	errStr  string
	failed  bool
	expires time.Time
}

const (
	availTTL = 5 * time.Minute
	cmdTTL   = 2 * time.Minute
)

// clearMemo drops all memoized state. Called from SetSystem so tests that swap
// the System implementation don't see stale LookPath / command output.
func clearMemo() {
	availMemo = sync.Map{}
	cmdMemo = sync.Map{}
}

// memoizedHasCommand is HasCommand with a short-lived per-process cache.
// LookPath walks $PATH on every call; detectors call it ~25x per binary.
func memoizedHasCommand(name string) bool {
	if e, ok := availMemo.Load(name); ok {
		entry := e.(availEntry)
		if time.Now().Before(entry.expires) {
			return entry.ok
		}
	}
	_, err := currentSystem().LookPath(name)
	ok := err == nil
	availMemo.Store(name, availEntry{ok: ok, expires: time.Now().Add(availTTL)})
	return ok
}

// memoizedRunCmd caches idempotent command output for cmdTTL. Keyed on the full
// argv so different args don't collide. Only use for read-only queries whose
// output is stable within a single CLI run (prefix/list/env). Per-binary
// queries (pacman -Qo <path>, pacman -Qi <pkg>) must NOT use this.
func memoizedRunCmd(name string, args ...string) (string, error) {
	key := name + "\x00" + strings.Join(args, "\x00")
	if e, ok := cmdMemo.Load(key); ok {
		entry := e.(cmdEntry)
		if time.Now().Before(entry.expires) {
			if entry.failed {
				return entry.out, errFromString(entry.errStr)
			}
			return entry.out, nil
		}
	}
	out, err := RunCmd(name, args...)
	entry := cmdEntry{out: out, expires: time.Now().Add(cmdTTL)}
	if err != nil {
		entry.failed = true
		entry.errStr = err.Error()
	}
	cmdMemo.Store(key, entry)
	return out, err
}

func errFromString(s string) error {
	if s == "" {
		return errMemoFailed
	}
	return &memoError{s}
}

var errMemoFailed = &memoError{"command failed"}

type memoError struct{ s string }

func (e *memoError) Error() string { return e.s }

// ---- Shared query helpers (all memoized) ----

func npmPrefix() string {
	out, _ := memoizedRunCmd("npm", "prefix", "-g")
	return out
}

func npmGlobalList() string {
	out, _ := memoizedRunCmd("npm", "list", "-g", "--depth=0", "--json")
	return out
}

func pnpmPrefix() string {
	out, _ := memoizedRunCmd("pnpm", "prefix", "-g")
	return out
}

func pnpmGlobalList() string {
	out, _ := memoizedRunCmd("pnpm", "list", "-g", "--depth=0")
	return out
}

func yarnGlobalDir() string {
	out, _ := memoizedRunCmd("yarn", "global", "dir")
	return out
}

func yarnGlobalList() string {
	out, _ := memoizedRunCmd("yarn", "global", "list", "--depth=0")
	return out
}

func cargoInstallList() (string, error) {
	return memoizedRunCmd("cargo", "install", "--list")
}

func pipxListShort() (string, error) {
	return memoizedRunCmd("pipx", "list", "--short")
}

func pipxListFull() (string, error) {
	return memoizedRunCmd("pipx", "list")
}

func uvToolList() (string, error) {
	return memoizedRunCmd("uv", "tool", "list")
}

func brewPrefix() string {
	out, _ := memoizedRunCmd("brew", "--prefix")
	return out
}

func goGopath() string {
	out, _ := memoizedRunCmd("go", "env", "GOPATH")
	return out
}

func flatpakAppList() (string, error) {
	return memoizedRunCmd("flatpak", "list", "--app", "--columns=application,origin,installation")
}

func flatpakInfo(appID string) string {
	out, _ := memoizedRunCmd("flatpak", "info", appID)
	return out
}

func snapList() (string, error) {
	return memoizedRunCmd("snap", "list")
}

func snapInfo(name string) string {
	out, _ := memoizedRunCmd("snap", "info", name)
	return out
}

func gemListDetail() (string, error) {
	return memoizedRunCmd("gem", "list", "-d")
}

func composerGlobalHome() string {
	out, _ := memoizedRunCmd("composer", "global", "config", "home")
	return out
}

func bunGlobalList() (string, error) {
	return memoizedRunCmd("bun", "pm", "ls", "--global")
}

// containsWord reports whether word appears as a whole whitespace/comma-
// separated token in s (case-sensitive). Prevents "black" matching "blackd".
func containsWord(s, word string) bool {
	if word == "" {
		return false
	}
	for _, f := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == ':' || r == '"' || r == '\'' || r == '(' || r == ')'
	}) {
		if f == word {
			return true
		}
	}
	return false
}

// matchExecutablesLine reports whether an "Executables: a, b, c" line from
// `gem list -d` contains exactly base.
func matchExecutablesLine(line, base string) bool {
	idx := strings.Index(line, ":")
	if idx == -1 {
		return false
	}
	rest := line[idx+1:]
	for _, f := range strings.FieldsFunc(rest, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == '\n'
	}) {
		if f == base {
			return true
		}
	}
	return false
}
