package audio

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// Recorder is a thread-safe WAV recorder used by the ACE-Step audio path. The
// TUI thread calls ToggleStart / ToggleStop to start or stop recording; the
// audio goroutine (running inside the streamer) calls Write on every chunk
// pushed to the speaker. Write is a no-op when recording is inactive, so it
// is safe to leave the tap installed and toggle on/off at runtime.
//
// SF2 has its own per-Root WAV writer wired inside audio.Root.Stream (it ran
// before ACE-Step recording existed); this Recorder mirrors that capability
// for the ACE-Step engine so the press-r UX is engine-agnostic.
type Recorder struct {
	mu         sync.Mutex
	wav        *WAVWriter
	path       string
	sampleRate int
}

// NewRecorder constructs an inactive Recorder. sampleRate is the WAV header
// rate written on ToggleStart; the audio thread must push frames at this
// rate (mismatches will pitch the WAV up or down on playback).
func NewRecorder(sampleRate int) *Recorder {
	return &Recorder{sampleRate: sampleRate}
}

// Active reports whether recording is currently in flight, and the path the
// active WAV is being written to.
func (r *Recorder) Active() (bool, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.wav != nil, r.path
}

// ToggleStart begins recording. tag is mixed into the filename so the user
// can tell SF2 and ACE-Step captures apart. Returns the absolute path of the
// new WAV. Errors if recording is already in progress.
func (r *Recorder) ToggleStart(tag string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.wav != nil {
		return "", errors.New("already recording")
	}
	if tag == "" {
		tag = "rec"
	}
	// UnixNano (vs Unix) so rapid toggle cycles never collide on the same
	// second. The number isn't user-facing — chronological order is what
	// matters and that's preserved.
	name := fmt.Sprintf("termus-%s-%d.wav", tag, time.Now().UnixNano())
	path, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	w, err := NewWAVWriter(path, r.sampleRate, 2)
	if err != nil {
		return "", err
	}
	r.wav = w
	r.path = path
	return path, nil
}

// ToggleStop ends recording and finalises the WAV file. No-op when not
// recording. Returns the close error (which can leave a partial WAV on disk).
func (r *Recorder) ToggleStop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.wav == nil {
		return nil
	}
	err := r.wav.Close()
	r.wav = nil
	r.path = ""
	return err
}

// Write appends one chunk of stereo frames to the active WAV, or drops the
// write when recording is inactive. Called from the audio thread on every
// streamer chunk; the mutex is uncontended outside of toggle events.
func (r *Recorder) Write(frames [][2]float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.wav == nil {
		return
	}
	if err := r.wav.Write(frames); err != nil {
		// Mirror audio.Root.Stream's policy: on write error, close the file
		// silently and stop recording. The TUI will not learn about this
		// (the toggle path returns success) — same trade-off as SF2.
		_ = r.wav.Close()
		r.wav = nil
		r.path = ""
	}
}
