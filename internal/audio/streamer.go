package audio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

// AudioProducer generates the next track's WAV bytes given a sequence number.
// Implementations may be the ACE-Step HTTP client or a mock for tests.
//
// Produce should block until either the audio is ready or ctx is cancelled.
// It must be safe to call concurrently from one background goroutine.
type AudioProducer interface {
	Produce(ctx context.Context, seq int) ([]byte, error)
}

// AudioSink is the destination for decoded streams. The default sink wraps
// gopxl/beep speaker; tests replace it with a recording sink to verify
// playback order without depending on a real output device.
//
// Play plays s synchronously: it MUST block until s has finished streaming
// (or ctx is cancelled). Implementations are expected to handle their own
// resource teardown when ctx is cancelled.
type AudioSink interface {
	Play(ctx context.Context, s beep.Streamer, format beep.Format) error
}

// ScopeSink receives mono samples drawn from whatever the streamer is
// currently playing, so the TUI visualizer (which reads from scope.Ring) can
// move while ACE-Step audio plays. The streamer calls Write on every frame it
// hands to the AudioSink. A nil ScopeSink is fine — it disables the tap.
//
// The concrete production implementation is *scope.Ring, whose Write takes
// mono samples; we deliberately keep this as a minimal interface so the
// audio package doesn't depend on internal/scope.
type ScopeSink interface {
	Write(samples []float64)
}

// StreamerConfig controls the streaming behaviour.
type StreamerConfig struct {
	// Producer is the source of audio. Required.
	Producer AudioProducer

	// Sink is the destination. When nil, NewStreamer fills in a speaker-backed
	// sink that initialises the OS audio device on first Play.
	Sink AudioSink

	// ScopeSink, when non-nil, receives a mono mix of every frame the
	// streamer pushes to the AudioSink. This is how the ACE-Step playback
	// path feeds the TUI scope visualizer; the SF2 path wires its own
	// scope.Ring inside audio.Root.Stream. Leave nil for headless mode
	// (or tests) to disable the tap.
	ScopeSink ScopeSink

	// QueueDepth is the look-ahead depth. Default 2 means: while track N
	// plays, tracks N+1 and N+2 may already be generated and waiting.
	QueueDepth int

	// CrossfadeSec is the equal-power crossfade overlap between adjacent
	// tracks, in seconds. Default 3.0.
	CrossfadeSec float64

	// MaxTracks limits the total number of tracks the streamer will play.
	// 0 = infinite. Useful for tests and for the --max-tracks CLI flag.
	MaxTracks int

	// Logger receives one human-readable line per state change. nil =
	// io.Discard.
	Logger io.Writer
}

// Status is a snapshot of the streamer's runtime state. Safe for the TUI to
// poll on a timer.
type Status struct {
	Playing    bool
	CurrentSeq int
	QueueDepth int
	LastError  error
}

// trackBuffer is one queued (or in-flight) track from the producer.
type trackBuffer struct {
	Seq  int
	Data []byte
	Err  error
}

// Streamer pumps audio from the producer to the sink with look-ahead
// queueing and equal-power crossfade between tracks.
//
// Lifecycle:
//
//	s := NewStreamer(cfg)
//	if err := s.Start(ctx); err != nil { ... }
//	... time passes; tracks play; status updates ...
//	s.Stop()        // graceful shutdown, waits for goroutines
//
// Concurrency: Start and Stop are single-threaded (call once each). Skip,
// Status, and the producer/player internals use a mutex.
type Streamer struct {
	cfg StreamerConfig

	queue chan trackBuffer
	skip  chan struct{}

	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.RWMutex
	status Status

	started bool
}

// NewStreamer builds a streamer from cfg, filling defaults. Returns a non-nil
// Streamer even when cfg.Producer is nil; Start will fail in that case.
func NewStreamer(cfg StreamerConfig) *Streamer {
	if cfg.QueueDepth <= 0 {
		cfg.QueueDepth = 2
	}
	if cfg.CrossfadeSec <= 0 {
		cfg.CrossfadeSec = 3.0
	}
	if cfg.Logger == nil {
		cfg.Logger = io.Discard
	}
	if cfg.Sink == nil {
		cfg.Sink = &speakerSink{}
	}
	return &Streamer{
		cfg:   cfg,
		queue: make(chan trackBuffer, cfg.QueueDepth),
		skip:  make(chan struct{}, 1),
	}
}

// Start spawns the producer + player goroutines and returns immediately.
// Returns an error only when the streamer was already started or when the
// configuration is invalid.
func (s *Streamer) Start(ctx context.Context) error {
	if s.cfg.Producer == nil {
		return fmt.Errorf("streamer: nil Producer")
	}
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("streamer: already started")
	}
	s.started = true
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.wg.Add(2)
	go s.producerLoop(ctx)
	go s.playerLoop(ctx)
	return nil
}

// Stop cancels both goroutines and waits for them to exit. Safe to call
// multiple times. After Stop the streamer cannot be restarted.
func (s *Streamer) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.setPlaying(false)
}

// Skip signals the player to abandon the current track and move to the next.
// Non-blocking; if a Skip is already pending the call is a no-op.
func (s *Streamer) Skip() {
	select {
	case s.skip <- struct{}{}:
	default:
	}
}

// Status returns a snapshot of the streamer's state. Safe to call from any
// goroutine.
func (s *Streamer) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Streamer) producerLoop(ctx context.Context) {
	defer s.wg.Done()
	seq := 0
	for {
		// Honour MaxTracks: once we've enqueued that many, close the
		// queue and let the player drain.
		if s.cfg.MaxTracks > 0 && seq >= s.cfg.MaxTracks {
			close(s.queue)
			return
		}
		buf, err := s.cfg.Producer.Produce(ctx, seq)
		if ctx.Err() != nil {
			close(s.queue)
			return
		}
		if err != nil {
			// Send the error and stop producing further tracks. The
			// player loop will surface it via Status and exit.
			fmt.Fprintf(s.cfg.Logger, "streamer: producer error at seq=%d: %v\n", seq, err)
			select {
			case s.queue <- trackBuffer{Seq: seq, Err: err}:
			case <-ctx.Done():
			}
			close(s.queue)
			return
		}
		select {
		case s.queue <- trackBuffer{Seq: seq, Data: buf}:
		case <-ctx.Done():
			close(s.queue)
			return
		}
		s.updateQueueDepth()
		seq++
	}
}

func (s *Streamer) playerLoop(ctx context.Context) {
	defer s.wg.Done()
	// previous holds the just-played streamer when there's a pending
	// crossfade. We can't pre-mix without an active speaker context; the
	// sink owns the device. So the crossfade is implemented by handing
	// the sink a chained streamer where the trailing tail of the
	// previous track is gain-faded down and the new track is gain-faded
	// up, summed in a beep.Mixer for the overlap window.
	//
	// On first iteration there's no previous, so the new track plays
	// alone with no fade.

	var pending *playableTrack
	for {
		var tb trackBuffer
		var ok bool
		select {
		case <-ctx.Done():
			return
		case tb, ok = <-s.queue:
			if !ok {
				// Producer drained. Play any pending track alone
				// and exit.
				if pending != nil {
					s.playOne(ctx, pending, nil)
				}
				return
			}
		}
		if tb.Err != nil {
			s.setLastError(tb.Err)
			return
		}
		next, err := decodeTrack(tb.Data, tb.Seq)
		if err != nil {
			fmt.Fprintf(s.cfg.Logger, "streamer: decode seq=%d: %v\n", tb.Seq, err)
			s.setLastError(err)
			return
		}
		if pending != nil {
			s.playOne(ctx, pending, next)
		}
		pending = next
		s.updateQueueDepth()
	}
}

// playableTrack is a decoded track plus its format. The streamer is consumed
// once.
type playableTrack struct {
	Seq    int
	Stream beep.StreamSeekCloser
	Format beep.Format
}

// decodeTrack wraps wav.Decode for the in-memory WAV bytes returned by the
// producer. The returned StreamSeekCloser must be Close()d by the caller
// once the sink is done with it.
func decodeTrack(data []byte, seq int) (*playableTrack, error) {
	rdr := bytes.NewReader(data)
	stream, format, err := wav.Decode(rdr)
	if err != nil {
		return nil, fmt.Errorf("seq=%d: wav decode: %w", seq, err)
	}
	return &playableTrack{Seq: seq, Stream: stream, Format: format}, nil
}

// playOne plays current and, if next != nil, crossfades into next during the
// last cfg.CrossfadeSec of current. After this returns, current.Stream has
// been Close()d; next.Stream's head has been consumed up through the
// overlap window and remains positioned for normal playback on the next
// invocation.
//
// When next is nil (last track in MaxTracks runs, or producer drained), the
// current track plays straight through and the trailing fade-out is just a
// gain ramp to zero so the speaker doesn't pop.
func (s *Streamer) playOne(ctx context.Context, current *playableTrack, next *playableTrack) {
	defer current.Stream.Close()

	s.setStatus(Status{
		Playing:    true,
		CurrentSeq: current.Seq,
		QueueDepth: s.snapshotQueueDepth(),
	})

	fadeSamples := int(s.cfg.CrossfadeSec * float64(current.Format.SampleRate))
	totalSamples := current.Stream.Len()
	bodySamples := totalSamples - fadeSamples
	if bodySamples < 0 {
		bodySamples = 0
		fadeSamples = totalSamples
	}

	var toPlay beep.Streamer
	if next == nil || fadeSamples <= 0 {
		// Last track, or crossfade window doesn't fit. Apply a simple
		// fade-out tail so playback ends cleanly.
		body := beep.Take(bodySamples, current.Stream)
		tailFade := effects.Transition(
			beep.Take(fadeSamples, current.Stream),
			fadeSamples, 1.0, 0.0, effects.TransitionEqualPower,
		)
		toPlay = beep.Seq(body, tailFade)
	} else {
		// Body of current plays normally.
		body := beep.Take(bodySamples, current.Stream)

		// Tail of current fades out via equal-power; head of next fades
		// in. Mixer sums them sample-for-sample.
		fadeOut := effects.Transition(
			beep.Take(fadeSamples, current.Stream),
			fadeSamples, 1.0, 0.0, effects.TransitionEqualPower,
		)
		// Per the brief: equal-power. The default TransitionEqualPower
		// uses a cosine that, paired with the inverse on the other
		// stream, keeps perceived loudness constant.
		fadeIn := effects.Transition(
			beep.Take(fadeSamples, next.Stream),
			fadeSamples, 0.0, 1.0, effects.TransitionEqualPower,
		)
		mixer := &beep.Mixer{}
		mixer.Add(fadeOut)
		mixer.Add(fadeIn)
		overlap := beep.Take(fadeSamples, mixer)
		toPlay = beep.Seq(body, overlap)
	}

	// Skip support: wrap toPlay in a stopper.
	stopCh := make(chan struct{})
	var src beep.Streamer = toPlay
	// Scope tap: if the config has a ScopeSink, wrap the stream so every
	// frame the speaker consumes also flows into the TUI visualizer's ring
	// buffer. Mirrors how audio.Root.Stream feeds scope in the SF2 path
	// (see internal/audio/root.go). With no ScopeSink (headless mode or
	// tests) this wrapper is skipped entirely.
	if s.cfg.ScopeSink != nil {
		src = &scopeTapStreamer{src: src, sink: s.cfg.ScopeSink}
	}
	wrap := &skippableStreamer{src: src, stop: stopCh}

	playDone := make(chan error, 1)
	go func() {
		playDone <- s.cfg.Sink.Play(ctx, wrap, current.Format)
	}()

	for {
		select {
		case <-ctx.Done():
			close(stopCh)
			<-playDone
			return
		case <-s.skip:
			close(stopCh)
			<-playDone
			return
		case err := <-playDone:
			if err != nil && !errors.Is(err, context.Canceled) {
				fmt.Fprintf(s.cfg.Logger, "streamer: sink error seq=%d: %v\n", current.Seq, err)
				s.setLastError(err)
			}
			return
		}
	}
}

// scopeTapStreamer is a transparent passthrough that, after the underlying
// streamer produces a chunk of samples, mixes them down to mono and pushes
// them into a ScopeSink. The mono mix matches what audio.Root.Stream does for
// the SF2 path — (L + R) * 0.5 — so the visualizer sees comparable amplitudes
// regardless of which engine is feeding it.
//
// SP25 visualizer normalization: ACE-Step renders typically peak around
// -3 dBFS, while SF2 procedural output usually lands at -12 to -18 dBFS.
// That makes the TUI scope visualizer look like a wall of waveform on AI
// tracks and a thin ribbon on SF2 tracks. We apply a fixed -6 dB attenuation
// (×0.5) to the scope-tap copy only, which brings ACE-Step closer to SF2's
// typical scope range. The speaker output is untouched — only the visualizer
// tap is scaled.
const scopeTapGain = 0.5

type scopeTapStreamer struct {
	src  beep.Streamer
	sink ScopeSink
	mono []float64 // reused per Stream() call
}

func (t *scopeTapStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = t.src.Stream(samples)
	if n > 0 && t.sink != nil {
		if cap(t.mono) < n {
			t.mono = make([]float64, n)
		}
		mono := t.mono[:n]
		for i := 0; i < n; i++ {
			mono[i] = (samples[i][0] + samples[i][1]) * 0.5 * scopeTapGain
		}
		t.sink.Write(mono)
	}
	return n, ok
}

func (t *scopeTapStreamer) Err() error {
	return t.src.Err()
}

// skippableStreamer wraps a beep.Streamer with a "stop" signal that causes
// it to return ok=false on the next Stream call. This is the cooperative
// interrupt mechanism used by Skip.
type skippableStreamer struct {
	src    beep.Streamer
	stop   chan struct{}
	closed bool
}

func (s *skippableStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.closed {
		return 0, false
	}
	select {
	case <-s.stop:
		s.closed = true
		return 0, false
	default:
	}
	return s.src.Stream(samples)
}

func (s *skippableStreamer) Err() error {
	return s.src.Err()
}

func (s *Streamer) setStatus(st Status) {
	s.mu.Lock()
	prev := s.status
	st.LastError = prev.LastError
	s.status = st
	s.mu.Unlock()
}

func (s *Streamer) setPlaying(p bool) {
	s.mu.Lock()
	s.status.Playing = p
	s.mu.Unlock()
}

func (s *Streamer) setLastError(err error) {
	s.mu.Lock()
	s.status.LastError = err
	s.mu.Unlock()
}

func (s *Streamer) updateQueueDepth() {
	s.mu.Lock()
	s.status.QueueDepth = len(s.queue)
	s.mu.Unlock()
}

func (s *Streamer) snapshotQueueDepth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status.QueueDepth
}

// speakerSink is the production AudioSink. It initialises the OS speaker on
// first Play and reuses it thereafter (re-initialising on each track would
// glitch the device). The speaker package is global, so multiple concurrent
// speakerSinks are not supported.
type speakerSink struct {
	mu          sync.Mutex
	initialised bool
	sampleRate  beep.SampleRate
}

func (s *speakerSink) Play(ctx context.Context, stream beep.Streamer, format beep.Format) error {
	s.mu.Lock()
	// (Re-)initialise the speaker if this is the first track for this sink
	// OR if the sample rate has changed since the last init. The latter
	// happens routinely when termus hot-switches between SF2 (44.1k) and
	// ACE-Step (48k); beep.speaker.Init is safe to call multiple times and
	// switches the underlying device to the new rate.
	if !s.initialised || format.SampleRate != s.sampleRate {
		// Clear any queued audio before reinit so we don't carry over
		// half-decoded buffers at the wrong rate.
		if s.initialised {
			speaker.Clear()
		}
		// 1/10s buffer is a reasonable starting point; small enough
		// for crossfade transitions to feel snappy without underruns.
		bufSize := format.SampleRate.N(100 * time.Millisecond)
		if err := speaker.Init(format.SampleRate, bufSize); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("speakerSink: speaker.Init: %w", err)
		}
		s.initialised = true
		s.sampleRate = format.SampleRate
	}
	s.mu.Unlock()

	done := make(chan struct{})
	speaker.Play(beep.Seq(stream, beep.Callback(func() { close(done) })))
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		speaker.Clear()
		return ctx.Err()
	}
}
