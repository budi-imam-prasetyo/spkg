package detector

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type queryTestSystem struct {
	prefix string
	calls  atomic.Int32
	fail   bool
}

func (s *queryTestSystem) LookPath(name string) (string, error) {
	if name == "npm" || name == "bun" {
		return "/mock/" + name, nil
	}
	return "", fmt.Errorf("absent")
}
func (s *queryTestSystem) CmdContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	out := ""
	switch name + " " + strings.Join(args, " ") {
	case "npm prefix -g":
		out = s.prefix
	case "npm list -g --depth=0 --json":
		out = `{"dependencies":{"typescript":{"version":"5.0.0"}}}`
	case "bun --version":
		out = "1.3.0"
	case "audit shared":
		s.calls.Add(1)
		time.Sleep(100 * time.Millisecond)
		if s.fail {
			return exec.CommandContext(ctx, "sh", "-c", "exit 9")
		}
		out = "ok"
	}
	return exec.CommandContext(ctx, "printf", "%s", out)
}
func TestNpmRejectsSharedBinWithoutPackageEvidence(t *testing.T) {
	SetSystem(&queryTestSystem{prefix: "/usr/local"})
	defer SetSystem(nil)
	r := (&NpmDetector{}).Detect("/usr/local/bin/unrelated-native-app")
	if r != nil {
		t.Fatalf("unrelated executable claimed as %s package %s with proof %v", r.Manager, r.Package, r.ConfirmReasons)
	}
}
func TestNpmSymlinkUsesPackageName(t *testing.T) {
	SetSystem(&queryTestSystem{prefix: "/audit/node"})
	defer SetSystem(nil)
	r := DetectFirst(ResolvedPath{Original: "/audit/node/bin/tsc", Real: "/audit/node/lib/node_modules/typescript/bin/tsc"})
	if r == nil || r.Package != "typescript" {
		t.Fatalf("want typescript, got %+v", r)
	}
}

func TestNpmResolvesRealSymlinks(t *testing.T) {
	prefix := t.TempDir()
	SetSystem(&queryTestSystem{prefix: prefix})
	defer SetSystem(nil)
	for _, tc := range []struct{ bin, pkg, script string }{
		{"tsc", "typescript", "bin/tsc"},
		{"tool", "@scope/toolkit", "dist/cli.js"},
		{"npx", "npm", "bin/npx-cli.js"},
	} {
		t.Run(tc.bin, func(t *testing.T) {
			target := filepath.Join(prefix, "lib/node_modules", tc.pkg, tc.script)
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, []byte("#!/bin/sh\n"), 0700); err != nil {
				t.Fatal(err)
			}
			binDir := filepath.Join(prefix, "bin")
			if err := os.MkdirAll(binDir, 0700); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(binDir, tc.bin)
			if err := os.Symlink(target, alias); err != nil {
				t.Fatal(err)
			}
			r := (&NpmDetector{}).Detect(alias)
			if r == nil || r.Package != tc.pkg {
				t.Fatalf("want package %q, got %+v", tc.pkg, r)
			}
		})
	}
}
func TestBunRuntimePreservesOwnerAndCommands(t *testing.T) {
	SetSystem(&queryTestSystem{})
	defer SetSystem(nil)
	p := filepath.Join(HomeDir(), ".bun", "bin", "bun")
	r := DetectWithRuntime(ResolvedPath{Original: p, Real: p}, "bun")
	if r.Manager != ManagerBun || r.UpgradeCmd != "bun upgrade" || r.Package != "bun" {
		t.Fatalf("bun owner changed to %s; upgrade=%q", r.Manager, r.UpgradeCmd)
	}
}

func TestRuntimeVersionsExcludeExecutablePath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, tc := range []struct {
		detector   RuntimeDetector
		path, want string
	}{
		{&MiseDetector{}, ".local/share/mise/installs/node/24.0.0/bin/node", "Node 24.0.0"},
		{&AsdfDetector{}, ".asdf/installs/python/3.13.0/bin/python", "Python 3.13.0"},
		{&SdkmanDetector{}, ".sdkman/candidates/java/21.0.1/bin/java", "Java 21.0.1"},
	} {
		t.Run(string(tc.detector.Name()), func(t *testing.T) {
			_, version, _, _ := tc.detector.Detect(filepath.Join(HomeDir(), tc.path))
			if version != tc.want {
				t.Fatalf("want %q, got %q", tc.want, version)
			}
		})
	}
}

func TestSdkmanCurrentResolvesVersion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := filepath.Join(HomeDir(), ".sdkman/candidates/java")
	target := filepath.Join(dir, "21.0.1/bin/java")
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("java"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("21.0.1", filepath.Join(dir, "current")); err != nil {
		t.Fatal(err)
	}
	_, version, _, _ := (&SdkmanDetector{}).Detect(filepath.Join(dir, "current/bin/java"))
	if version != "Java 21.0.1" {
		t.Fatalf("current version = %q", version)
	}
}
func TestMemoizedCommandCoalescesConcurrentCalls(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint("failure=", fail), func(t *testing.T) {
			s := &queryTestSystem{fail: fail}
			SetSystem(s)
			defer SetSystem(nil)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					out, err := memoizedRunCmd("audit", "shared")
					if (err != nil) != fail || (!fail && out != "ok") {
						t.Errorf("query result = %q, %v", out, err)
					}
				}()
			}
			close(start)
			wg.Wait()
			if n := s.calls.Load(); n != 1 {
				t.Fatalf("same cold query executed %d times, want 1", n)
			}
		})
	}
}

func TestRenamedAppImageHeader(t *testing.T) {
	for _, kind := range []byte{0, 1, 2, 3} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "renamed-app")
			header := []byte{'\x7f', 'E', 'L', 'F', 0, 0, 0, 0, 'A', 'I', kind}
			if err := os.WriteFile(path, header, 0700); err != nil {
				t.Fatal(err)
			}
			r := (&AppImageDetector{}).Detect(path)
			want := kind == 1 || kind == 2
			if (r != nil) != want {
				t.Fatalf("AppImage type %d: got %+v", kind, r)
			}
		})
	}
}
func TestDesktopLauncherDoesNotClaimInterpreter(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".local/share/applications")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ordinary.desktop"), []byte("[Desktop Entry]\nName=Ordinary Python App\nExec=/usr/bin/python3 /opt/app/main.py\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r := (&AppImageDetector{}).Detect("/usr/bin/python3")
	if r != nil {
		t.Fatalf("ordinary Python interpreter claimed as %s; remove=%q", r.Manager, r.RemoveCmd)
	}
}
