package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ryoukaii/spkg/internal/cache"
	"github.com/ryoukaii/spkg/internal/detector"
	"github.com/ryoukaii/spkg/internal/output"
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
		Use:   "pkg-origin [command]",
		Short: "Detect which package manager installed an executable",
		Long: `pkg-origin detects which package manager installed a given executable.
It supports pacman, paru, yay, bun, npm, pnpm, yarn, pip, pipx, uv,
cargo, go, gem, composer, flatpak, snap, homebrew, nix, docker, podman,
and more.

It also detects runtime/version managers: fnm, nvm, volta, asdf, mise,
pyenv, conda, rbenv, RVM, SDKMAN!, rustup, and others.`,
		Args: cobra.ArbitraryArgs,
		RunE: runDetect,
	}

	root.PersistentFlags().StringVarP(&format, "output", "o", "pretty", "Output format: pretty, json, yaml, table, plain")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose detection output")
	root.PersistentFlags().BoolVar(&noCache, "no-cache", false, "Skip cache")
	root.PersistentFlags().IntVar(&cacheTTLSec, "cache-ttl", 600, "Cache TTL in seconds")

	root.AddCommand(newCacheCmd())
	root.AddCommand(newExplainCmd())
	root.AddCommand(newAllCmd())
	root.AddCommand(newVersionCmd())

	return root
}

func runDetect(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	name := args[0]
	f := parseFormat(format)

	execPath, err := detector.ResolveExecPath(name)
	if err != nil {
		output.PrintNotFound(name, f)
		return nil
	}

	if !noCache {
		if cached, found := cache.Get(execPath); found {
			if verbose {
				printVerbose(name, execPath, true)
			}
			cached.Executable = name
			output.PrintResult(cached, f, verbose)
			return nil
		}
	}

	if verbose {
		printVerbose(name, execPath, false)
	}

	r := detector.DetectWithRuntime(execPath, name)

	if !noCache {
		cache.Set(execPath, r)
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
	execPath, err := detector.ResolveExecPath(name)
	if err != nil {
		fmt.Printf("%s %s not found in PATH\n", crossMark, name)
		return nil
	}

	f := parseFormat(format)
	r := detector.DetectWithRuntime(execPath, name)

	explain := buildExplain(name, execPath, r)
	r.Explain = explain
	output.PrintResult(r, f, verbose)
	return nil
}

func buildExplain(name, execPath string, r *detector.Result) string {
	var b strings.Builder
	dir := filepath.Dir(execPath)
	home, _ := os.UserHomeDir()

	b.WriteString(fmt.Sprintf("%s is located at:\n", name))
	b.WriteString(fmt.Sprintf("  %s\n\n", execPath))

	if strings.HasPrefix(dir, home) {
		b.WriteString("That directory is in your home folder.\n\n")
	} else if strings.HasPrefix(dir, "/usr") {
		b.WriteString("That directory is owned by the system.\n\n")
	}

	// Runtime explanation
	if r.Runtime != "" {
		b.WriteString(fmt.Sprintf("Runtime: %s\n", r.Runtime))
		if r.RuntimeVersion != "" {
			b.WriteString(fmt.Sprintf("Runtime version: %s\n", r.RuntimeVersion))
		}
		if r.RuntimeManager != "" {
			b.WriteString(fmt.Sprintf("Runtime manager: %s\n", r.RuntimeManager))
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
			b.WriteString(fmt.Sprintf("Virtual env: %s\n", r.VirtualEnv))
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
		b.WriteString(fmt.Sprintf("Installed by %s.\n", string(r.Manager)))
	}

	return b.String()
}

func newAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "List all detectable executables and their package managers",
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
	for _, bin := range commonBins {
		execPath, err := detector.ResolveExecPath(bin)
		if err != nil {
			continue
		}
		r := detector.DetectWithRuntime(execPath, bin)
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

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pkg-origin v1.1.0")
		},
	}
}

func parseFormat(f string) output.Format {
	switch strings.ToLower(f) {
	case "json":
		return output.FormatJSON
	case "yaml":
		return output.FormatYAML
	case "table":
		return output.FormatTable
	case "plain":
		return output.FormatPlain
	default:
		return output.FormatPretty
	}
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
