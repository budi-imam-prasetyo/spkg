package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ryoukaii/spkg/internal/cache"
	"github.com/ryoukaii/spkg/internal/detector"
	"github.com/ryoukaii/spkg/internal/output"
	"github.com/spf13/cobra"
)

var (
	format      string
	verbose     bool
	noCache     bool
	cacheTTLSec int
)

func NewRootCmd() *cobra.Command {
	cache.Init(10*time.Minute, 30*time.Minute)

	root := &cobra.Command{
		Use:     "spkg [executable...]",
		Short:   "Detect which package manager installed an executable",
		Version: appVersion,
		Long: `spkg detects which package manager installed a given executable.
It supports pacman, bun, npm, pnpm, yarn, pip, pipx, uv,
cargo, go, gem, composer, flatpak, snap, homebrew, nix, docker, podman,
and more.

It also detects runtime/version managers: fnm, nvm, volta, asdf, mise,
pyenv, conda, rbenv, RVM, SDKMAN!, rustup, and others.

You can check multiple executables at once:
  spkg ls cat grep`,
		Args: cobra.ArbitraryArgs,
		// Silence the automatic usage dump: callers (notably "executable not
		// found") want a clean error line, not the full help text. main.go prints
		// the error itself.
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := output.ParseFormat(format); !ok {
				return fmt.Errorf("invalid output format %q (expected table, json, yaml, plain, or pretty)", format)
			}
			if cacheTTLSec > 0 {
				cache.SetTTL(time.Duration(cacheTTLSec) * time.Second)
			}
			return nil
		},
		RunE: runDetect,
	}

	root.PersistentFlags().StringVarP(&format, "output", "o", "table", "Output format: table, json, yaml, plain, pretty")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose detection output")
	root.PersistentFlags().BoolVar(&noCache, "no-cache", false, "Skip cache")
	root.PersistentFlags().IntVar(&cacheTTLSec, "cache-ttl", 600, "Cache TTL in seconds (default: 600)")

	root.AddCommand(newCacheCmd())
	root.AddCommand(newExplainCmd())
	root.AddCommand(newAllCmd())
	root.AddCommand(newVersionCmd())

	return root
}

func runDetect(cmd *cobra.Command, args []string) error {
	detector.SetDiagnostics(verbose, os.Stderr)
	defer detector.SetDiagnostics(false, nil)
	if len(args) == 0 {
		return cmd.Help()
	}

	f := parseFormat(format)

	// Multiple executables
	if len(args) > 1 {
		// results is indexed by argument position so output order always
		// matches input order, regardless of which args hit the cache.
		results := make([]*detector.Result, len(args))
		missIdx := make([]int, 0, len(args))
		missPaths := make([]detector.ResolvedPath, 0, len(args))

		for i, name := range args {
			rp, err := detector.ResolveExecPath(name)
			if err != nil {
				// Skip not found executables silently in batch mode.
				continue
			}

			if !noCache {
				if cached, found := cache.Get(rp.Real); found {
					copy := *cached
					copy.Executable = name
					results[i] = &copy
					continue
				}
			}
			missIdx = append(missIdx, i)
			missPaths = append(missPaths, rp)
		}

		// Batch-detect the cache misses, then place them back in position.
		if len(missIdx) > 0 {
			missNames := make([]string, len(missIdx))
			for n, i := range missIdx {
				missNames[n] = args[i]
			}

			for n, r := range detector.DetectBatch(missPaths, missNames) {
				if !noCache {
					cache.Set(missPaths[n].Real, r)
				}
				// Report the name the user typed, not the symlink-resolved
				// basename (npm -> npm-cli.js, node -> node, etc.).
				r.Executable = missNames[n]
				results[missIdx[n]] = r
			}
		}

		// Drop nil slots (unresolved executables).
		found := make([]*detector.Result, 0, len(results))
		for _, r := range results {
			if r != nil {
				found = append(found, r)
			}
		}
		results = found

		if len(results) == 0 {
			fmt.Println("No executables found.")
			return fmt.Errorf("none of the specified executables were found")
		}

		output.PrintResults(results, f)
		return nil
	}

	// Single executable (original logic)
	name := args[0]

	rp, err := detector.ResolveExecPath(name)
	if err != nil {
		output.PrintNotFound(name, f)
		return err
	}

	// Cache by real path
	if !noCache {
		if cached, found := cache.Get(rp.Real); found {
			if verbose {
				printVerbose(name, rp.Real, true)
			}
			// Copy to avoid mutating cached object
			copy := *cached
			copy.Executable = name
			output.PrintResult(&copy, f, verbose)
			return nil
		}
	}

	if verbose {
		printVerbose(name, rp.Real, false)
	}

	r := detector.DetectWithRuntime(rp, name)
	// Report the name the user typed, not the symlink-resolved basename
	// (npm -> npm-cli.js, node -> node, etc.).
	r.Executable = name

	if !noCache {
		cache.Set(rp.Real, r)
	}

	output.PrintResult(r, f, verbose)
	return nil
}

func newExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain [executable]",
		Short: "Explain how an executable was installed and why it is in PATH",
		Args:  cobra.ExactArgs(1),
		RunE:  runExplain,
	}
}

func runExplain(cmd *cobra.Command, args []string) error {
	name := args[0]
	rp, err := detector.ResolveExecPath(name)
	if err != nil {
		fmt.Printf("%s %s not found in PATH\n", crossMark, name)
		return err
	}

	f := parseFormat(format)
	r := detector.DetectWithRuntime(rp, name)

	explain := buildExplain(name, rp, r)
	r.Explain = explain
	output.PrintResult(r, f, verbose)
	return nil
}

func buildExplain(name string, rp detector.ResolvedPath, r *detector.Result) string {
	var b strings.Builder
	dir := filepath.Dir(rp.Real)
	home := detector.HomeDir()

	fmt.Fprintf(&b, "%s is located at:\n", name)
	fmt.Fprintf(&b, "  %s\n\n", rp.Real)

	if rp.Original != rp.Real {
		fmt.Fprintf(&b, "Symlink: %s\n", rp.Original)
		fmt.Fprintf(&b, "Target:  %s\n\n", rp.Real)
	}

	if strings.HasPrefix(dir, home) {
		b.WriteString("That directory is in your home folder.\n\n")
	} else if strings.HasPrefix(dir, "/usr") {
		b.WriteString("That directory is owned by the system.\n\n")
	}

	// Runtime explanation
	if r.Runtime != "" {
		fmt.Fprintf(&b, "Runtime: %s\n", r.Runtime)
		if r.RuntimeVersion != "" {
			fmt.Fprintf(&b, "Runtime version: %s\n", r.RuntimeVersion)
		}
		if r.RuntimeManager != "" {
			fmt.Fprintf(&b, "Runtime manager: %s\n", r.RuntimeManager)
		}
		b.WriteString("\n")
	}

	// Package manager explanation
	switch r.Manager {
	case detector.ManagerBun:
		b.WriteString("Bun manages global packages in ~/.bun/bin.\n")
		b.WriteString("Because PATH prioritizes ~/.bun/bin, the active executable comes from Bun.\n")
	case detector.ManagerNpm:
		b.WriteString("npm manages global packages in its global prefix.\n")
		b.WriteString("The active executable comes from npm's global installation.\n")
	case detector.ManagerPacman:
		b.WriteString("pacman installed this as a system package.\n")
	case detector.ManagerPipx:
		b.WriteString("pipx manages Python CLI tools in isolated virtual environments.\n")
		if r.VirtualEnv != "" {
			fmt.Fprintf(&b, "Virtual env: %s\n", r.VirtualEnv)
		}
	case detector.ManagerCargo:
		b.WriteString("cargo installed this from crates.io to ~/.cargo/bin.\n")
	case detector.ManagerGo:
		b.WriteString("go install placed this in $GOPATH/bin.\n")
	case detector.ManagerFlatpak:
		b.WriteString("flatpak installed this as a sandboxed application.\n")
	case detector.ManagerSnap:
		b.WriteString("snap installed this as a confined package.\n")
	case detector.ManagerBrew:
		b.WriteString("Homebrew (Linuxbrew) installed this.\n")
	case detector.ManagerNix:
		b.WriteString("Nix package manager installed this.\n")
	case detector.ManagerManual:
		b.WriteString("No package manager claims this executable.\n")
		b.WriteString("It was likely installed manually or built from source.\n")
	default:
		fmt.Fprintf(&b, "Installed by %s.\n", string(r.Manager))
	}

	return b.String()
}

func newAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Scan common Linux executables and list their package managers",
		Args:  cobra.NoArgs,
		RunE:  runAll,
	}
}

func runAll(cmd *cobra.Command, args []string) error {
	f := parseFormat(format)
	commonBins := []string{
		"git", "node", "npm", "python", "python3", "docker", "code",
		"go", "cargo", "rustc", "bun", "pnpm", "yarn", "pip", "pip3",
		"flatpak", "snap", "brew", "gem", "composer", "java", "ruby",
		"php", "lua", "rust-analyzer", "lazygit", "ffmpeg", "jq", "rg",
		"fd", "bat", "exa", "delta", "fzf", "ripgrep", "tealdeer",
	}

	var results []*detector.Result
	var misses []detector.ResolvedPath
	var missNames []string
	for _, bin := range commonBins {
		rp, err := detector.ResolveExecPath(bin)
		if err != nil {
			continue
		}
		// Honor --no-cache for the bulk scan.
		if !noCache {
			if cached, found := cache.Get(rp.Real); found {
				copy := *cached
				copy.Executable = bin
				results = append(results, &copy)
				continue
			}
		}
		misses = append(misses, rp)
		missNames = append(missNames, bin)
	}

	for i, r := range detector.DetectBatch(misses, missNames) {
		if !noCache {
			cache.Set(misses[i].Real, r)
		}
		results = append(results, r)
	}

	if len(results) == 0 {
		fmt.Println("No packages detected.")
		return nil
	}

	output.PrintResults(results, f)
	return nil
}

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the detection cache",
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the cache",
		Run: func(cmd *cobra.Command, args []string) {
			cache.Clear()
			fmt.Println("Cache cleared.")
		},
	}

	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show cache statistics",
		Run: func(cmd *cobra.Command, args []string) {
			s := cache.Stats()
			fmt.Printf("Cached items: %d\n", s["items"])
		},
	}

	cmd.AddCommand(clearCmd, statsCmd)
	return cmd
}

var appVersion = "dev"

// SetVersion is called from main() with the value injected at build time via
// -ldflags "-X main.version=...". Falls back to "dev" for local builds.
func SetVersion(v string) {
	if v != "" {
		appVersion = v
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("spkg %s\n", appVersion)
		},
	}
}

func parseFormat(f string) output.Format {
	parsed, _ := output.ParseFormat(f)
	return parsed
}

var crossMark = "✗"

func printVerbose(name, execPath string, fromCache bool) {
	fmt.Fprintf(os.Stderr, "Searching executable %q...\n", name)
	fmt.Fprintf(os.Stderr, "Found %s\n", execPath)
	if fromCache {
		fmt.Fprintf(os.Stderr, "Result loaded from cache.\n")
	} else {
		fmt.Fprintf(os.Stderr, "Running package manager detectors...\n")
		fmt.Fprintf(os.Stderr, "Running runtime/version manager detectors...\n")
	}
}
