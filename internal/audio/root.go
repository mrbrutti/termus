package audio

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/synth"
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
	volume      atomic.Uint32 // 0..100
	paused      atomic.Bool
	streaming   atomic.Bool
	scopeMuted  atomic.Bool // set true when this Root must stop feeding the shared scope.Ring
	debugStatus atomic.Value

	// command channels for non-hot-path events
	recCmd   chan recordCmd
	algoSwap chan swapReq // UI thread pushes; audio thread picks up at top of next Stream call

	// audio-thread-owned state (do not touch from UI)
	wav *WAVWriter

	// fade state during a crossfade swap (audio thread only)
	fadeOutLeft int           // frames remaining of fade-out (algo == old)
	fadeInLeft  int           // frames remaining of fade-in  (algo == new)
	fadeTotal   int           // total frames per half (for ratio math)
	pendingAlgo gen.Algorithm // algorithm waiting to be swapped in at fade midpoint

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

// swapReq is one queued algorithm swap. fadeFrames is the per-half fade
// length: the swap fades out for fadeFrames, replaces the algorithm, then
// fades in for fadeFrames. Zero means an immediate swap with no fade.
type swapReq struct {
	algo       gen.Algorithm
	fadeFrames int
}

// Default fade lengths used by the public swap methods. The audio thread runs
// at 44.1 kHz, so 8820 frames ≈ 200 ms and 88200 frames ≈ 2 s.
const (
	defaultSwapFade     = 8820  // ~200 ms — keyboard cycling
	defaultPlaylistFade = 88200 // ~2 s — playlist transitions
)

// NewRoot constructs a Root for the given algorithm and scope sink.
func NewRoot(algo gen.Algorithm, ring *scope.Ring) *Root {
	r := &Root{
		algo:     algo,
		scope:    ring,
		recCmd:   make(chan recordCmd, 4),
		algoSwap: make(chan swapReq, 4),
	}
	r.volume.Store(70)
	r.storeDebugStatus(algo)
	return r
}

// SwapAlgorithm hot-swaps the running algorithm with a short (~200 ms)
// fade-out / fade-in to avoid clicks when called from keyboard cycling.
// Non-blocking: drops the swap if the channel is full (UI mash protection).
func (r *Root) SwapAlgorithm(algo gen.Algorithm) {
	r.queueSwap(swapReq{algo: algo, fadeFrames: defaultSwapFade})
}

// SwapAlgorithmFade is like SwapAlgorithm but with a caller-specified fade
// length. Used for playlist transitions, which want a longer (~2 s) crossfade.
func (r *Root) SwapAlgorithmFade(algo gen.Algorithm, fadeFrames int) {
	if fadeFrames < 0 {
		fadeFrames = 0
	}
	r.queueSwap(swapReq{algo: algo, fadeFrames: fadeFrames})
}

func (r *Root) queueSwap(req swapReq) {
	select {
	case r.algoSwap <- req:
	default:
	}
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

// DebugStatus returns the latest status snapshot published by the audio
// thread.
func (r *Root) DebugStatus() gen.DebugStatus {
	if v := r.debugStatus.Load(); v != nil {
		return v.(gen.DebugStatus)
	}
	return gen.DebugStatus{}
}

// MuteScope tells the audio thread to stop writing samples to the shared
// scope.Ring. Used by the SF2->ACE-Step pre-roll bridge: once ACE-Step's
// streamer becomes the speaker's primary source, the SF2 Root must stop
// racing it on the (single-writer) ring or the visualiser shows a frozen
// artifact. Idempotent.
func (r *Root) MuteScope() { r.scopeMuted.Store(true) }

// ToggleRecord starts or stops recording. Filename pattern:
// `termus-<seed>-<unix>.wav` in the current directory.
func (r *Root) ToggleRecord() (string, error) {
	if !r.streaming.Load() {
		return "", fmt.Errorf("audio backend not ready")
	}
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
	r.streaming.Store(true)
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
		r.renderWithFades(samples)
	}

	// Scope tap: mix L+R from the final output to mono and push.
	// scope.Ring is single-writer; the SF2->ACE-Step bridge calls
	// MuteScope on the outgoing Root so the ACE-Step streamer's scope
	// tap becomes the only writer (concurrent writers produce a frozen
	// artifact in the centre of the visualiser).
	if !r.scopeMuted.Load() && r.scope != nil {
		mono := r.left // reuse left buffer for the mono mix
		for i := range mono {
			mono[i] = (samples[i][0] + samples[i][1]) * 0.5
		}
		r.scope.Write(mono)
	}

	// WAV tap.
	if r.wav != nil {
		if err := r.wav.Write(samples); err != nil {
			_ = r.wav.Close()
			r.wav = nil
			// v1: stop recording silently on write error. See spec error handling.
		}
	}
	r.storeDebugStatus(r.algo)

	return n, true
}

// renderWithFades fills `samples` with audio from r.algo, splitting the
// buffer at any swap boundary so fade-out/fade-in envelopes can be applied
// across the join. Walks the buffer in segments that share a single
// fade phase (out, in, or none).
func (r *Root) renderWithFades(samples [][2]float64) {
	n := len(samples)

	for i := 0; i < n; {
		// How many samples fit before the next fade-phase boundary?
		segN := n - i
		switch {
		case r.fadeOutLeft > 0 && r.fadeOutLeft < segN:
			segN = r.fadeOutLeft
		case r.fadeOutLeft == 0 && r.fadeInLeft > 0 && r.fadeInLeft < segN:
			segN = r.fadeInLeft
		}

		// Render this segment with the active algorithm.
		masterGain := float64(r.volume.Load()) / 100.0 * gen.EffectiveOutputGain(r.algo)
		r.algo.Next(r.left[i:i+segN], r.right[i:i+segN])

		// Apply master gain modulated by fade envelope (if any).
		switch {
		case r.fadeOutLeft > 0:
			// gain at segment offset j = (fadeOutLeft - j) / fadeTotal
			for j := 0; j < segN; j++ {
				g := masterGain * float64(r.fadeOutLeft-j) / float64(r.fadeTotal)
				samples[i+j][0] = r.left[i+j] * g
				samples[i+j][1] = r.right[i+j] * g
			}
			r.fadeOutLeft -= segN
			if r.fadeOutLeft == 0 {
				// Crossover: swap to pending algo, start fade-in.
				r.algo = r.pendingAlgo
				r.pendingAlgo = nil
				r.fadeInLeft = r.fadeTotal
				r.storeDebugStatus(r.algo)
			}
		case r.fadeInLeft > 0:
			// gain at segment offset j = 1 - (fadeInLeft - j) / fadeTotal
			for j := 0; j < segN; j++ {
				g := masterGain * (1.0 - float64(r.fadeInLeft-j)/float64(r.fadeTotal))
				samples[i+j][0] = r.left[i+j] * g
				samples[i+j][1] = r.right[i+j] * g
			}
			r.fadeInLeft -= segN
		default:
			for j := 0; j < segN; j++ {
				samples[i+j][0] = r.left[i+j] * masterGain
				samples[i+j][1] = r.right[i+j] * masterGain
			}
		}

		i += segN
	}
}

// Err implements beep.Streamer.
func (r *Root) Err() error { return nil }

// handleCommands drains record and algo-swap commands. Audio thread only.
func (r *Root) handleCommands() {
	for {
		select {
		case req := <-r.algoSwap:
			// Reject the swap if a fade-in is already in progress — wait for
			// it to finish before queueing the next one (it'll just be lost).
			// Fade-out is fine to override (pendingAlgo just changes).
			if r.fadeInLeft > 0 {
				continue
			}
			if req.fadeFrames <= 0 {
				r.algo = req.algo
				r.storeDebugStatus(r.algo)
				continue
			}
			r.fadeTotal = req.fadeFrames
			r.fadeOutLeft = req.fadeFrames
			r.pendingAlgo = req.algo
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
				w, err := NewWAVWriter(path, synth.SampleRate, 2)
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

func (r *Root) storeDebugStatus(algo gen.Algorithm) {
	r.debugStatus.Store(gen.SnapshotDebugStatus(algo))
}
