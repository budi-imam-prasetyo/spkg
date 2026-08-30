package detector

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

// fakeSystem lets tests drive HasCommand/RunCmd without depending on whatever
// tooling happens to be installed on the host.
type fakeSystem struct {
	available map[string]string
	stdout    string
	exitCode  int
	calls     []string
}

func (f *fakeSystem) LookPath(file string) (string, error) {
	if path, ok := f.available[file]; ok {
		return path, nil
	}
	return "", errors.New("not found")
}

func (f *fakeSystem) CmdContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	f.calls = append(f.calls, name)
	if f.exitCode != 0 {
		return exec.CommandContext(ctx, "sh", "-c", "exit "+itoa(f.exitCode))
	}
	return exec.CommandContext(ctx, "printf", "%s", f.stdout)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func TestHasCommandUsesInjectedSystem(t *testing.T) {
	fake := &fakeSystem{available: map[string]string{"spkg-fake-tool": "/fake/bin/spkg-fake-tool"}}
	SetSystem(fake)
	t.Cleanup(func() { SetSystem(nil) })

	if !HasCommand("spkg-fake-tool") {
		t.Fatal("HasCommand should report the injected tool as available")
	}
	if HasCommand("spkg-absent-tool") {
		t.Fatal("HasCommand should not invent tools that the System rejects")
	}
}

func TestRunCmdUsesInjectedSystem(t *testing.T) {
	fake := &fakeSystem{stdout: "injected-output\n"}
	SetSystem(fake)
	t.Cleanup(func() { SetSystem(nil) })

	out, err := RunCmd("pacman", "-Qo", "/bin/git")
	if err != nil {
		t.Fatalf("RunCmd: %v", err)
	}
	if out != "injected-output" {
		t.Fatalf("out = %q, want injected-output", out)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "pacman" {
		t.Fatalf("calls = %#v, want one pacman call", fake.calls)
	}
}

func TestSetSystemNilRestoresRealSystem(t *testing.T) {
	SetSystem(&fakeSystem{})
	SetSystem(nil)
	if _, ok := currentSystem().(realSystem); !ok {
		t.Fatalf("currentSystem() = %T, want realSystem", currentSystem())
	}
}
