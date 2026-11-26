package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ryoukaii/spkg/internal/detector"
)

type aliasTestSystem struct{ dir string }

func (s aliasTestSystem) LookPath(name string) (string, error) {
	if name == "first" || name == "second" {
		return filepath.Join(s.dir, name), nil
	}
	return "", fmt.Errorf("absent")
}
func (s aliasTestSystem) CmdContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", "exit 1")
}
func TestCacheSeparatesSymlinkAliases(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	real := filepath.Join(dir, "real")
	if err := os.WriteFile(real, []byte("x"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"first", "second"} {
		if err := os.Symlink(real, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	detector.SetSystem(aliasTestSystem{dir})
	defer detector.SetSystem(nil)
	var last detector.Result
	for _, name := range []string{"first", "second", "first", "second"} {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{name, "-o", "json"})
		old := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w
		err = cmd.Execute()
		closeWriteErr := w.Close()
		os.Stdout = old
		if closeWriteErr != nil {
			t.Fatal(closeWriteErr)
		}
		data, readErr := io.ReadAll(r)
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		if readErr != nil {
			t.Fatal(readErr)
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &last); err != nil {
			t.Fatal(err)
		}
		if last.Package != name || last.Executable != name {
			t.Fatalf("query %q returned package=%q executable=%q", name, last.Package, last.Executable)
		}
	}
	if last.Package != "second" {
		t.Fatalf("query second returned executable=%q but cached package=%q", last.Executable, last.Package)
	}
}
