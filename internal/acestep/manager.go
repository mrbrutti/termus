package acestep

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// StatusEvent reports daemon lifecycle progress. Streamed to the TUI/CLI
// during EnsureReady.
type StatusEvent struct {
	Phase   string // "checking-install" | "installing" | "starting-daemon" | "loading-model" | "ready"
	Message string
	Err     error
}

// Manager owns the ACE-Step daemon subprocess. Lifecycle:
//
//	mgr := &Manager{Installer: inst, Port: 7790, Logger: os.Stderr}
//	client, err := mgr.EnsureReady(ctx, statusCh)
//	... use client ...
//	mgr.Shutdown(ctx)
//
// Concurrency: EnsureReady is single-call (guarded by the mutex). Shutdown
// is idempotent.
type Manager struct {
	Installer *Installer
	Port      int
	Logger    io.Writer

	// OnProgress, if non-nil, is called with parsed RENDER_PROGRESS lines
	// from the daemon's stderr (see services/acestep/server.py). The
	// callback may fire from a goroutine; implementations should be
	// non-blocking. Set this before EnsureReady is called so the spawned
	// daemon's stream is captured from the first byte.
	OnProgress ProgressFunc

	// For tests, allow swapping daemon-spawning behavior.
	spawner    daemonSpawner
	probeFn    func(ctx context.Context, url string) (HealthResponse, error)
	timeNowFn  func() time.Time
	sleepFn    func(d time.Duration)
	healthURL  string // computed from Port; tests can override via newManagerForTest
	clientFn   func(baseURL string, timeout time.Duration) *Client
	maxWaitSec int // health-poll cap

	mu      sync.Mutex
	cmd     *exec.Cmd
	client  *Client
	running bool
}

// daemonSpawner is the seam used by tests to avoid actually exec-ing the
// Python server. Start returns a cmd-like handle and nil if launch succeeded.
type daemonSpawner interface {
	Start(ctx context.Context, dir string, py string, args []string, logger io.Writer) (*exec.Cmd, error)
}

type realSpawner struct{}

func (realSpawner) Start(ctx context.Context, dir string, py string, args []string, logger io.Writer) (*exec.Cmd, error) {
	cmd := exec.Command(py, args...)
	cmd.Dir = dir
	// Inherit our process group so when termus exits, the child does too.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: false}
	cmd.Stdout = logger
	cmd.Stderr = logger
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start daemon: %w", err)
	}
	return cmd, nil
}

// defaultPort is the conventional ACE-Step port. Mirrors services/acestep/server.py.
const defaultPort = 7790

func (m *Manager) port() int {
	if m.Port > 0 {
		return m.Port
	}
	return defaultPort
}

func (m *Manager) baseURL() string {
	return fmt.Sprintf("http://localhost:%d", m.port())
}

func (m *Manager) logger() io.Writer {
	if m.Logger == nil {
		return io.Discard
	}
	return m.Logger
}

// EnsureReady ensures the daemon is up and the model is loaded, installing
// the bootstrap toolchain first if necessary. Streams status events to ch
// (may be nil). Returns a Client ready for /render calls.
func (m *Manager) EnsureReady(ctx context.Context, ch chan<- StatusEvent) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Defaults for the seams.
	if m.spawner == nil {
		m.spawner = realSpawner{}
	}
	if m.probeFn == nil {
		m.probeFn = m.defaultProbe
	}
	if m.clientFn == nil {
		m.clientFn = NewClient
	}
	if m.timeNowFn == nil {
		m.timeNowFn = time.Now
	}
	if m.sleepFn == nil {
		m.sleepFn = time.Sleep
	}
	if m.maxWaitSec == 0 {
		m.maxWaitSec = 180 // 3-minute cap on model warmup
	}

	emit := func(ev StatusEvent) {
		if ch == nil {
			return
		}
		select {
		case ch <- ev:
		default:
		}
	}

	// 1. Install check.
	emit(StatusEvent{Phase: "checking-install", Message: "verifying ACE-Step install"})
	if m.Installer == nil {
		return nil, errors.New("manager: nil Installer")
	}
	if !m.Installer.IsInstalled() {
		emit(StatusEvent{Phase: "installing", Message: "bootstrapping toolchain"})
		if err := m.Installer.EnsureInstalled(ctx); err != nil {
			emit(StatusEvent{Phase: "installing", Err: err})
			return nil, fmt.Errorf("install: %w", err)
		}
	}

	// 2. Reuse existing daemon if the port is already open + healthy.
	if existing := m.probeExisting(ctx); existing != nil {
		emit(StatusEvent{Phase: "ready", Message: "found existing daemon on port " + strconv.Itoa(m.port())})
		m.client = existing
		return existing, nil
	}

	// 3. Spawn the daemon.
	emit(StatusEvent{Phase: "starting-daemon", Message: "launching ACE-Step daemon"})
	py := filepath.Join(m.Installer.ServiceDir, "venv", "bin", "python")
	serverPy := filepath.Join(m.Installer.ServiceDir, "server.py")
	// Tee the daemon's stderr through a parser that picks up
	// RENDER_PROGRESS markers (see services/acestep/server.py). If
	// OnProgress is nil, progressTee returns a plain passthrough.
	logger := progressTee(m.logger(), m.OnProgress)
	cmd, err := m.spawner.Start(ctx, m.Installer.ServiceDir, py, []string{serverPy}, logger)
	if err != nil {
		emit(StatusEvent{Phase: "starting-daemon", Err: err})
		return nil, fmt.Errorf("spawn daemon: %w", err)
	}
	m.cmd = cmd
	m.running = true

	// 4. Poll /health until loaded=true.
	emit(StatusEvent{Phase: "loading-model", Message: "loading model into memory"})
	client, err := m.waitForHealthy(ctx, emit)
	if err != nil {
		// Tear down the half-started daemon so the user can retry.
		_ = m.shutdownLocked(context.Background())
		return nil, err
	}
	m.client = client
	emit(StatusEvent{Phase: "ready", Message: "ACE-Step ready"})
	return client, nil
}

// probeExisting returns a Client if the configured port is already serving a
// loaded ACE-Step daemon. Returns nil otherwise (including when the port is
// closed, the response doesn't parse, or loaded=false).
func (m *Manager) probeExisting(ctx context.Context) *Client {
	// Quick TCP check first to avoid a multi-second HTTP timeout on a closed port.
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", m.port()), 500*time.Millisecond)
	if err != nil {
		return nil
	}
	_ = conn.Close()
	hCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	h, err := m.probeFn(hCtx, m.baseURL())
	if err != nil || !h.Loaded {
		return nil
	}
	return m.clientFn(m.baseURL(), 5*time.Minute)
}

// waitForHealthy polls /health every second until loaded=true, the context
// is cancelled, or maxWaitSec elapses.
func (m *Manager) waitForHealthy(ctx context.Context, emit func(StatusEvent)) (*Client, error) {
	deadline := m.timeNowFn().Add(time.Duration(m.maxWaitSec) * time.Second)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if m.timeNowFn().After(deadline) {
			return nil, fmt.Errorf("daemon did not become healthy within %ds", m.maxWaitSec)
		}
		hCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		h, err := m.probeFn(hCtx, m.baseURL())
		cancel()
		if err != nil {
			// Daemon not up yet; keep polling.
			m.sleepFn(1 * time.Second)
			continue
		}
		if h.Loaded {
			return m.clientFn(m.baseURL(), 5*time.Minute), nil
		}
		// Surface warmup state to the UI.
		if h.Error != "" {
			emit(StatusEvent{Phase: "loading-model", Message: "warmup error: " + h.Error})
			return nil, fmt.Errorf("daemon reported warmup error: %s", h.Error)
		}
		emit(StatusEvent{Phase: "loading-model", Message: fmt.Sprintf("loading (backend=%s)", h.Backend)})
		m.sleepFn(1 * time.Second)
	}
}

// defaultProbe issues a real /health request.
func (m *Manager) defaultProbe(ctx context.Context, url string) (HealthResponse, error) {
	c := NewClient(url, 5*time.Second)
	return c.Health(ctx)
}

// Shutdown sends SIGTERM, waits up to 5s, then SIGKILLs. Idempotent.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.shutdownLocked(ctx)
}

func (m *Manager) shutdownLocked(_ context.Context) error {
	if !m.running || m.cmd == nil || m.cmd.Process == nil {
		return nil
	}
	pid := m.cmd.Process.Pid
	_ = m.cmd.Process.Signal(syscall.SIGTERM)
	done := make(chan error, 1)
	go func() { done <- m.cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		fmt.Fprintf(m.logger(), "acestep: SIGTERM timed out for pid %d, sending SIGKILL\n", pid)
		_ = m.cmd.Process.Kill()
		<-done
	}
	m.running = false
	m.cmd = nil
	return nil
}

// IsRunning reports whether the manager has a live daemon subprocess.
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// ensureNotRunning is a small helper exposed for tests that want to reset
// state between scenarios.
func (m *Manager) ensureNotRunning() {
	m.mu.Lock()
	m.running = false
	m.cmd = nil
	m.mu.Unlock()
}

// AvailablePort returns a port in [start, start+span) that is currently
// unbound. Useful when callers want to avoid the default 7790 collision.
func AvailablePort(start, span int) (int, error) {
	for p := start; p < start+span; p++ {
		l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", p))
		if err != nil {
			continue
		}
		_ = l.Close()
		return p, nil
	}
	return 0, fmt.Errorf("no available port in [%d, %d)", start, start+span)
}

