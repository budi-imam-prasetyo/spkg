package detector

import (
	"path/filepath"
	"testing"
)

func TestBandFor(t *testing.T) {
	cases := []struct {
		score int
		want  ConfidenceLevel
	}{
		{0, LevelUnknown},
		{30, LevelLow},
		{49, LevelLow},
		{50, LevelMedium},
		{89, LevelMedium},
		{90, LevelHigh},
		{99, LevelHigh},
		{100, LevelCertain},
	}
	for _, c := range cases {
		if got := BandFor(c.score); got != c.want {
			t.Errorf("BandFor(%d) = %q, want %q", c.score, got, c.want)
		}
	}
}

func TestCombineConfidenceManualFallback(t *testing.T) {
	score, reasons := ManualConfidence()
	if score != int(ConfidenceUnknown) {
		t.Errorf("manual confidence = %d, want 0", score)
	}
	if len(reasons) != 1 {
		t.Errorf("manual reasons = %d, want 1", len(reasons))
	}
}

func TestDetectWithRuntimeManualFallbackIsUnknown(t *testing.T) {
	rp := ResolvedPath{
		Original: "/tmp/spkg-unmanaged/spkg-unmanaged-xyz-123",
		Real:     "/tmp/spkg-unmanaged/spkg-unmanaged-xyz-123",
	}
	r := DetectWithRuntime(rp, "spkg-unmanaged-xyz-123")
	if r.Manager != ManagerManual {
		t.Fatalf("manager = %s, want manual", r.Manager)
	}
	if r.Confidence != int(ConfidenceUnknown) {
		t.Fatalf("manual confidence = %d, want 0", r.Confidence)
	}
	if len(r.Reasons) != 1 || r.Reasons[0] != "no package manager matched; assuming a manual install" {
		t.Fatalf("manual reasons = %#v, want manual fallback reason", r.Reasons)
	}
}

func TestCombineConfidenceDBBacked(t *testing.T) {
	score, reasons := CombineConfidence(combineInput{
		Manager:        ManagerPacman,
		StructuralBand: ConfidenceMedium,
		StructuralReasons: []Reason{
			{Signal: "path_match", Detail: "path is under a system install dir", Weight: int(ConfidenceMedium)},
		},
		Confirmed: true,
		ConfirmReasons: []Reason{
			{Signal: "manager_db", Detail: "pacman -Qo confirms ownership by package git", Weight: int(ConfidenceHigh)},
		},
		RuntimeOwnsPath: false,
		ExcludedBy:      5,
	})
	if score != int(ConfidenceHigh) {
		t.Errorf("db-backed confidence = %d, want %d", score, int(ConfidenceHigh))
	}
	foundDB := false
	for _, r := range reasons {
		if r.Signal == "manager_db" {
			foundDB = true
		}
	}
	if !foundDB {
		t.Errorf("expected a manager_db reason, got %+v", reasons)
	}
}

func TestCombineConfidenceExclusionBonusCapped(t *testing.T) {
	// Many exclusions should cap the bonus at +10, not push a low score to high.
	score, _ := CombineConfidence(combineInput{
		Manager:           ManagerBrew,
		StructuralBand:    ConfidenceLow,
		StructuralReasons: []Reason{{Signal: "path_match", Detail: "under brew prefix", Weight: int(ConfidenceLow)}},
		Confirmed:         false,
		ExcludedBy:        50,
	})
	if score > int(ConfidenceLow)+10 {
		t.Errorf("exclusion bonus not capped: score=%d", score)
	}
}

func TestScorePacman(t *testing.T) {
	d := &PacmanDetector{}
	band, reasons := d.Score("/usr/bin/git")
	if band != ConfidenceMedium {
		t.Errorf("Score(/usr/bin/git) band = %d, want %d", band, int(ConfidenceMedium))
	}
	if len(reasons) == 0 {
		t.Errorf("expected at least one reason")
	}
	band, _ = d.Score("/home/u/.local/share/fnm/node")
	if band != ConfidenceUnknown {
		t.Errorf("Score(fnmdir) band = %d, want unknown", band)
	}
}

func TestScoreFnmRuntimeDir(t *testing.T) {
	// npm Score() should flag node-runtime dirs as medium (structural).
	d := &NpmDetector{}
	band, _ := d.Score("/home/u/.local/share/fnm/versions/node/v20/bin/node")
	if band != ConfidenceMedium {
		t.Errorf("Score(fnm node) band = %d, want medium", band)
	}
	band, _ = d.Score("/usr/bin/curl")
	if band != ConfidenceUnknown {
		t.Errorf("Score(curl) band = %d, want unknown", band)
	}
}

func TestScoreCargo(t *testing.T) {
	d := &CargoDetector{}
	home := HomeDir()
	band, _ := d.Score(filepath.Join(home, ".cargo", "bin", "ripgrep"))
	if band != ConfidenceMedium {
		t.Errorf("Score(cargo bin) band = %d, want medium", band)
	}
	band, _ = d.Score("/usr/bin/ls")
	if band != ConfidenceUnknown {
		t.Errorf("Score(/usr/bin/ls) band = %d, want unknown", band)
	}
}

func TestDetectWithRuntimeConfidence_Pacman(t *testing.T) {
	rp := ResolvedPath{Original: "/usr/bin/git", Real: "/usr/bin/git"}
	r := DetectWithRuntime(rp, "git")
	if r.Manager != ManagerPacman {
		t.Fatalf("manager = %s, want pacman", r.Manager)
	}
	if r.Confidence < int(ConfidenceHigh) {
		t.Errorf("pacman git confidence = %d, want >= %d (high)", r.Confidence, int(ConfidenceHigh))
	}
	if len(r.Reasons) == 0 {
		t.Errorf("expected reasons for pacman detection")
	}
}

func TestRuntimeOwnerCommands(t *testing.T) {
	r := &Result{
		Manager:        ManagerNpm,
		Package:        "node",
		Runtime:        RuntimeFnm,
		RuntimeVersion: "Node.js 24.20.0",
	}
	applyRuntimeOwner(r, RuntimeFnm, "24.20.0")
	if r.Manager != ManagerFnm {
		t.Fatalf("manager = %s, want fnm", r.Manager)
	}
	if r.InstallCmd != "fnm install 24.20.0" || r.UpgradeCmd != "fnm install --latest" || r.RemoveCmd != "fnm uninstall 24.20.0" {
		t.Fatalf("unexpected fnm commands: install=%q upgrade=%q remove=%q", r.InstallCmd, r.UpgradeCmd, r.RemoveCmd)
	}
}

func TestDetectWithRuntimeConfidence_FnmNode(t *testing.T) {
	if !HasCommand("fnm") {
		t.Skip("fnm not installed; skipping runtime-owner confidence test")
	}
	home := HomeDir()
	rp := ResolvedPath{
		Original: filepath.Join(home, ".local", "share", "fnm", "node-versions", "v24.15.0", "installation", "bin", "node"),
		Real:     filepath.Join(home, ".local", "share", "fnm", "node-versions", "v24.15.0", "installation", "bin", "node"),
	}
	r := DetectWithRuntime(rp, "node")
	if r.Manager != ManagerFnm {
		t.Fatalf("manager = %s, want fnm", r.Manager)
	}
	if !r.RuntimeOwnsPathLabel() {
		t.Errorf("expected runtime-owned flag set")
	}
	if r.Confidence <= int(ConfidenceLow) {
		t.Errorf("fnm node confidence = %d, want above low", r.Confidence)
	}
}

// RuntimeOwnsPathLabel is a tiny convenience for tests: true when the manager is a
// runtime/version manager owner.
func (r *Result) RuntimeOwnsPathLabel() bool {
	return r.Manager == ManagerFnm || r.Manager == ManagerNvm ||
		r.Manager == ManagerVolta || r.Manager == ManagerRustup
}
