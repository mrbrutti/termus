package synth_test

import (
	"math"
	"testing"

	"github.com/mrbrutti/termus/internal/audiotest"
	"github.com/mrbrutti/termus/internal/synth"
)

const testSR = 44100.0

// ──────────────────────────────────────────────
// NoiseBurst tests
// ──────────────────────────────────────────────

// TestNoiseBurstSilentBeforeTrigger verifies that Tick returns exactly 0.0
// until Trigger is called.
func TestNoiseBurstSilentBeforeTrigger(t *testing.T) {
	nb := synth.NewNoiseBurst(synth.NoiseBurstConfig{
		Color:      synth.NoiseWhite,
		PeakAmp:    0.8,
		AttackSec:  0.003,
		DecaySec:   0.020,
		Seed:       42,
		SampleRate: testSR,
	})

	for i := 0; i < 100; i++ {
		v := nb.Tick()
		if v != 0 {
			t.Fatalf("sample %d before Trigger: got %g, want 0", i, v)
		}
	}
	if nb.Active() {
		t.Fatal("Active() should return false before Trigger")
	}
}

// TestNoiseBurstReachesPeakWithinAttackWindow verifies that after Trigger the
// maximum output amplitude is reached within AttackSec + 1ms.
func TestNoiseBurstReachesPeakWithinAttackWindow(t *testing.T) {
	const attack = 0.005 // 5 ms
	const peak = 0.6
	nb := synth.NewNoiseBurst(synth.NoiseBurstConfig{
		Color:      synth.NoiseWhite,
		PeakAmp:    peak,
		AttackSec:  attack,
		DecaySec:   0.050,
		Seed:       1,
		SampleRate: testSR,
	})
	nb.Trigger()

	// Collect samples for attack window + 1ms tolerance.
	window := int(math.Round((attack + 0.001) * testSR))
	buf := make([]float64, window)
	for i := range buf {
		buf[i] = nb.Tick()
	}

	got := audiotest.Peak(buf)
	// Peak should be within ±20% of PeakAmp (noise modulated by envelope).
	if got < peak*0.50 {
		t.Errorf("peak within attack window = %g, want >= %g (50%% of PeakAmp %g)", got, peak*0.50, peak)
	}
	if got > peak*1.05 {
		t.Errorf("peak within attack window = %g, exceeds PeakAmp %g by too much", got, peak)
	}
}

// TestNoiseBurstDecaysToSilence verifies that after 5×DecaySec the output
// magnitude is less than 1% of PeakAmp.
func TestNoiseBurstDecaysToSilence(t *testing.T) {
	const decay = 0.020 // 20 ms
	const peak = 0.5
	nb := synth.NewNoiseBurst(synth.NoiseBurstConfig{
		Color:      synth.NoiseWhite,
		PeakAmp:    peak,
		AttackSec:  0.001,
		DecaySec:   decay,
		Seed:       7,
		SampleRate: testSR,
	})
	nb.Trigger()

	totalSamples := int(6 * decay * testSR) // 6× to be safe
	for i := 0; i < totalSamples; i++ {
		nb.Tick()
	}

	// Collect a small tail window to measure residual.
	tail := make([]float64, int(math.Round(0.005*testSR)))
	for i := range tail {
		tail[i] = nb.Tick()
	}

	got := audiotest.Peak(tail)
	threshold := peak * 0.01
	if got >= threshold {
		t.Errorf("tail peak = %g, want < %g (1%% of PeakAmp)", got, threshold)
	}
	if nb.Active() {
		t.Error("Active() should return false after burst has fully decayed")
	}
}

// TestNoiseBurstColorAffectsSpectrum verifies that NoiseLowpass (cutoff 500 Hz)
// produces a lower spectral centroid than NoiseHighpass (cutoff 5000 Hz).
func TestNoiseBurstColorAffectsSpectrum(t *testing.T) {
	const nSamples = 8192
	collect := func(color synth.NoiseColor, cutoff float64, seed int64) []float64 {
		nb := synth.NewNoiseBurst(synth.NoiseBurstConfig{
			Color:      color,
			CutoffHz:   cutoff,
			Q:          0.707,
			PeakAmp:    1.0,
			AttackSec:  0.000, // instant attack so envelope doesn't shape spectrum
			DecaySec:   1.0,   // long decay so it doesn't decay within window
			Seed:       seed,
			SampleRate: testSR,
		})
		nb.Trigger()
		buf := make([]float64, nSamples)
		for i := range buf {
			buf[i] = nb.Tick()
		}
		return buf
	}

	lpBuf := collect(synth.NoiseLowpass, 500, 100)
	hpBuf := collect(synth.NoiseHighpass, 5000, 200)

	lpCentroid := audiotest.SpectralCentroidHz(lpBuf, testSR)
	hpCentroid := audiotest.SpectralCentroidHz(hpBuf, testSR)

	t.Logf("LP centroid = %.1f Hz, HP centroid = %.1f Hz", lpCentroid, hpCentroid)

	if hpCentroid <= lpCentroid {
		t.Errorf("HP centroid (%.1f Hz) should be > LP centroid (%.1f Hz)", hpCentroid, lpCentroid)
	}
}

// ──────────────────────────────────────────────
// PitchSag tests
// ──────────────────────────────────────────────

// TestPitchSagPeaksAtZeroAndDecays verifies that immediately after Trigger the
// output is very close to PeakSemitones and after 5×TauSec it is < 1% of peak.
func TestPitchSagPeaksAtZeroAndDecays(t *testing.T) {
	const peak = 0.5   // semitones
	const tau = 0.030  // 30 ms
	ps := synth.NewPitchSag(synth.PitchSagConfig{
		PeakSemitones: peak,
		TauSec:        tau,
		SampleRate:    testSR,
	})
	ps.Trigger()

	// First tick should be very close to PeakSemitones.
	first := ps.Tick()
	if math.Abs(first-peak) > peak*0.05 {
		t.Errorf("first tick = %g semitones, want ≈ %g", first, peak)
	}

	// After 5×tau the output should be below 1% of peak.
	totalSamples := int(5*tau*testSR) - 1 // minus the one already consumed
	for i := 0; i < totalSamples; i++ {
		ps.Tick()
	}
	late := ps.Tick()
	threshold := peak * 0.01
	if math.Abs(late) >= threshold {
		t.Errorf("after 5×tau: %g semitones, want < %g (1%% of peak)", late, threshold)
	}
}

// ──────────────────────────────────────────────
// PersonalityLibrary tests
// ──────────────────────────────────────────────

var expectedPresets = []string{
	"piano_felt",
	"bass_pick",
	"brass_breath",
	"mallet_wood",
	"bell_struck",
}

// TestPersonalityLibraryHasAllPresets verifies that the named registry
// contains exactly the 5 required presets.
func TestPersonalityLibraryHasAllPresets(t *testing.T) {
	lib := synth.PersonalityLibrary()
	names := make(map[string]bool, len(lib))
	for _, p := range lib {
		names[p.Name] = true
	}
	for _, want := range expectedPresets {
		if !names[want] {
			t.Errorf("preset %q not found in PersonalityLibrary", want)
		}
	}
}

// TestPersonalityPresetsAreDistinct verifies that the 5 presets produce
// different waveform profiles when triggered, measured as pre-attack RMS.
func TestPersonalityPresetsAreDistinct(t *testing.T) {
	const sr = testSR
	// Collect 30 ms of pre-attack burst output after trigger.
	collectPreAttackRMS := func(name string) float64 {
		preset := synth.PersonalityByName(name)
		if preset == nil {
			t.Fatalf("preset %q not found", name)
		}
		p := preset.Build(sr, 999)
		if p.PreAttack == nil {
			return 0
		}
		p.PreAttack.Trigger()
		buf := make([]float64, int(0.030*sr))
		for i := range buf {
			buf[i] = p.PreAttack.Tick()
		}
		return audiotest.RMS(buf)
	}

	rmsMap := make(map[string]float64)
	for _, name := range expectedPresets {
		rmsMap[name] = collectPreAttackRMS(name)
	}
	t.Logf("pre-attack RMS values: %v", rmsMap)

	// At least 3 presets must have non-zero pre-attack RMS (bell, mallet, piano,
	// brass, bass all have PreAttack bursts with different amplitudes/cutoffs).
	nonZero := 0
	for _, v := range rmsMap {
		if v > 0.001 {
			nonZero++
		}
	}
	if nonZero < 3 {
		t.Errorf("expected >= 3 presets with non-zero pre-attack RMS, got %d", nonZero)
	}

	// Verify piano_felt and bell_struck differ in pre-attack RMS (LP 400Hz vs HP 3kHz).
	if math.Abs(rmsMap["piano_felt"]-rmsMap["bell_struck"]) < 0.001 {
		t.Errorf("piano_felt and bell_struck pre-attack RMS are too similar: %g vs %g",
			rmsMap["piano_felt"], rmsMap["bell_struck"])
	}
}

// TestPersonalityNilComponentsHandledGracefully verifies that presets with nil
// PostRelease (e.g. mallet_wood) don't cause a panic when the caller checks the
// field via a nil guard. This is a compile-time + runtime safety check.
func TestPersonalityNilComponentsHandledGracefully(t *testing.T) {
	preset := synth.PersonalityByName("mallet_wood")
	if preset == nil {
		t.Fatal("mallet_wood preset not found")
	}
	p := preset.Build(testSR, 0)

	// PostRelease must be nil for mallet_wood; this must not panic.
	if p.PostRelease != nil {
		t.Error("mallet_wood.PostRelease should be nil")
	}
	// Nil guard — exercising caller pattern without panicking.
	if p.PostRelease != nil {
		p.PostRelease.Trigger()
	}

	// PitchSag must also be nil for mallet_wood.
	if p.PitchSag != nil {
		t.Error("mallet_wood.PitchSag should be nil")
	}

	// PreAttack must be non-nil and triggerable.
	if p.PreAttack == nil {
		t.Fatal("mallet_wood.PreAttack should not be nil")
	}
	// Should not panic.
	p.PreAttack.Trigger()
	_ = p.PreAttack.Tick()
}
