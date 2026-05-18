package synth

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/madelynnblue/go-dsp/fft"
)

// TestIRLibraryHasEightPresets verifies the registry exports exactly 8 presets.
func TestIRLibraryHasEightPresets(t *testing.T) {
	presets := IRLibrary()
	if len(presets) != 8 {
		t.Fatalf("IRLibrary() returned %d presets, want 8", len(presets))
	}
}

// TestIRByNameRoundtrip verifies every name in IRLibrary() resolves via IRByName.
func TestIRByNameRoundtrip(t *testing.T) {
	for _, p := range IRLibrary() {
		got := IRByName(p.Name)
		if got == nil {
			t.Errorf("IRByName(%q) returned nil", p.Name)
			continue
		}
		if got.Name != p.Name {
			t.Errorf("IRByName(%q).Name = %q", p.Name, got.Name)
		}
	}
}

// TestIRByNameMissingReturnsNil verifies IRByName returns nil for unknown names.
func TestIRByNameMissingReturnsNil(t *testing.T) {
	if got := IRByName("not-a-preset"); got != nil {
		t.Errorf("IRByName(%q) = %+v, want nil", "not-a-preset", got)
	}
}

// TestEachPresetGeneratesBuffer verifies that each preset's Generate function
// returns a non-empty buffer whose length is within ±20% of RT60Sec*sampleRate.
func TestEachPresetGeneratesBuffer(t *testing.T) {
	const sr = 44100.0
	const seed = 1
	for _, p := range IRLibrary() {
		p := p // capture
		t.Run(p.Name, func(t *testing.T) {
			buf := p.Generate(sr, seed)
			if len(buf) == 0 {
				t.Fatalf("%s: Generate returned empty buffer", p.Name)
			}
			expected := p.RT60Sec * sr
			got := float64(len(buf))
			ratio := got / expected
			if ratio < 0.80 || ratio > 1.25 {
				t.Errorf("%s: buffer length %d (%.3fs), expected ≈ %.0f (%.3fs), ratio=%.2f",
					p.Name, len(buf), got/sr, expected, p.RT60Sec, ratio)
			}
		})
	}
}

// TestPresetsHaveDistinctRT60s verifies that bedroom_small and cathedral
// have RT60s that differ by at least 2 seconds (they should: 0.4 vs 4.0).
func TestPresetsHaveDistinctRT60s(t *testing.T) {
	small := IRByName("bedroom_small")
	cat := IRByName("cathedral")
	if small == nil || cat == nil {
		t.Fatal("required presets not found")
	}
	diff := cat.RT60Sec - small.RT60Sec
	if diff < 2.0 {
		t.Errorf("RT60 difference between cathedral (%.2fs) and bedroom_small (%.2fs) = %.2fs, want ≥ 2.0s",
			cat.RT60Sec, small.RT60Sec, diff)
	}
}

// TestSpringTankHasDistinctSpectralContent verifies that spring_tank has a
// higher spectral centroid than cathedral (spring = bright/chirpy; cathedral = dark tail).
func TestSpringTankHasDistinctSpectralContent(t *testing.T) {
	const sr = 44100.0
	const seed = 42

	springPreset := IRByName("spring_tank")
	catPreset := IRByName("cathedral")
	if springPreset == nil || catPreset == nil {
		t.Fatal("required presets not found")
	}

	springBuf := springPreset.Generate(sr, seed)
	catBuf := catPreset.Generate(sr, seed)

	springCentroid := spectralCentroidHz(springBuf, sr)
	catCentroid := spectralCentroidHz(catBuf, sr)

	t.Logf("spring_tank centroid=%.1f Hz, cathedral centroid=%.1f Hz", springCentroid, catCentroid)

	if springCentroid <= catCentroid {
		t.Errorf("spring_tank centroid (%.1f Hz) should be higher than cathedral (%.1f Hz)",
			springCentroid, catCentroid)
	}
}

// TestReverbBusAppliesPreDelay verifies that a click fed through a ReverbBus
// with PreDelayMs=20 produces its first non-zero output at or after the
// pre-delay offset (minus FFT-convolver block latency of 512 samples).
func TestReverbBusAppliesPreDelay(t *testing.T) {
	const sr = 44100.0
	const preDelayMs = 20.0
	preDelaySamples := int(preDelayMs * 0.001 * sr) // 882 samples

	bus, err := NewReverbBus(ReverbBusConfig{
		IRName:     "bedroom_small",
		PreDelayMs: preDelayMs,
		WetDB:      0,
		SampleRate: sr,
		Seed:       1,
	})
	if err != nil {
		t.Fatalf("NewReverbBus: %v", err)
	}

	// Feed a click at t=0, then silence.
	totalSamples := preDelaySamples + 1024
	var firstNonZero int = -1
	const blockLatency = 512 // FFTConvolver block latency

	for i := 0; i < totalSamples; i++ {
		var in float64
		if i == 0 {
			in = 1.0
		}
		wetL, _ := bus.Tick(in, in)
		if firstNonZero < 0 && math.Abs(wetL) > 1e-12 {
			firstNonZero = i
		}
	}

	if firstNonZero < 0 {
		t.Fatal("ReverbBus produced no non-zero output after click")
	}
	// The first non-zero output must be at or after (preDelaySamples - blockLatency).
	minExpected := preDelaySamples - blockLatency
	if minExpected < 0 {
		minExpected = 0
	}
	if firstNonZero < minExpected {
		t.Errorf("first non-zero output at sample %d, want ≥ %d (preDelay=%d, blockLatency=%d)",
			firstNonZero, minExpected, preDelaySamples, blockLatency)
	}
}

// TestReverbBusIsDeterministic verifies that two buses with the same config
// and seed produce identical output.
func TestReverbBusIsDeterministic(t *testing.T) {
	cfg := ReverbBusConfig{
		IRName:     "jazz_club",
		PreDelayMs: 15,
		WetDB:      -6,
		SampleRate: 44100,
		Seed:       99,
	}

	bus1, err := NewReverbBus(cfg)
	if err != nil {
		t.Fatalf("NewReverbBus (1): %v", err)
	}
	bus2, err := NewReverbBus(cfg)
	if err != nil {
		t.Fatalf("NewReverbBus (2): %v", err)
	}

	const samples = 4096
	for i := 0; i < samples; i++ {
		var in float64
		if i == 0 {
			in = 1.0
		}
		l1, r1 := bus1.Tick(in, in)
		l2, r2 := bus2.Tick(in, in)
		if l1 != l2 || r1 != r2 {
			t.Fatalf("sample %d: bus1=(%g,%g) bus2=(%g,%g) — not deterministic",
				i, l1, r1, l2, r2)
		}
	}
}

// spectralCentroidHz returns the magnitude-weighted mean frequency of buf.
// Inlined here to avoid an import cycle: audiotest imports synth, so synth
// tests cannot import audiotest.
func spectralCentroidHz(buf []float64, sampleRate float64) float64 {
	if len(buf) < 2 {
		return 0
	}
	n := 1
	for n*2 <= len(buf) {
		n *= 2
	}
	windowed := make([]float64, n)
	for i := 0; i < n; i++ {
		w := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
		windowed[i] = buf[i] * w
	}
	spec := fft.FFTReal(windowed)
	var num, den float64
	for k := 0; k < n/2; k++ {
		mag := cmplx.Abs(spec[k])
		freq := float64(k) * sampleRate / float64(n)
		num += freq * mag
		den += mag
	}
	if den == 0 {
		return 0
	}
	return num / den
}
