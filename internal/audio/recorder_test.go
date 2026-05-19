package audio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRecorderWriteIsNoopWhenIdle verifies the audio-thread Write is safe
// to call when no recording is active (the streamer leaves the tap wired
// across the whole session, so Write is invoked on every chunk regardless
// of toggle state).
func TestRecorderWriteIsNoopWhenIdle(t *testing.T) {
	r := NewRecorder(48000)
	active, _ := r.Active()
	if active {
		t.Fatal("Recorder should start inactive")
	}
	// Should not panic / error.
	r.Write([][2]float64{{0.1, -0.1}, {0.2, -0.2}})
}

// TestRecorderRoundtrip exercises a start → write → stop → start → stop
// cycle and verifies that each toggle produces a non-empty WAV at a fresh
// path.
func TestRecorderRoundtrip(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	r := NewRecorder(48000)

	path1, err := r.ToggleStart("ace")
	if err != nil {
		t.Fatalf("ToggleStart: %v", err)
	}
	// macOS resolves /tmp -> /private/tmp; resolve both ends before
	// comparing so the assertion isn't platform-fragile.
	gotDir, _ := filepath.EvalSymlinks(filepath.Dir(path1))
	wantDir, _ := filepath.EvalSymlinks(tmp)
	if gotDir != wantDir {
		t.Fatalf("expected WAV under %s, got %s", wantDir, gotDir)
	}
	if !strings.Contains(filepath.Base(path1), "ace") {
		t.Fatalf("expected tag in filename, got %s", path1)
	}

	r.Write([][2]float64{{0.5, -0.5}, {0.4, -0.4}, {0.3, -0.3}})
	if err := r.ToggleStop(); err != nil {
		t.Fatalf("ToggleStop: %v", err)
	}

	info, err := os.Stat(path1)
	if err != nil {
		t.Fatalf("stat WAV: %v", err)
	}
	if info.Size() < 44 { // WAV header is 44 bytes
		t.Fatalf("WAV too small: %d bytes", info.Size())
	}

	// Second toggle cycle must produce a fresh path.
	path2, err := r.ToggleStart("ace")
	if err != nil {
		t.Fatalf("second ToggleStart: %v", err)
	}
	if path2 == path1 {
		t.Fatalf("expected fresh path on second start, got duplicate %s", path1)
	}
	_ = r.ToggleStop()
}

// TestRecorderDoubleStartFails — starting while already active must error
// rather than silently leak the old WAV handle.
func TestRecorderDoubleStartFails(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	r := NewRecorder(48000)
	if _, err := r.ToggleStart("ace"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ToggleStart("ace"); err == nil {
		t.Fatal("expected error on second start while active")
	}
	_ = r.ToggleStop()
}
