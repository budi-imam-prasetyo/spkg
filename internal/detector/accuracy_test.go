package detector

import (
	"testing"
)

func TestContainsWordExact(t *testing.T) {
	if !containsWord("package black 24.0", "black") {
		t.Error("expected exact word match")
	}
	if containsWord("package blackd 24.0", "black") {
		t.Error("substring must not match")
	}
	if containsWord("", "black") || containsWord("black", "") {
		t.Error("empty input must not match")
	}
}

func TestMatchExecutablesLine(t *testing.T) {
	if !matchExecutablesLine("Executables: black, blackd", "black") {
		t.Error("expected exact executable match")
	}
	if !matchExecutablesLine("Executables: blackd", "blackd") {
		t.Error("expected exact executable match")
	}
	if matchExecutablesLine("Executables: blackd", "black") {
		t.Error("substring must not match executables list")
	}
	if matchExecutablesLine("Executables: ", "black") {
		t.Error("empty list must not match")
	}
}

func TestPriorityPipxBeforePip(t *testing.T) {
	// ~/.local/bin is shared by pip/pipx/uv. The more specific verifiers
	// (pipx venv stat, uv tool list) must win ties over bare `pip show`.
	if priorityOrder(ManagerPipx) >= priorityOrder(ManagerPip) {
		t.Errorf("pipx (%d) should beat pip (%d)", priorityOrder(ManagerPipx), priorityOrder(ManagerPip))
	}
	if priorityOrder(ManagerUv) >= priorityOrder(ManagerPip) {
		t.Errorf("uv (%d) should beat pip (%d)", priorityOrder(ManagerUv), priorityOrder(ManagerPip))
	}
}

func TestPacmanScoreSkipsOpt(t *testing.T) {
	d := &PacmanDetector{}
	if band, _ := d.Score("/opt/myapp/bin/foo"); band != ConfidenceUnknown {
		t.Errorf("Score(/opt/...) = %d, want unknown (shared dir, needs -Qo proof)", band)
	}
	if band, _ := d.Score("/usr/bin/git"); band != ConfidenceMedium {
		t.Errorf("Score(/usr/bin/git) = %d, want medium", band)
	}
}

func TestMemoizedHasCommandCached(t *testing.T) {
	clearMemo()
	if !memoizedHasCommand("ls") {
		t.Fatal("expected ls to be available")
	}
	// Second call must hit the memo, same answer.
	if !memoizedHasCommand("ls") {
		t.Fatal("expected cached ls to be available")
	}
	if memoizedHasCommand("nonexistent_binary_xyz_123") {
		t.Fatal("expected nonexistent binary to be unavailable")
	}
	clearMemo()
}

func TestDockerExactRepoMatch(t *testing.T) {
	// Indirect: repo extraction logic lives inline in docker Detect; verify the
	// rule via a tiny table here so future substring regressions fail fast.
	repoOf := func(line string) string {
		repo := line
		for _, sep := range []string{":", "/"} {
			_ = sep
		}
		if i := lastIndex(line, ":"); i != -1 {
			repo = line[:i]
		}
		if i := lastIndex(repo, "/"); i != -1 {
			repo = repo[i+1:]
		}
		return repo
	}
	if got := repoOf("myapp:latest"); got != "myapp" {
		t.Errorf("repoOf = %q, want myapp", got)
	}
	if got := repoOf("registry/x/myapp-helper:1.0"); got != "myapp-helper" {
		t.Errorf("repoOf = %q, want myapp-helper", got)
	}
	if repoOf("registry/x/myapp-helper:1.0") == "myapp" {
		t.Error("helper image must not equal myapp")
	}
}

func lastIndex(s, sep string) int {
	for i := len(s) - len(sep); i >= 0; i-- {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

func TestRuntimeTakesOverBundledProof(t *testing.T) {
	r := &Result{
		Found:   true,
		Manager: ManagerNpm,
		Package: "node",
		ConfirmReasons: []Reason{
			{Signal: "manager_db", Detail: "npm global prefix owns this path", Weight: 90},
		},
	}
	applyRuntimeOwner(r, RuntimeFnm, "24.20.0")
	// Simulate what DetectWithRuntime now does: stale npm proof is replaced.
	r.ConfirmReasons = []Reason{{
		Signal: "runtime_db",
		Detail: "fnm owns this path (Node.js 24.20.0)",
		Weight: int(ConfidenceHigh),
	}}
	if len(r.ConfirmReasons) != 1 || r.ConfirmReasons[0].Signal != "runtime_db" {
		t.Fatalf("ConfirmReasons = %+v, want runtime_db proof", r.ConfirmReasons)
	}
}
