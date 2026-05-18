package synth_test

import (
	"math"
	"testing"

	"github.com/mrbrutti/termus/internal/synth"
)

const sr = float64(synth.SampleRate)

// TestLFOSineProducesExpectedRate verifies that a 2 Hz sine LFO produces
// approximately 4 zero-crossings per second (2 per cycle × 2 cycles).
func TestLFOSineProducesExpectedRate(t *testing.T) {
	lfo := synth.NewLFO(sr, synth.LFOConfig{
		Shape:  synth.LFOSine,
		RateHz: 2.0,
		Depth:  1.0,
	})
	const secs = 1
	samples := int(sr) * secs
	buf := make([]float64, samples)
	for i := range buf {
		buf[i] = lfo.Tick()
	}
	crossings := 0
	for i := 1; i < len(buf); i++ {
		if (buf[i-1] <= 0 && buf[i] > 0) || (buf[i-1] >= 0 && buf[i] < 0) {
			crossings++
		}
	}
	// 2 Hz × 2 crossings/cycle × 1 s = 4 crossings (±1 tolerance).
	if crossings < 3 || crossings > 5 {
		t.Fatalf("2 Hz sine LFO: %d zero-crossings, want ~4", crossings)
	}
}

// TestLFOTriangleStaysInBounds verifies that every tick is within [-Depth, +Depth].
func TestLFOTriangleStaysInBounds(t *testing.T) {
	const depth = 0.75
	lfo := synth.NewLFO(sr, synth.LFOConfig{
		Shape:  synth.LFOTriangle,
		RateHz: 3.0,
		Depth:  depth,
	})
	samples := int(sr * 5)
	for i := 0; i < samples; i++ {
		v := lfo.Tick()
		if v < -depth-1e-9 || v > depth+1e-9 {
			t.Fatalf("sample %d out of bounds: %g (depth=%g)", i, v, depth)
		}
	}
}

// TestLFOFadeInRamps verifies that the LFO output starts near zero and that
// the peak amplitude near the end of FadeInSec approaches the configured depth.
func TestLFOFadeInRamps(t *testing.T) {
	const fadeInSec = 0.5
	const depth = 1.0
	lfo := synth.NewLFO(sr, synth.LFOConfig{
		Shape:     synth.LFOSine,
		RateHz:    5.0,
		Depth:     depth,
		FadeInSec: fadeInSec,
	})

	// First sample should be at (or very near) zero amplitude (fade env=0).
	first := lfo.Tick()
	if math.Abs(first) > 0.01 {
		t.Fatalf("first sample with fade-in = %g, want ~0", first)
	}

	// Advance to the last 10% of the fade window and track peak amplitude.
	totalFadeSamples := int(fadeInSec * sr)
	// Skip to 90% through the fade.
	for i := 1; i < int(float64(totalFadeSamples)*0.9); i++ {
		lfo.Tick()
	}
	// Collect the remaining ~10% of fade samples and record the peak.
	var peak float64
	for i := int(float64(totalFadeSamples) * 0.9); i < totalFadeSamples; i++ {
		if v := math.Abs(lfo.Tick()); v > peak {
			peak = v
		}
	}
	// The sine completes several cycles in that window, so the peak should
	// reach ≥ 80% of depth.
	if peak < depth*0.8 {
		t.Fatalf("peak in last 10%% of fade = %g, want ≥ %g (depth=%g)", peak, depth*0.8, depth)
	}
}

// TestLFODelayHoldsAtZero verifies that the first DelaySec of ticks all
// return exactly zero.
func TestLFODelayHoldsAtZero(t *testing.T) {
	const delaySec = 0.2
	lfo := synth.NewLFO(sr, synth.LFOConfig{
		Shape:    synth.LFOSine,
		RateHz:   5.0,
		Depth:    1.0,
		DelaySec: delaySec,
	})
	delaySamples := int(delaySec * sr)
	for i := 0; i < delaySamples; i++ {
		v := lfo.Tick()
		if v != 0 {
			t.Fatalf("delay sample %d = %g, want 0", i, v)
		}
	}
}

// TestLFOResetReturnsToInitialState verifies that after Reset(), the LFO
// output sequence matches the original first N samples.
func TestLFOResetReturnsToInitialState(t *testing.T) {
	cfg := synth.LFOConfig{
		Shape:  synth.LFOSine,
		RateHz: 10.0,
		Depth:  1.0,
	}
	lfo := synth.NewLFO(sr, cfg)
	const n = 100
	first := make([]float64, n)
	for i := range first {
		first[i] = lfo.Tick()
	}
	lfo.Reset()
	for i := 0; i < n; i++ {
		got := lfo.Tick()
		if math.Abs(got-first[i]) > 1e-12 {
			t.Fatalf("sample %d after Reset: got %g, want %g", i, got, first[i])
		}
	}
}

// TestLFOSampleHoldChangesAtCycleBoundary verifies that the held value is
// constant between positive-going cycle boundaries.
func TestLFOSampleHoldChangesAtCycleBoundary(t *testing.T) {
	const rateHz = 10.0
	lfo := synth.NewLFO(sr, synth.LFOConfig{
		Shape:  synth.LFOSampleHold,
		RateHz: rateHz,
		Depth:  1.0,
		Seed:   42,
	})
	// Cycle length in samples.
	cycleSamples := int(sr / rateHz) // 4410 @ 44.1 kHz

	// Tick one full cycle and collect values.
	vals := make([]float64, cycleSamples)
	for i := range vals {
		vals[i] = lfo.Tick()
	}

	// Within the cycle, all samples should share the same value.
	v0 := vals[0]
	for i := 1; i < cycleSamples; i++ {
		if vals[i] != v0 {
			t.Fatalf("SampleHold changed mid-cycle at sample %d: %g → %g", i, v0, vals[i])
		}
	}

	// Tick another full cycle; its held value must differ from the first
	// (with overwhelming probability given a 64-bit RNG).
	next := make([]float64, cycleSamples)
	for i := range next {
		next[i] = lfo.Tick()
	}
	v1 := next[0]
	if v0 == v1 {
		t.Fatalf("SampleHold produced the same value in consecutive cycles: %g", v0)
	}
	// All samples in the second cycle must be the same value too.
	for i := 1; i < cycleSamples; i++ {
		if next[i] != v1 {
			t.Fatalf("SampleHold changed mid-cycle (cycle 2) at sample %d: %g → %g", i, v1, next[i])
		}
	}
}

// TestLFORandomWalkIsBounded verifies that over 10 seconds, the RandomWalk
// LFO stays within [-Depth, +Depth].
func TestLFORandomWalkIsBounded(t *testing.T) {
	const depth = 0.9
	lfo := synth.NewLFO(sr, synth.LFOConfig{
		Shape:  synth.LFORandomWalk,
		RateHz: 1.0, // rate doesn't affect walk; just here for completeness
		Depth:  depth,
		Seed:   7,
	})
	samples := int(sr * 10)
	for i := 0; i < samples; i++ {
		v := lfo.Tick()
		if v < -depth-1e-9 || v > depth+1e-9 {
			t.Fatalf("RandomWalk out of bounds at sample %d: %g (depth=%g)", i, v, depth)
		}
	}
}
