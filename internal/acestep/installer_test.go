package acestep

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// mockRunner records every Run call so tests can assert sequencing.
type mockRunner struct {
	mu        sync.Mutex
	calls     []mockCall
	lookups   map[string]string // name -> path; missing entries → lookup fails
	failOn    map[string]error  // index name → error to return
	stderrOut map[string][]string
}

type mockCall struct {
	Name string
	Args []string
	Dir  string
}

func (m *mockRunner) Lookup(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.lookups[name]; ok {
		return p, nil
	}
	return "", errors.New("not found")
}

func (m *mockRunner) Run(ctx context.Context, dir string, env []string, name string, args []string, onStderr func(string)) error {
	m.mu.Lock()
	m.calls = append(m.calls, mockCall{Name: name, Args: append([]string(nil), args...), Dir: dir})
	key := name + " " + strings.Join(args, " ")
	if lines, ok := m.stderrOut[key]; ok {
		for _, l := range lines {
			if onStderr != nil {
				onStderr(l)
			}
		}
	}
	failErr := m.failOn[key]
	m.mu.Unlock()
	return failErr
}

// mockFS lets tests pretend specific paths exist.
type mockFS struct {
	paths map[string]bool
}

func (m mockFS) Exists(path string) bool                       { return m.paths[path] }
func (m mockFS) MkdirAll(path string, perm os.FileMode) error  { m.paths[path] = true; return nil }

func TestIsInstalled_FalseWhenVenvMissing(t *testing.T) {
	i := &Installer{
		ServiceDir: "/svc",
		fs:         mockFS{paths: map[string]bool{}},
		runner:     &mockRunner{},
	}
	if i.IsInstalled() {
		t.Fatal("expected not installed when venv python is absent")
	}
}

func TestIsInstalled_FalseWhenVendorMissing(t *testing.T) {
	i := &Installer{
		ServiceDir: "/svc",
		fs: mockFS{paths: map[string]bool{
			filepath.Join("/svc", "venv", "bin", "python"): true,
		}},
		runner: &mockRunner{},
	}
	if i.IsInstalled() {
		t.Fatal("expected not installed when vendor clone is absent")
	}
}

func TestIsInstalled_TrueWhenBothPresent(t *testing.T) {
	i := &Installer{
		ServiceDir: "/svc",
		fs: mockFS{paths: map[string]bool{
			filepath.Join("/svc", "venv", "bin", "python"):     true,
			filepath.Join("/svc", "vendor", "ace-step"):        true,
			filepath.Join("/svc", "vendor", "ace-step", ".git"): true,
		}},
		runner: &mockRunner{},
	}
	if !i.IsInstalled() {
		t.Fatal("expected installed when venv + vendor exist")
	}
}

func TestEnsureInstalled_FullFreshInstall_CallsAllSteps(t *testing.T) {
	// Fresh install: nothing exists, uv install must run, venv must be created,
	// all steps must fire in order.
	home, err := homeDir()
	if err != nil {
		t.Skipf("no home dir: %v", err)
	}
	uvPathAfterInstall := filepath.Join(home, ".local", "bin", "uv")
	runner := &mockRunner{
		lookups: map[string]string{}, // uv NOT in PATH initially
	}
	fs := mockFS{paths: map[string]bool{}}
	events := make(chan InstallEvent, 256)
	i := &Installer{
		ServiceDir: "/svc",
		Events:     events,
		runner:     runner,
		fs:         fs,
	}
	// After uv install, the uv binary appears at $HOME/.local/bin/uv.
	// We achieve that by mutating fs.paths between the curl|sh call and
	// the existence check, but our mock fires after Run returns and
	// before the existence check. Simplest: just pre-mark the path as
	// present and let the curl step proceed.
	fs.paths[uvPathAfterInstall] = true
	i.fs = fs

	if err := i.EnsureInstalled(context.Background()); err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	close(events)

	// Collect call names.
	got := make([]string, 0, len(runner.calls))
	for _, c := range runner.calls {
		got = append(got, c.Name+" "+strings.Join(c.Args, " "))
	}
	// We expect: sh -c curl|sh, uvPath python install 3.11, uvPath venv ...,
	// uvPath pip install -r requirements.txt, git clone, uvPath pip install -e,
	// venv python -c model-download.
	wantPrefixes := []string{
		"sh -c curl",
		uvPathAfterInstall + " python install 3.11",
		uvPathAfterInstall + " venv",
		uvPathAfterInstall + " pip install --python",
		"git clone",
		uvPathAfterInstall + " pip install --python " + filepath.Join("/svc", "venv", "bin", "python") + " -e",
		filepath.Join("/svc", "venv", "bin", "python") + " -c",
	}
	if len(got) != len(wantPrefixes) {
		t.Fatalf("call count: got %d want %d\n  got: %v\n  want prefixes: %v", len(got), len(wantPrefixes), got, wantPrefixes)
	}
	for i, prefix := range wantPrefixes {
		if !strings.HasPrefix(got[i], prefix) {
			t.Errorf("call[%d]: %q does not start with %q", i, got[i], prefix)
		}
	}

	// Sanity: a "ready" event fired.
	sawReady := false
	for ev := range events {
		if ev.Phase == "ready" {
			sawReady = true
		}
	}
	if !sawReady {
		t.Error("no 'ready' event emitted")
	}
}

func TestEnsureInstalled_SkipsVenvWhenPresent(t *testing.T) {
	runner := &mockRunner{
		lookups: map[string]string{"uv": "/usr/local/bin/uv"},
	}
	fs := mockFS{paths: map[string]bool{
		filepath.Join("/svc", "venv", "bin", "python"): true,
		// vendor missing so the clone step still runs.
	}}
	i := &Installer{
		ServiceDir: "/svc",
		runner:     runner,
		fs:         fs,
	}
	if err := i.EnsureInstalled(context.Background()); err != nil {
		t.Fatalf("EnsureInstalled: %v", err)
	}
	// Verify NO "uv venv" call was made.
	for _, c := range runner.calls {
		if len(c.Args) > 0 && c.Args[0] == "venv" {
			t.Fatalf("unexpected venv create call: %v", c.Args)
		}
	}
}

func TestEnsureInstalled_PropagatesStepError(t *testing.T) {
	runner := &mockRunner{
		lookups: map[string]string{"uv": "/usr/local/bin/uv"},
		failOn: map[string]error{
			"/usr/local/bin/uv python install 3.11": errors.New("network down"),
		},
	}
	fs := mockFS{paths: map[string]bool{}}
	events := make(chan InstallEvent, 64)
	i := &Installer{
		ServiceDir: "/svc",
		Events:     events,
		runner:     runner,
		fs:         fs,
	}
	err := i.EnsureInstalled(context.Background())
	if err == nil {
		t.Fatal("expected error from python step")
	}
	if !strings.Contains(err.Error(), "network down") {
		t.Errorf("error chain missing inner cause: %v", err)
	}
}

func TestEstimatedSize_ReportsRoughBreakdown(t *testing.T) {
	i := &Installer{ServiceDir: "/svc"}
	sz := i.EstimatedSize()
	if sz.Total() < 10_000_000_000 {
		t.Errorf("expected >10GB total, got %d", sz.Total())
	}
	if sz.Model < sz.Deps {
		t.Errorf("expected model > deps: %d vs %d", sz.Model, sz.Deps)
	}
}

func homeDir() (string, error) {
	return os.UserHomeDir()
}
