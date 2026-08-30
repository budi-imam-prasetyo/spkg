package cli

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ryoukaii/spkg/internal/cache"
)

func TestRootDefaultsToTwoColumnTable(t *testing.T) {
	cmd := NewRootCmd()
	flag := cmd.PersistentFlags().Lookup("output")
	if flag == nil || flag.DefValue != "table" {
		t.Fatalf("output default = %v, want table", flag)
	}
}

func TestRootRejectsInvalidOutputFormat(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"git", "--output", "invalid"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "invalid output format") {
		t.Fatalf("Execute error = %v, want invalid output format", err)
	}
}

func TestRootAcceptsMultipleExecutables(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"git", "definitely-not-a-real-binary"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("batch mode should not fail when some executables exist: %v", err)
	}
}

func TestRootBatchAllMissingFails(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"definitely-not-a-real-binary", "also-not-real"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error when no executables resolve")
	}
}

// Regression: output order must follow input order even when some args are
// already cached and others need detection.
func TestRootBatchPreservesInputOrder(t *testing.T) {
	// Use coreutils every Linux distro ships with a stable basename. python3,
	// node, etc. are symlinks whose resolved basename varies across distros
	// (python3 -> python3.12 on Ubuntu), which makes the position assertion
	// brittle. We only assert ordering, not the manager.
	bins := []string{"ls", "cat", "grep"}
	for _, b := range bins {
		if _, err := exec.LookPath(b); err != nil {
			t.Skipf("%s not installed; cannot test order deterministically", b)
		}
	}
	hit := bins[1] // "cat": the cache-hit entry, sits in the middle.

	// Clear so the warm-up below is the only cache hit in the batch run.
	cache.Clear()

	// Warm the cache for "cat" only, so it's a cache hit while the others
	// are misses, and it sits in the middle of the argument list.
	cmd := NewRootCmd()
	_ = cmd.PersistentFlags().Set("output", "plain")
	cmd.SetArgs([]string{hit})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	_ = cmd.Execute()

	cmd = NewRootCmd()
	_ = cmd.PersistentFlags().Set("output", "plain")
	cmd.SetArgs(bins)

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	err = cmd.Execute()
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("batch execute: %v", err)
	}

	var out bytes.Buffer
	if _, err := out.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}

	// Find the first data row of each executable and compare their positions.
	pos := func(name string) int {
		for i, line := range strings.Split(out.String(), "\n") {
			if strings.HasPrefix(line, name+"	") || strings.HasPrefix(line, name+" ") {
				return i
			}
		}
		return -1
	}

	a, b, c := pos(bins[0]), pos(bins[1]), pos(bins[2])
	if a < 0 || b < 0 || c < 0 {
		t.Fatalf("missing an entry in output:\n%s", out.String())
	}
	if a >= b || b >= c {
		t.Fatalf("output order not preserved (%s=%d %s=%d %s=%d):\n%s", bins[0], a, bins[1], b, bins[2], c, out.String())
	}
}

func TestRootSupportsVersionFlag(t *testing.T) {
	SetVersion("v-test")
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute --version: %v", err)
	}
	if !strings.Contains(out.String(), "v-test") {
		t.Fatalf("version output = %q, want v-test", out.String())
	}
}
