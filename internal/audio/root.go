package audio

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
)

// Compile-time assertion: Root must satisfy Commander.
var _ Commander = (*Root)(nil)

// Root is a beep.Streamer that pulls samples from a gen.Algorithm, applies
// master volume, tees the output to a scope ring, and optionally to a WAV
// writer. The audio goroutine reads commands via atomics; UI goroutines call
// SetVolume / TogglePause / ToggleRecord.
type Root struct {
	algo  gen.Algorithm
	scope *scope.Ring

	// hot-path atomics
	volume atomic.Uint32 // 0..100
	paused atomic.Bool

	// command channel for record start/stop (non-hot-path)
	recCmd chan recordCmd

	// audio-thread-owned state (do not touch from UI)
	wav *WAVWriter

	// internal scratch buffers
	left, right []float64

	// seed for filename when recording starts
	seed int64
}

type recordCmd struct {
	start bool
	reply chan recordReply
}

type recordReply struct {
	path string
	err  error
}

// NewRoot constructs a Root for the given algorithm and scope sink.
func NewRoot(algo gen.Algorithm, ring *scope.Ring) *Root {
	r := &Root{
		algo:   algo,
		scope:  ring,
		recCmd: make(chan recordCmd, 4),
	}
	r.volume.Store(70)
	return r
}

// SetSeed records the seed used to construct the algorithm; used in WAV filenames.
func (r *Root) SetSeed(s int64) { r.seed = s }

// SetVolume sets master volume (0..100, clamped).
func (r *Root) SetVolume(pct int) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	r.volume.Store(uint32(pct))
}

// TogglePause flips paused state. While paused, Stream emits silence but still
// runs the algorithm in the background to keep time moving (so unpausing feels
// continuous). v1 keeps it simple: skip the algorithm and emit zeros.
func (r *Root) TogglePause() { r.paused.Store(!r.paused.Load()) }

// ToggleRecord starts or stops recording. Filename pattern:
// `termus-<seed>-<unix>.wav` in the current directory.
func (r *Root) ToggleRecord() (string, error) {
	reply := make(chan recordReply, 1)
	r.recCmd <- recordCmd{start: r.wavStartRequested(), reply: reply}
	rep := <-reply
	return rep.path, rep.err
}

// wavStartRequested decides whether the next toggle should start or stop;
// callers don't see internal state.
func (r *Root) wavStartRequested() bool {
	// The audio thread owns r.wav. The UI thread races on read here only to
	// decide START vs STOP, which is okay — worst case the user double-clicks
	// and the audio thread sees two STARTs back-to-back; we handle that there.
	return r.wav == nil
}

// Stream implements beep.Streamer.
func (r *Root) Stream(samples [][2]float64) (n int, ok bool) {
	r.handleCommands()

	n = len(samples)
	if cap(r.left) < n {
		r.left = make([]float64, n)
		r.right = make([]float64, n)
	}
	r.left = r.left[:n]
	r.right = r.right[:n]

	if r.paused.Load() {
		for i := range samples {
			samples[i][0] = 0
			samples[i][1] = 0
		}
	} else {
		r.algo.Next(r.left, r.right)
		gain := float64(r.volume.Load()) / 100.0
		for i := range samples {
			samples[i][0] = r.left[i] * gain
			samples[i][1] = r.right[i] * gain
		}
	}

	// Scope tap: mix L+R from the final output to mono and push.
	mono := r.left // reuse left buffer for the mono mix
	for i := range mono {
		mono[i] = (samples[i][0] + samples[i][1]) * 0.5
	}
	r.scope.Write(mono)

	// WAV tap.
	if r.wav != nil {
		if err := r.wav.Write(samples); err != nil {
			_ = r.wav.Close()
			r.wav = nil
			// v1: stop recording silently on write error. See spec error handling.
		}
	}

	return n, true
}

// Err implements beep.Streamer.
func (r *Root) Err() error { return nil }

// handleCommands drains record commands. Audio thread only.
func (r *Root) handleCommands() {
	for {
		select {
		case cmd := <-r.recCmd:
			if cmd.start {
				if r.wav != nil {
					cmd.reply <- recordReply{err: fmt.Errorf("already recording")}
					continue
				}
				name := fmt.Sprintf("termus-%d-%d.wav", r.seed, time.Now().Unix())
				path, err := filepath.Abs(name)
				if err != nil {
					cmd.reply <- recordReply{err: err}
					continue
				}
				w, err := NewWAVWriter(path, 44100, 2)
				if err != nil {
					cmd.reply <- recordReply{err: err}
					continue
				}
				r.wav = w
				cmd.reply <- recordReply{path: path}
			} else {
				if r.wav == nil {
					cmd.reply <- recordReply{}
					continue
				}
				err := r.wav.Close()
				r.wav = nil
				cmd.reply <- recordReply{err: err}
			}
		default:
			return
		}
	}
}
