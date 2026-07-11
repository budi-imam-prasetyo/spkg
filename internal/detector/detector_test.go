package detector

import (
	"testing"
)

func TestResolveExecPath(t *testing.T) {
	path, err := ResolveExecPath("ls")
	if err != nil {
		t.Fatalf("expected to find ls, got error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestResolveExecPath_NotFound(t *testing.T) {
	_, err := ResolveExecPath("nonexistent_binary_xyz_123")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
}

func TestHasCommand(t *testing.T) {
	if !HasCommand("ls") {
		t.Error("expected ls to be available")
	}
	if HasCommand("nonexistent_binary_xyz_123") {
		t.Error("expected nonexistent binary to not be available")
	}
}

func TestRunCmd(t *testing.T) {
	out, err := RunCmd("echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Errorf("expected 'hello', got %q", out)
	}
}

func TestRunCmd_Error(t *testing.T) {
	_, err := RunCmd("false")
	if err == nil {
		t.Fatal("expected error from false command")
	}
}

func TestBoolPtr(t *testing.T) {
	truePtr := BoolPtr(true)
	if truePtr == nil || !*truePtr {
		t.Error("expected true pointer")
	}
	falsePtr := BoolPtr(false)
	if falsePtr == nil || *falsePtr {
		t.Error("expected false pointer")
	}
}

func TestManagerConstants(t *testing.T) {
	managers := []Manager{
		ManagerPacman, ManagerBun, ManagerNpm, ManagerPnpm,
		ManagerYarn, ManagerPip, ManagerPipx, ManagerUv,
		ManagerCargo, ManagerGo, ManagerGem, ManagerComposer,
		ManagerFlatpak, ManagerSnap, ManagerBrew, ManagerNix,
		ManagerDocker, ManagerPodman, ManagerAppImage, ManagerManual,
	}
	for _, m := range managers {
		if m == "" {
			t.Error("manager should not be empty")
		}
	}
}

func TestRegistry(t *testing.T) {
	all := GetAll()
	if len(all) == 0 {
		t.Error("expected at least one registered detector")
	}
}

func TestDetectFirst(t *testing.T) {
	r := DetectFirst("/usr/bin/ls")
	if r == nil {
		t.Fatal("expected to find ls")
	}
	if !r.Found {
		t.Error("expected Found to be true")
	}
}

func TestPriorityOrder(t *testing.T) {
	if priorityOrder(ManagerPacman) >= priorityOrder(ManagerManual) {
		t.Error("pacman should have higher priority than manual")
	}
	if priorityOrder(ManagerBun) >= priorityOrder(ManagerManual) {
		t.Error("bun should have higher priority than manual")
	}
}

func BenchmarkDetectFirst(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetectFirst("/usr/bin/ls")
	}
}

func BenchmarkHasCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HasCommand("ls")
	}
}

func BenchmarkRunCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunCmd("echo", "test")
	}
}
