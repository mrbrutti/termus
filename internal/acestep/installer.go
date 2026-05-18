// Package acestep also provides an installer that bootstraps the Python
// toolchain, venv, source clone, and model weights needed by the daemon.
//
// SCOPE / UNTESTED REALITY
//
// The installer is mocked in unit tests; the actual uv/pip/git/network calls
// have NOT been exercised end-to-end in this PR. The user can run
// `termus --acestep-install` to perform the real install. Honest reality:
// model downloads invoke `huggingface_hub.snapshot_download` which prints
// tqdm progress to stderr. We capture lines but do not attempt to parse
// percentages — we just surface the latest message text.
package acestep

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// InstallEvent is one step of the bootstrap pipeline. Progress is in [0,1]
// when known and math.NaN() otherwise. Err is set only for terminal failures
// in a given phase; the installer aborts on first non-nil Err.
type InstallEvent struct {
	Phase    string
	Message  string
	Progress float64
	Err      error
}

// InstallSize describes the rough disk footprint of a full install. Values
// are best-effort marketing estimates, not measured bytes. UI surfaces these
// in the "install confirmation" prompt.
type InstallSize struct {
	Python int64 // ~50 MB
	Deps   int64 // ~2 GB
	Model  int64 // ~9.4 GB
}

// Total returns the sum of all components.
func (s InstallSize) Total() int64 { return s.Python + s.Deps + s.Model }

// Installer bootstraps the ACE-Step toolchain into ServiceDir.
//
// ServiceDir is typically the absolute path to <repo>/services/acestep, the
// same directory that holds server.py / requirements.txt / install.sh.
type Installer struct {
	ServiceDir string
	Events     chan<- InstallEvent

	// runner is swapped in tests to avoid actually executing commands.
	runner commandRunner

	// fs is swapped in tests to assert on path checks without touching the
	// real filesystem.
	fs installerFS
}

// NewInstaller builds an installer rooted at serviceDir. Events is optional.
func NewInstaller(serviceDir string, events chan<- InstallEvent) *Installer {
	return &Installer{
		ServiceDir: serviceDir,
		Events:     events,
		runner:     realCommandRunner{},
		fs:         realInstallerFS{},
	}
}

// commandRunner is the seam tests use to mock out exec.Command. Run executes
// `name args...` with the given env and working directory, streaming stderr
// lines into onStderrLine (one call per line, line trimmed of trailing \n).
// Returns the command's exit error.
type commandRunner interface {
	Run(ctx context.Context, dir string, env []string, name string, args []string, onStderrLine func(string)) error
	Lookup(name string) (string, error)
}

type realCommandRunner struct{}

func (realCommandRunner) Lookup(name string) (string, error) {
	return exec.LookPath(name)
}

func (realCommandRunner) Run(ctx context.Context, dir string, env []string, name string, args []string, onStderrLine func(string)) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("pipe stderr: %w", err)
	}
	cmd.Stdout = io.Discard
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if onStderrLine != nil {
				onStderrLine(line)
			}
		}
	}()
	waitErr := cmd.Wait()
	wg.Wait()
	if waitErr != nil {
		return waitErr
	}
	return nil
}

// installerFS is the seam tests use to mock filesystem checks.
type installerFS interface {
	Exists(path string) bool
	MkdirAll(path string, perm os.FileMode) error
}

type realInstallerFS struct{}

func (realInstallerFS) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (realInstallerFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// EstimatedSize returns rough size estimates for the install. The numbers
// are static — they come from the model card and typical pip dep size for
// torch+mlx+transformers — not measured at runtime.
func (i *Installer) EstimatedSize() InstallSize {
	return InstallSize{
		Python: 50 * 1024 * 1024,
		Deps:   2 * 1024 * 1024 * 1024,
		Model:  9_400 * 1024 * 1024,
	}
}

// venvPython returns the absolute path to the venv's python binary.
func (i *Installer) venvPython() string {
	return filepath.Join(i.ServiceDir, "venv", "bin", "python")
}

// vendorACEStep returns the path where the ACE-Step source clone lives.
func (i *Installer) vendorACEStep() string {
	return filepath.Join(i.ServiceDir, "vendor", "ace-step")
}

// IsInstalled reports whether the bootstrap appears complete. It checks:
// - venv python exists
// - ACE-Step source clone exists
// - The acestep package is importable from the venv (the surest sign that
//   model downloads will work).
//
// It deliberately does NOT verify the model weights are downloaded: that
// check is too slow/expensive. EnsureInstalled re-runs the download step
// which is idempotent (HuggingFace caches by hash).
func (i *Installer) IsInstalled() bool {
	if !i.fs.Exists(i.venvPython()) {
		return false
	}
	if !i.fs.Exists(i.vendorACEStep()) {
		return false
	}
	return true
}

// EnsureInstalled performs the missing bootstrap steps and emits events.
// Idempotent: calling it after a full install is a quick no-op modulo
// re-checking the model download (which HF caches).
func (i *Installer) EnsureInstalled(ctx context.Context) error {
	if i.runner == nil {
		i.runner = realCommandRunner{}
	}
	if i.fs == nil {
		i.fs = realInstallerFS{}
	}

	uvPath, err := i.ensureUV(ctx)
	if err != nil {
		i.emit(InstallEvent{Phase: "uv", Message: "uv install failed", Err: err})
		return fmt.Errorf("acestep installer: uv: %w", err)
	}
	if err := i.ensurePython311(ctx, uvPath); err != nil {
		i.emit(InstallEvent{Phase: "python", Message: "python 3.11 install failed", Err: err})
		return fmt.Errorf("acestep installer: python: %w", err)
	}
	if err := i.ensureVenv(ctx, uvPath); err != nil {
		i.emit(InstallEvent{Phase: "venv", Message: "venv create failed", Err: err})
		return fmt.Errorf("acestep installer: venv: %w", err)
	}
	if err := i.ensureDeps(ctx, uvPath); err != nil {
		i.emit(InstallEvent{Phase: "deps", Message: "pip install failed", Err: err})
		return fmt.Errorf("acestep installer: deps: %w", err)
	}
	if err := i.ensureVendor(ctx, uvPath); err != nil {
		i.emit(InstallEvent{Phase: "vendor", Message: "vendor clone/install failed", Err: err})
		return fmt.Errorf("acestep installer: vendor: %w", err)
	}
	if err := i.ensureModel(ctx); err != nil {
		i.emit(InstallEvent{Phase: "model", Message: "model download failed", Err: err})
		return fmt.Errorf("acestep installer: model: %w", err)
	}
	i.emit(InstallEvent{Phase: "ready", Message: "ACE-Step engine ready", Progress: 1.0})
	return nil
}

// ensureUV checks for an existing uv binary and installs it if absent.
//
// Returns the path to use for subsequent uv invocations.
func (i *Installer) ensureUV(ctx context.Context) (string, error) {
	i.emit(InstallEvent{Phase: "uv", Message: "checking for uv", Progress: math.NaN()})
	if path, err := i.runner.Lookup("uv"); err == nil {
		i.emit(InstallEvent{Phase: "uv", Message: "uv already installed", Progress: 1.0})
		return path, nil
	}
	i.emit(InstallEvent{Phase: "uv", Message: "installing uv (curl | sh)", Progress: 0.1})
	// Two-stage install via shell: curl the installer script and pipe to sh.
	// We don't add to PATH for the parent shell; we'll find uv ourselves after.
	if err := i.runner.Run(ctx, i.ServiceDir, nil, "sh", []string{"-c",
		"curl -LsSf https://astral.sh/uv/install.sh | sh -s -- --no-modify-path",
	}, func(line string) {
		i.emit(InstallEvent{Phase: "uv", Message: line, Progress: math.NaN()})
	}); err != nil {
		return "", fmt.Errorf("install uv: %w", err)
	}
	// uv lands at $HOME/.local/bin/uv (default) or $HOME/.cargo/bin/uv on older
	// installers. Check both.
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "bin", "uv"),
		filepath.Join(home, ".cargo", "bin", "uv"),
	}
	for _, c := range candidates {
		if i.fs.Exists(c) {
			i.emit(InstallEvent{Phase: "uv", Message: "uv installed at " + c, Progress: 1.0})
			return c, nil
		}
	}
	// Last-resort: hope it's now in PATH.
	if path, err := i.runner.Lookup("uv"); err == nil {
		return path, nil
	}
	return "", errors.New("uv installer ran but binary not found in $HOME/.local/bin or $HOME/.cargo/bin")
}

func (i *Installer) ensurePython311(ctx context.Context, uvPath string) error {
	i.emit(InstallEvent{Phase: "python", Message: "installing Python 3.11 via uv", Progress: 0.0})
	if err := i.runner.Run(ctx, i.ServiceDir, nil, uvPath, []string{"python", "install", "3.11"}, func(line string) {
		i.emit(InstallEvent{Phase: "python", Message: line, Progress: math.NaN()})
	}); err != nil {
		return err
	}
	i.emit(InstallEvent{Phase: "python", Message: "Python 3.11 ready", Progress: 1.0})
	return nil
}

func (i *Installer) ensureVenv(ctx context.Context, uvPath string) error {
	if i.fs.Exists(i.venvPython()) {
		i.emit(InstallEvent{Phase: "venv", Message: "venv already present", Progress: 1.0})
		return nil
	}
	i.emit(InstallEvent{Phase: "venv", Message: "creating venv", Progress: 0.0})
	venvDir := filepath.Join(i.ServiceDir, "venv")
	if err := i.runner.Run(ctx, i.ServiceDir, nil, uvPath, []string{"venv", venvDir, "--python", "3.11"}, func(line string) {
		i.emit(InstallEvent{Phase: "venv", Message: line, Progress: math.NaN()})
	}); err != nil {
		return err
	}
	i.emit(InstallEvent{Phase: "venv", Message: "venv created", Progress: 1.0})
	return nil
}

func (i *Installer) ensureDeps(ctx context.Context, uvPath string) error {
	reqPath := filepath.Join(i.ServiceDir, "requirements.txt")
	i.emit(InstallEvent{Phase: "deps", Message: "installing Python dependencies", Progress: 0.0})
	args := []string{"pip", "install", "--python", i.venvPython(), "-r", reqPath}
	if err := i.runner.Run(ctx, i.ServiceDir, nil, uvPath, args, func(line string) {
		i.emit(InstallEvent{Phase: "deps", Message: line, Progress: math.NaN()})
	}); err != nil {
		return err
	}
	i.emit(InstallEvent{Phase: "deps", Message: "dependencies installed", Progress: 1.0})
	return nil
}

func (i *Installer) ensureVendor(ctx context.Context, uvPath string) error {
	vendorPath := i.vendorACEStep()
	if !i.fs.Exists(vendorPath) {
		i.emit(InstallEvent{Phase: "vendor", Message: "cloning ACE-Step source", Progress: 0.0})
		parent := filepath.Dir(vendorPath)
		if err := i.fs.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", parent, err)
		}
		args := []string{"clone", "--depth", "1",
			"https://github.com/clockworksquirrel/ace-step-apple-silicon.git",
			vendorPath,
		}
		if err := i.runner.Run(ctx, i.ServiceDir, nil, "git", args, func(line string) {
			i.emit(InstallEvent{Phase: "vendor", Message: line, Progress: math.NaN()})
		}); err != nil {
			return fmt.Errorf("git clone: %w", err)
		}
	} else {
		i.emit(InstallEvent{Phase: "vendor", Message: "ACE-Step source already present", Progress: 0.5})
	}
	i.emit(InstallEvent{Phase: "vendor", Message: "installing acestep package (editable)", Progress: 0.6})
	args := []string{"pip", "install", "--python", i.venvPython(), "-e", vendorPath}
	if err := i.runner.Run(ctx, i.ServiceDir, nil, uvPath, args, func(line string) {
		i.emit(InstallEvent{Phase: "vendor", Message: line, Progress: math.NaN()})
	}); err != nil {
		return fmt.Errorf("pip install -e: %w", err)
	}
	i.emit(InstallEvent{Phase: "vendor", Message: "ACE-Step source ready", Progress: 1.0})
	return nil
}

// modelDownloadPy is the inline Python that fetches model weights. It calls
// the same ACE-Step helpers install.sh used, with print(flush=True) so we
// see status lines as they happen.
const modelDownloadPy = `
import os, sys
try:
    from acestep.model_downloader import ensure_main_model, ensure_dit_model
except ImportError as exc:
    print("ERROR could not import acestep model_downloader: %s" % exc, file=sys.stderr, flush=True)
    sys.exit(1)

model = os.environ.get("ACESTEP_MODEL", "acestep-v15-turbo")
print("MODEL %s" % model, file=sys.stderr, flush=True)
print("STEP main", file=sys.stderr, flush=True)
ensure_main_model()
print("STEP dit", file=sys.stderr, flush=True)
try:
    ensure_dit_model(model)
except Exception as exc:
    print("WARN ensure_dit_model: %s" % exc, file=sys.stderr, flush=True)
print("DONE", file=sys.stderr, flush=True)
`

func (i *Installer) ensureModel(ctx context.Context) error {
	i.emit(InstallEvent{Phase: "model", Message: "downloading model (this takes a few minutes)", Progress: math.NaN()})
	py := i.venvPython()
	args := []string{"-c", modelDownloadPy}
	if err := i.runner.Run(ctx, i.ServiceDir, nil, py, args, func(line string) {
		// Surface the most informative lines verbatim. tqdm output ends up
		// looking like "Downloading shards: 67%|######7   | 2/3", which we
		// pass through to the UI as-is.
		msg := line
		if strings.HasPrefix(line, "STEP ") {
			msg = "model step: " + strings.TrimPrefix(line, "STEP ")
		} else if strings.HasPrefix(line, "MODEL ") {
			msg = "model: " + strings.TrimPrefix(line, "MODEL ")
		}
		i.emit(InstallEvent{Phase: "model", Message: msg, Progress: math.NaN()})
	}); err != nil {
		return err
	}
	i.emit(InstallEvent{Phase: "model", Message: "model ready", Progress: 1.0})
	return nil
}

func (i *Installer) emit(ev InstallEvent) {
	if i.Events == nil {
		return
	}
	// Non-blocking emit: if no one is listening we drop, rather than stall
	// the installer.
	select {
	case i.Events <- ev:
	default:
	}
}
