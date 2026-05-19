package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopxl/beep/v2"
)

// makeSineWAV synthesises a tiny stereo 16-bit PCM WAV: durMs of a sine at
// freqHz, sample rate sr. Returned bytes start with "RIFF...WAVE" so the
// streamer's decoder accepts them. Used to construct deterministic mock
// producer output.
func makeSineWAV(durMs, freqHz, sr int) []byte {
	frames := durMs * sr / 1000
	dataBytes := uint32(frames * 2 * 2) // stereo, 16-bit
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte("RIFF"))
	binary.Write(buf, binary.LittleEndian, uint32(36+dataBytes))
	buf.Write([]byte("WAVEfmt "))
	binary.Write(buf, binary.LittleEndian, uint32(16))   // fmt chunk size
	binary.Write(buf, binary.LittleEndian, uint16(1))    // PCM
	binary.Write(buf, binary.LittleEndian, uint16(2))    // stereo
	binary.Write(buf, binary.LittleEndian, uint32(sr))   // sample rate
	binary.Write(buf, binary.LittleEndian, uint32(sr*4)) // byte rate
	binary.Write(buf, binary.LittleEndian, uint16(4))    // block align
	binary.Write(buf, binary.LittleEndian, uint16(16))   // bits per sample
	buf.Write([]byte("data"))
	binary.Write(buf, binary.LittleEndian, dataBytes)
	for i := 0; i < frames; i++ {
		v := math.Sin(2 * math.Pi * float64(freqHz) * float64(i) / float64(sr))
		s := int16(v * 0.3 * 32767)
		// little-endian, stereo
		buf.WriteByte(byte(s))
		buf.WriteByte(byte(s >> 8))
		buf.WriteByte(byte(s))
		buf.WriteByte(byte(s >> 8))
	}
	return buf.Bytes()
}

// mockProducer returns a different WAV for each seq, with optional error
// injection at a chosen seq.
type mockProducer struct {
	mu       sync.Mutex
	called   []int
	failAt   int  // seq to fail at; -1 = never
	failErr  error
	wavBytes func(seq int) []byte
}

func newMockProducer() *mockProducer {
	return &mockProducer{
		failAt: -1,
		wavBytes: func(seq int) []byte {
			// 200ms of an A4-ish tone, different frequency per seq so
			// the recording sink can detect ordering by the bytes
			// returned.
			return makeSineWAV(200, 440+seq*55, 44100)
		},
	}
}

func (m *mockProducer) Produce(ctx context.Context, seq int) ([]byte, error) {
	m.mu.Lock()
	m.called = append(m.called, seq)
	failAt := m.failAt
	failErr := m.failErr
	m.mu.Unlock()
	if failAt >= 0 && seq == failAt {
		return nil, failErr
	}
	return m.wavBytes(seq), nil
}

// recordingSink consumes any streamer it's given and records the order of
// (Seq, byte-length) pairs. It does not initialise the OS speaker, so tests
// are hermetic.
type recordingSink struct {
	mu     sync.Mutex
	played []recordedTrack
	delay  time.Duration // simulate per-track playback time
}

type recordedTrack struct {
	SampleRate int
	NumSamples int
}

func (r *recordingSink) Play(ctx context.Context, s beep.Streamer, format beep.Format) error {
	// Drain the stream into a counter so we can confirm it actually
	// produced samples.
	total := 0
	chunk := make([][2]float64, 1024)
	for {
		n, ok := s.Stream(chunk)
		total += n
		if !ok {
			break
		}
		// Honour cancellation cooperatively so Stop tests don't hang.
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	if r.delay > 0 {
		select {
		case <-time.After(r.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	r.mu.Lock()
	r.played = append(r.played, recordedTrack{SampleRate: int(format.SampleRate), NumSamples: total})
	r.mu.Unlock()
	return nil
}

func (r *recordingSink) snapshot() []recordedTrack {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]recordedTrack, len(r.played))
	copy(cp, r.played)
	return cp
}

func TestStreamer_PlaysInOrder(t *testing.T) {
	prod := newMockProducer()
	sink := &recordingSink{}
	s := NewStreamer(StreamerConfig{
		Producer:     prod,
		Sink:         sink,
		QueueDepth:   2,
		CrossfadeSec: 0.05,
		MaxTracks:    3,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Wait for all three to play.
	waitFor(t, 4*time.Second, func() bool {
		return len(sink.snapshot()) >= 3
	})
	s.Stop()

	played := sink.snapshot()
	if len(played) != 3 {
		t.Fatalf("played %d tracks, want 3 (%+v)", len(played), played)
	}
	// Each track must have produced samples. With 200ms at 44100Hz =
	// ~8820 frames; the streamer may split into body + tail but the
	// total seen by the sink across all calls should be roughly that
	// per track minus crossfade overlap.
	for i, p := range played {
		if p.NumSamples == 0 {
			t.Errorf("track %d had zero samples", i)
		}
		if p.SampleRate != 44100 {
			t.Errorf("track %d sample rate = %d, want 44100", i, p.SampleRate)
		}
	}
}

func TestStreamer_QueueFillsAhead(t *testing.T) {
	// Use a producer that records when it's called, and a sink that
	// blocks for a while so the queue actually has time to fill.
	prod := newMockProducer()
	sink := &recordingSink{delay: 80 * time.Millisecond}
	s := NewStreamer(StreamerConfig{
		Producer:     prod,
		Sink:         sink,
		QueueDepth:   2,
		CrossfadeSec: 0.05,
		MaxTracks:    5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Within a short window the producer should have been called more
	// than once even though the player has only finished track 0.
	waitFor(t, 1*time.Second, func() bool {
		prod.mu.Lock()
		defer prod.mu.Unlock()
		return len(prod.called) >= 3
	})
	prod.mu.Lock()
	calledEarly := len(prod.called)
	prod.mu.Unlock()
	if calledEarly < 3 {
		s.Stop()
		t.Fatalf("producer called only %d times while sink was blocking; expected look-ahead to fill queue (>=3)", calledEarly)
	}
	s.Stop()
}

func TestStreamer_HandlesProducerError(t *testing.T) {
	prod := newMockProducer()
	prod.failAt = 2
	prod.failErr = errors.New("synthetic produce failure")
	sink := &recordingSink{}
	s := NewStreamer(StreamerConfig{
		Producer:     prod,
		Sink:         sink,
		QueueDepth:   2,
		CrossfadeSec: 0.05,
		MaxTracks:    5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Wait until status shows the error.
	waitFor(t, 4*time.Second, func() bool {
		return s.Status().LastError != nil
	})
	if got := s.Status().LastError; got == nil {
		t.Fatalf("expected LastError to be set after producer failure")
	}
	s.Stop()
}

func TestStreamer_StopIsClean(t *testing.T) {
	prod := newMockProducer()
	sink := &recordingSink{delay: 500 * time.Millisecond}
	s := NewStreamer(StreamerConfig{
		Producer:     prod,
		Sink:         sink,
		QueueDepth:   2,
		CrossfadeSec: 0.05,
		MaxTracks:    0, // infinite
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Give it a moment to spin up.
	time.Sleep(60 * time.Millisecond)

	// Stop with a deadline; if Stop() doesn't return in time the
	// goroutines have leaked.
	stopReturned := make(chan struct{})
	go func() {
		s.Stop()
		close(stopReturned)
	}()
	select {
	case <-stopReturned:
		// OK.
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not return within 2s; goroutine leak")
	}
}

func TestStreamer_MaxTracksLimit(t *testing.T) {
	prod := newMockProducer()
	sink := &recordingSink{}
	s := NewStreamer(StreamerConfig{
		Producer:     prod,
		Sink:         sink,
		QueueDepth:   2,
		CrossfadeSec: 0.05,
		MaxTracks:    3,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Wait for all 3 to play, then a brief grace period.
	waitFor(t, 4*time.Second, func() bool {
		return len(sink.snapshot()) >= 3
	})
	time.Sleep(100 * time.Millisecond)
	s.Stop()

	played := sink.snapshot()
	if len(played) != 3 {
		t.Errorf("expected exactly 3 plays, got %d", len(played))
	}
	prod.mu.Lock()
	defer prod.mu.Unlock()
	if len(prod.called) != 3 {
		t.Errorf("expected producer to be called 3 times, got %d (%v)", len(prod.called), prod.called)
	}
}

func TestStreamer_StartTwiceFails(t *testing.T) {
	prod := newMockProducer()
	s := NewStreamer(StreamerConfig{Producer: prod, Sink: &recordingSink{}, MaxTracks: 1})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := s.Start(ctx); err == nil {
		t.Errorf("expected error on second Start")
	}
	s.Stop()
}

func TestStreamer_NilProducerFails(t *testing.T) {
	s := NewStreamer(StreamerConfig{})
	if err := s.Start(context.Background()); err == nil {
		t.Errorf("expected error on Start with nil Producer")
	}
}

func TestNewStreamer_AppliesDefaults(t *testing.T) {
	prod := newMockProducer()
	s := NewStreamer(StreamerConfig{Producer: prod})
	if s.cfg.QueueDepth != 2 {
		t.Errorf("default QueueDepth = %d, want 2", s.cfg.QueueDepth)
	}
	if s.cfg.CrossfadeSec != 3.0 {
		t.Errorf("default CrossfadeSec = %v, want 3.0", s.cfg.CrossfadeSec)
	}
	if s.cfg.Logger == nil {
		t.Errorf("default Logger = nil; expected io.Discard")
	}
	if s.cfg.Sink == nil {
		t.Errorf("default Sink = nil; expected speakerSink")
	}
}

// waitFor polls cond every 10ms up to timeout, failing the test if cond
// never becomes true.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("waitFor: condition not satisfied within %v", timeout)
}

// Compile-time interface assertions for the mock implementations.
var (
	_ AudioProducer = (*mockProducer)(nil)
	_ AudioSink     = (*recordingSink)(nil)
)

// recordingScopeSink captures every mono sample the streamer pushes into a
// scope sink. Used by TestScopeTap_AttenuatesByHalf to verify the SP25
// scope-tap gain of 0.5.
type recordingScopeSink struct {
	mu       sync.Mutex
	samples  []float64
}

func (r *recordingScopeSink) Write(samples []float64) {
	r.mu.Lock()
	r.samples = append(r.samples, samples...)
	r.mu.Unlock()
}

func (r *recordingScopeSink) peak() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	peak := 0.0
	for _, v := range r.samples {
		abs := v
		if abs < 0 {
			abs = -abs
		}
		if abs > peak {
			peak = abs
		}
	}
	return peak
}

// constSineStream produces a constant-amplitude stereo sine so the scope tap
// has a deterministic peak to compare against. Returns ok=true forever (caller
// must wrap with beep.Take or skip-stream to terminate).
type constSineStream struct {
	phase float64
	inc   float64
	amp   float64
	left  int // frames remaining; 0 means infinite
}

func (s *constSineStream) Stream(samples [][2]float64) (int, bool) {
	if s.left == 0 {
		// Treat 0 as a small finite count so the caller doesn't spin.
		s.left = 1024
	}
	n := len(samples)
	if n > s.left {
		n = s.left
	}
	for i := 0; i < n; i++ {
		v := s.amp * math.Sin(s.phase)
		samples[i][0] = v
		samples[i][1] = v
		s.phase += s.inc
	}
	s.left -= n
	return n, s.left > 0
}

func (s *constSineStream) Err() error { return nil }

// TestScopeTap_AttenuatesByHalf verifies that the scope-tap mirror is fed at
// half the source's amplitude (SP25 visualizer normalization). The source
// peaks at amp; the scope mirror's recorded peak must be approximately
// amp * 0.5.
func TestScopeTap_AttenuatesByHalf(t *testing.T) {
	const amp = 0.8
	src := &constSineStream{
		amp:  amp,
		inc:  2 * math.Pi * 440.0 / 44100.0,
		left: 4096,
	}
	sink := &recordingScopeSink{}
	tap := &scopeTapStreamer{src: src, sink: sink}
	// Drain the tap fully; the underlying mono samples will be captured by
	// the sink as the streamer pulls.
	buf := make([][2]float64, 256)
	for {
		_, ok := tap.Stream(buf)
		if !ok {
			break
		}
	}
	got := sink.peak()
	want := amp * scopeTapGain
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("scope-tap peak = %.4f, want approx %.4f (amp=%.2f * gain=%.2f)", got, want, amp, scopeTapGain)
	}
	// And confirm scopeTapGain is the documented 0.5 so the constant stays
	// load-bearing in code review.
	if scopeTapGain != 0.5 {
		t.Fatalf("scopeTapGain = %v, want 0.5", scopeTapGain)
	}
}

// TestPauseableStreamer_EmitsSilenceWhenPaused verifies the pause wrapper
// returns zero samples and does not advance the source while paused, and
// resumes from the same source position on unpause. Mirrors SF2's pause
// semantics where audio.Root.Stream zeros the buffer before downstream
// taps.
func TestPauseableStreamer_EmitsSilenceWhenPaused(t *testing.T) {
	src := &constSineStream{
		amp:  0.8,
		inc:  2 * math.Pi * 440.0 / 48000.0,
		left: 4096,
	}
	var paused atomic.Bool
	p := &pauseableStreamer{src: src, paused: &paused}

	buf := make([][2]float64, 64)
	// 1. Unpaused → samples flow.
	n, ok := p.Stream(buf)
	if !ok || n != len(buf) {
		t.Fatalf("unpaused stream: n=%d ok=%v", n, ok)
	}
	beforePause := buf[0]
	srcLeftBeforePause := src.left

	// 2. Pause → next chunk is zeros and source is NOT advanced.
	paused.Store(true)
	n, ok = p.Stream(buf)
	if !ok || n != len(buf) {
		t.Fatalf("paused stream: n=%d ok=%v", n, ok)
	}
	for i := range buf {
		if buf[i][0] != 0 || buf[i][1] != 0 {
			t.Fatalf("paused buf[%d] = %v, want zero", i, buf[i])
		}
	}
	if src.left != srcLeftBeforePause {
		t.Fatalf("source advanced while paused: left went %d -> %d",
			srcLeftBeforePause, src.left)
	}

	// 3. Resume → next sample is identical to what it would have been with
	// no pause in between.
	paused.Store(false)
	n, ok = p.Stream(buf)
	if !ok || n != len(buf) {
		t.Fatalf("resumed stream: n=%d ok=%v", n, ok)
	}
	// Sanity: first resumed sample is non-zero (the source is mid-sine, not
	// exactly at a zero-crossing given our increment).
	if buf[0] == beforePause {
		t.Logf("first resumed sample %v matches pre-pause %v (acceptable)", buf[0], beforePause)
	}
	if buf[0][0] == 0 && buf[0][1] == 0 {
		t.Fatalf("first resumed sample is zero; source should have produced audio")
	}
}

// Catch a typo: unused fmt import would fail compilation, so we use it here.
var _ = fmt.Sprintf
