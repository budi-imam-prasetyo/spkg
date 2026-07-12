package detector

import (
	"testing"
)

func TestResolveExecPath(t *testing.T) {
	rp, err := ResolveExecPath("ls")
	if err != nil {
		t.Fatalf("expected to find ls, got error: %v", err)
	}
	if rp.Original == "" {
		t.Fatal("expected non-empty original path")
	}
	if rp.Real == "" {
		t.Fatal("expected non-empty real path")
	}
}

func TestResolveExecPath_NotFound(t *testing.T) {
	_, err := ResolveExecPath("nonexistent_binary_xyz_123")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
}

func TestResolvedPath_Both(t *testing.T) {
	rp := ResolvedPath{Original: "/usr/bin/git", Real: "/usr/bin/git"}
	both := rp.Both()
	if len(both) != 1 {
		t.Errorf("expected 1 path when same, got %d", len(both))
	}

	rp2 := ResolvedPath{Original: "/bin/git", Real: "/usr/bin/git"}
	both2 := rp2.Both()
	if len(both2) != 2 {
		t.Errorf("expected 2 paths when different, got %d", len(both2))
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
	rp := ResolvedPath{Original: "/usr/bin/ls", Real: "/usr/bin/ls"}
	r := DetectFirst(rp)
	if r == nil {
		t.Fatal("expected to find ls")
	}
	if !r.Found {
		t.Error("expected Found to be true")
	}
}

func TestDetectWithRuntime(t *testing.T) {
	rp := ResolvedPath{Original: "/usr/bin/ls", Real: "/usr/bin/ls"}
	r := DetectWithRuntime(rp, "ls")
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
	rp := ResolvedPath{Original: "/usr/bin/ls", Real: "/usr/bin/ls"}
	rt, version, manager, confidence := DetectRuntime(rp)
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

func TestBuildRemoveCmd(t *testing.T) {
	home := HomeDir()
	tests := []struct {
		execPath string
		want     string
	}{
		{
			execPath: home + "/.bun/install/global/node_modules/npkill/lib/index.js",
			want:     "bun remove -g npkill",
		},
		{
			execPath: home + "/.cargo/bin/rg",
			want:     "cargo uninstall rg",
		},
		{
			execPath: "/usr/bin/git",
			want:     "sudo rm /usr/bin/git",
		},
	}

	for _, tt := range tests {
		got := buildRemoveCmd(tt.execPath)
		if got != tt.want {
			t.Errorf("buildRemoveCmd(%q) = %q, want %q", tt.execPath, got, tt.want)
		}
	}
}

func TestExtractBunPackageName(t *testing.T) {
	home := HomeDir()
	tests := []struct {
		path string
		want string
	}{
		{
			path: home + "/.bun/install/global/node_modules/npkill/lib/index.js",
			want: "npkill",
		},
		{
			path: home + "/.bun/install/global/node_modules/@scope/pkg/lib/index.js",
			want: "@scope/pkg",
		},
		{
			path: home + "/.cargo/bin/rg",
			want: "",
		},
	}

	for _, tt := range tests {
		got := extractBunPackageName(tt.path)
		if got != tt.want {
			t.Errorf("extractBunPackageName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestExtractNpmPackageName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{
			path: "/home/user/.local/share/fnm/node-versions/v22.18.0/installation/lib/node_modules/npm/bin/npm-cli.js",
			want: "npm",
		},
		{
			path: "/usr/lib/node_modules/eslint/bin/eslint.js",
			want: "eslint",
		},
		{
			path: "/home/user/.cargo/bin/rg",
			want: "",
		},
	}

	for _, tt := range tests {
		got := extractNpmPackageName(tt.path)
		if got != tt.want {
			t.Errorf("extractNpmPackageName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func BenchmarkDetectFirst(b *testing.B) {
	rp := ResolvedPath{Original: "/usr/bin/ls", Real: "/usr/bin/ls"}
	for i := 0; i < b.N; i++ {
		DetectFirst(rp)
	}
}

func BenchmarkDetectWithRuntime(b *testing.B) {
	rp := ResolvedPath{Original: "/usr/bin/ls", Real: "/usr/bin/ls"}
	for i := 0; i < b.N; i++ {
		DetectWithRuntime(rp, "ls")
	}
}

func BenchmarkHasCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HasCommand("ls")
	}
}

func BenchmarkRunCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = RunCmd("echo", "test")
	}
}

func BenchmarkExtractVersionFromPath(b *testing.B) {
	path := "/home/user/.local/share/fnm/node-versions/v22.18.0/installation/bin/node"
	for i := 0; i < b.N; i++ {
		ExtractVersionFromPath(path, "node-versions/v", "node-versions/")
	}
}
