package acestep

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strconv"
	"sync"
	"testing"
	"time"
)

// fakeSpawner records spawn calls without forking.
type fakeSpawner struct {
	called int
	err    error
}

func (f *fakeSpawner) Start(ctx context.Context, dir, py string, args []string, logger io.Writer) (*exec.Cmd, error) {
	f.called++
	if f.err != nil {
		return nil, f.err
	}
	// Return a placeholder exec.Cmd with nil Process so Shutdown is a no-op.
	return &exec.Cmd{}, nil
}

// fakeProbe returns a fixed sequence of health responses, one per call.
type fakeProbe struct {
	mu   sync.Mutex
	idx  int
	rets []HealthResponse
	errs []error
}

func (f *fakeProbe) probe(ctx context.Context, url string) (HealthResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.idx >= len(f.rets) {
		f.idx = len(f.rets) - 1
	}
	h := f.rets[f.idx]
	var err error
	if f.idx < len(f.errs) {
		err = f.errs[f.idx]
	}
	f.idx++
	return h, err
}

// stubInstaller is an Installer wired to always-installed mock state.
func stubInstaller(t *testing.T, installed bool) *Installer {
	t.Helper()
	paths := map[string]bool{}
	if installed {
		paths["/svc/venv/bin/python"] = true
		paths["/svc/vendor/ace-step"] = true
	}
	return &Installer{
		ServiceDir: "/svc",
		runner:     &mockRunner{lookups: map[string]string{"uv": "/usr/local/bin/uv"}},
		fs:         mockFS{paths: paths},
	}
}

func TestEnsureReady_HappyPath_SkipsInstallWhenPresent(t *testing.T) {
	spawner := &fakeSpawner{}
	probe := &fakeProbe{rets: []HealthResponse{
		{Loaded: true, Backend: "mlx", ModelName: "acestep-v15-turbo"},
	}}
	m := &Manager{
		Installer:  stubInstaller(t, true),
		Port:       17790,
		spawner:    spawner,
		probeFn:    probe.probe,
		timeNowFn:  time.Now,
		sleepFn:    func(time.Duration) {},
		clientFn:   NewClient,
		maxWaitSec: 5,
	}
	ch := make(chan StatusEvent, 16)
	client, err := m.EnsureReady(context.Background(), ch)
	if err != nil {
		t.Fatalf("EnsureReady: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if spawner.called != 1 {
		t.Errorf("spawner called %d times, want 1", spawner.called)
	}
	close(ch)
	var phases []string
	for ev := range ch {
		phases = append(phases, ev.Phase)
	}
	// Should NOT include "installing".
	for _, p := range phases {
		if p == "installing" {
			t.Errorf("unexpected installing phase when already installed: %v", phases)
		}
	}
	// Should include ready.
	if phases[len(phases)-1] != "ready" {
		t.Errorf("last phase = %q, want ready", phases[len(phases)-1])
	}
}

func TestEnsureReady_WaitsForLoaded(t *testing.T) {
	spawner := &fakeSpawner{}
	probe := &fakeProbe{
		rets: []HealthResponse{
			{Loaded: false, Backend: "mlx"},
			{Loaded: false, Backend: "mlx"},
			{Loaded: true, Backend: "mlx"},
		},
	}
	m := &Manager{
		Installer:  stubInstaller(t, true),
		Port:       17791,
		spawner:    spawner,
		probeFn:    probe.probe,
		timeNowFn:  time.Now,
		sleepFn:    func(time.Duration) {}, // fast tests
		maxWaitSec: 30,
	}
	client, err := m.EnsureReady(context.Background(), nil)
	if err != nil {
		t.Fatalf("EnsureReady: %v", err)
	}
	if client == nil {
		t.Fatal("nil client")
	}
}

func TestEnsureReady_TimesOutIfNeverLoaded(t *testing.T) {
	spawner := &fakeSpawner{}
	probe := &fakeProbe{rets: []HealthResponse{
		{Loaded: false, Backend: "mlx"},
	}}
	// Force timeNowFn to advance fast so we don't wait real seconds.
	start := time.Now()
	calls := 0
	timeFn := func() time.Time {
		calls++
		return start.Add(time.Duration(calls) * time.Second)
	}
	m := &Manager{
		Installer:  stubInstaller(t, true),
		Port:       17792,
		spawner:    spawner,
		probeFn:    probe.probe,
		timeNowFn:  timeFn,
		sleepFn:    func(time.Duration) {},
		maxWaitSec: 3,
	}
	_, err := m.EnsureReady(context.Background(), nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestEnsureReady_UsesExistingDaemon(t *testing.T) {
	// Spin up a httptest server that exposes /health -> Loaded:true.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Loaded: true, Backend: "mlx", ModelName: "test",
		})
	}))
	defer srv.Close()

	// Parse the port from the httptest URL so we point the Manager at it.
	u, _ := url.Parse(srv.URL)
	host, portStr, _ := net.SplitHostPort(u.Host)
	if host == "127.0.0.1" {
		host = "localhost"
	}
	port, _ := strconv.Atoi(portStr)

	spawner := &fakeSpawner{}
	m := &Manager{
		Installer:  stubInstaller(t, true),
		Port:       port,
		spawner:    spawner,
		timeNowFn:  time.Now,
		sleepFn:    func(time.Duration) {},
		maxWaitSec: 5,
	}
	client, err := m.EnsureReady(context.Background(), nil)
	if err != nil {
		t.Fatalf("EnsureReady: %v", err)
	}
	if client == nil {
		t.Fatal("nil client")
	}
	if spawner.called != 0 {
		t.Errorf("expected spawner NOT called when daemon already present; got %d", spawner.called)
	}
}

func TestEnsureReady_FailsOnInstallError(t *testing.T) {
	// Installer reports not-installed and fails when EnsureInstalled runs
	// (uv lookup fails AND the curl step fails too).
	inst := &Installer{
		ServiceDir: "/svc",
		runner: &mockRunner{
			lookups: map[string]string{},
			failOn: map[string]error{
				"sh -c curl -LsSf https://astral.sh/uv/install.sh | sh -s -- --no-modify-path": context.DeadlineExceeded,
			},
		},
		fs: mockFS{paths: map[string]bool{}},
	}
	m := &Manager{
		Installer:  inst,
		Port:       17793,
		spawner:    &fakeSpawner{},
		probeFn:    func(context.Context, string) (HealthResponse, error) { return HealthResponse{Loaded: true}, nil },
		timeNowFn:  time.Now,
		sleepFn:    func(time.Duration) {},
		maxWaitSec: 1,
	}
	_, err := m.EnsureReady(context.Background(), nil)
	if err == nil {
		t.Fatal("expected install error to propagate")
	}
}

func TestShutdown_IsNoOpWhenNothingRunning(t *testing.T) {
	m := &Manager{Installer: stubInstaller(t, true)}
	if err := m.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown on idle manager returned err: %v", err)
	}
}

func TestAvailablePort_FindsOpenPort(t *testing.T) {
	p, err := AvailablePort(28000, 200)
	if err != nil {
		t.Fatalf("AvailablePort: %v", err)
	}
	if p < 28000 || p >= 28200 {
		t.Errorf("port %d outside expected range", p)
	}
}
