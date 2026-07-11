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

func TestRuntimeRegistry(t *testing.T) {
	all := GetAllRuntimes()
	if len(all) == 0 {
		t.Error("expected at least one registered runtime detector")
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

func TestDetectWithRuntime(t *testing.T) {
	r := DetectWithRuntime("/usr/bin/ls", "ls")
	if r == nil {
		t.Fatal("expected to find ls")
	}
	if !r.Found {
		t.Error("expected Found to be true")
	}
	if r.Confidence == 0 {
		t.Error("expected non-zero confidence")
	}
}

func TestDetectRuntime(t *testing.T) {
	rt, version, manager, confidence := DetectRuntime("/usr/bin/ls")
	// ls is not managed by any runtime
	if rt != RuntimeNone {
		t.Logf("Runtime detected: %s %s (manager: %s, confidence: %d)", rt, version, manager, confidence)
	}
}

func TestExtractVersionFromPath(t *testing.T) {
	tests := []struct {
		path    string
		markers []string
		want    string
	}{
		{
			path:    "/home/user/.local/share/fnm/node-versions/v22.18.0/installation/bin/node",
			markers: []string{"node-versions/v", "node-versions/"},
			want:    "22.18.0",
		},
		{
			path:    "/home/user/.nvm/versions/node/v20.10.0/bin/node",
			markers: []string{"versions/node/v"},
			want:    "20.10.0",
		},
		{
			path:    "/home/user/.pyenv/versions/3.12.0/bin/python",
			markers: []string{"versions/"},
			want:    "3.12.0",
		},
	}

	for _, tt := range tests {
		got := ExtractVersionFromPath(tt.path, tt.markers...)
		if got != tt.want {
			t.Errorf("ExtractVersionFromPath(%q, %v) = %q, want %q", tt.path, tt.markers, got, tt.want)
		}
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

func TestRuntimeConstants(t *testing.T) {
	runtimes := []RuntimeType{
		RuntimeNone, RuntimeFnm, RuntimeNvm, RuntimeVolta,
		RuntimeAsdf, RuntimeMise, RuntimePyenv, RuntimeConda,
		RuntimeRbenv, RuntimeRVM, RuntimeSdkman, RuntimeRustup,
		RuntimeGvm, RuntimeGoenv, RuntimePhpenv, RuntimeKerl,
	}
	for _, rt := range runtimes {
		if rt == RuntimeNone {
			continue
		}
		if rt == "" {
			t.Error("runtime should not be empty")
		}
	}
}

func BenchmarkDetectFirst(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetectFirst("/usr/bin/ls")
	}
}

func BenchmarkDetectWithRuntime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetectWithRuntime("/usr/bin/ls", "ls")
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

func BenchmarkExtractVersionFromPath(b *testing.B) {
	path := "/home/user/.local/share/fnm/node-versions/v22.18.0/installation/bin/node"
	for i := 0; i < b.N; i++ {
		ExtractVersionFromPath(path, "node-versions/v", "node-versions/")
	}
}
