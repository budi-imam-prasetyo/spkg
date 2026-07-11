package output

import (
	"bytes"
	"os"
	"testing"

	"github.com/ryoukaii/spkg/internal/detector"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintResult_JSON(t *testing.T) {
	r := &detector.Result{
		Found:          true,
		ExecutablePath: "/usr/bin/test",
		Manager:        detector.ManagerPacman,
		Package:        "test-pkg",
		Version:        "1.0.0",
		InstallCmd:     "sudo pacman -S test-pkg",
		UpgradeCmd:     "sudo pacman -Syu test-pkg",
		RemoveCmd:      "sudo pacman -Rns test-pkg",
	}

	output := captureStdout(t, func() {
		PrintResult(r, FormatJSON, false)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPrintResult_YAML(t *testing.T) {
	r := &detector.Result{
		Found:          true,
		ExecutablePath: "/usr/bin/test",
		Manager:        detector.ManagerBun,
		Package:        "test-pkg",
	}

	output := captureStdout(t, func() {
		PrintResult(r, FormatYAML, false)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPrintResult_Table(t *testing.T) {
	r := &detector.Result{
		Found:          true,
		ExecutablePath: "/usr/bin/test",
		Manager:        detector.ManagerPacman,
		Package:        "test-pkg",
		Version:        "1.0.0",
	}

	output := captureStdout(t, func() {
		PrintResult(r, FormatTable, false)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPrintResult_Plain(t *testing.T) {
	r := &detector.Result{
		Found:          true,
		ExecutablePath: "/usr/bin/test",
		Manager:        detector.ManagerPacman,
		Package:        "test-pkg",
	}

	output := captureStdout(t, func() {
		PrintResult(r, FormatPlain, false)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPrintResult_Pretty(t *testing.T) {
	r := &detector.Result{
		Found:          true,
		ExecutablePath: "/usr/bin/test",
		Manager:        detector.ManagerPacman,
		Package:        "test-pkg",
		Version:        "1.0.0",
	}

	output := captureStdout(t, func() {
		PrintResult(r, FormatPretty, false)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPrintNotFound(t *testing.T) {
	output := captureStdout(t, func() {
		PrintNotFound("nonexistent", FormatPretty)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPrintNotFound_JSON(t *testing.T) {
	output := captureStdout(t, func() {
		PrintNotFound("nonexistent", FormatJSON)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPrintResults(t *testing.T) {
	results := []*detector.Result{
		{Found: true, ExecutablePath: "/usr/bin/git", Manager: detector.ManagerPacman, Package: "git"},
		{Found: true, ExecutablePath: "/usr/bin/node", Manager: detector.ManagerNpm, Package: "node"},
	}

	output := captureStdout(t, func() {
		PrintResults(results, FormatTable)
	})

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func BenchmarkPrintResult_JSON(b *testing.B) {
	r := &detector.Result{
		Found:          true,
		ExecutablePath: "/usr/bin/test",
		Manager:        detector.ManagerPacman,
		Package:        "test-pkg",
	}

	for i := 0; i < b.N; i++ {
		PrintResult(r, FormatJSON, false)
	}
}

func BenchmarkPrintResult_Table(b *testing.B) {
	r := &detector.Result{
		Found:          true,
		ExecutablePath: "/usr/bin/test",
		Manager:        detector.ManagerPacman,
		Package:        "test-pkg",
	}

	for i := 0; i < b.N; i++ {
		PrintResult(r, FormatTable, false)
	}
}
