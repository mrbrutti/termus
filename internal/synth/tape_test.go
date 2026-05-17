// internal/synth/tape_test.go
package synth_test

import (
	"testing"

	"github.com/mrbrutti/termus/internal/audiotest"
	"github.com/mrbrutti/termus/internal/synth"
)

// Constructs a stereo WowFlutter with the given params and processes a long
// sine through it (mono → stereo input). Returns the left channel as a flat
// mono buffer for pitch analysis.
func runWowFlutterMono(t *testing.T, wf *synth.WowFlutter, freqHz, seconds float64) []float64 {
	t.Helper()
	const sr = 44100.0
	in := audiotest.Sine(freqHz, 0.8, sr, int(seconds*sr))
	outL := make([]float64, len(in))
	for i, s := range in {
		l, _ := wf.Tick(s, s)
		outL[i] = l
	}
	return outL
}

func TestWowFlutterIdentityWhenDepthsZero(t *testing.T) {
	wf := synth.NewWowFlutter(44100, synth.WowFlutterConfig{
		WowRateHz:         0.7,
		WowDepthCents:     0,
		FlutterRateHz:     6.0,
		FlutterDepthCents: 0,
	})
	in := audiotest.Sine(440, 0.8, 44100, 44100)
	for i, s := range in {
		l, r := wf.Tick(s, s)
		// Allow tiny drift from fractional-delay-line buffering startup;
		// after the first 256 samples the output should track the input
		// within floating-point precision.
		if i < 256 {
			continue
		}
		if absf(l-s) > 1e-6 || absf(r-s) > 1e-6 {
			t.Fatalf("zero-depth WowFlutter must pass input through; sample %d: in=%g L=%g R=%g", i, s, l, r)
		}
	}
}

func TestWowFlutterProducesExpectedWowDepth(t *testing.T) {
	wf := synth.NewWowFlutter(44100, synth.WowFlutterConfig{
		WowRateHz:         0.7,
		WowDepthCents:     20,
		FlutterRateHz:     6.0,
		FlutterDepthCents: 0, // wow only, for clean depth/rate recovery
	})
	out := runWowFlutterMono(t, wf, 440, 8.0)
	// Wow only: depth = 20 cents, rate = 0.7 Hz
	audiotest.AssertPitchModulationCents(t, out, 44100, 440, 20, 0.7, 2.0, 0.1)
}

func TestWowFlutterFlutterIsFasterThanWow(t *testing.T) {
	// Flutter only configuration — should be detectable as a higher rate
	// than the wow-only configuration above.
	wf := synth.NewWowFlutter(44100, synth.WowFlutterConfig{
		WowRateHz:         0.7,
		WowDepthCents:     0,
		FlutterRateHz:     6.0,
		FlutterDepthCents: 5,
	})
	out := runWowFlutterMono(t, wf, 440, 4.0)
	// Use a wide rate tolerance (autocorr struggles with short signals);
	// just verify rate is in the flutter band, not the wow band.
	cents := audiotest.PitchTrack(out, 44100, 440)
	perSecond := float64(len(cents)) / 4.0
	rate := audiotest.ModulationRateHz(cents, perSecond)
	if rate < 3.0 || rate > 10.0 {
		t.Fatalf("flutter rate = %.2f Hz, want in [3, 10]", rate)
	}
	depth := audiotest.ModulationDepthCents(cents)
	if depth < 3.0 || depth > 8.0 {
		t.Fatalf("flutter depth = %.2f cents, want ~5 ± 3", depth)
	}
}

func TestWowFlutterPreservesEnergy(t *testing.T) {
	wf := synth.NewWowFlutter(44100, synth.WowFlutterConfig{
		WowRateHz:         0.7,
		WowDepthCents:     15,
		FlutterRateHz:     6.0,
		FlutterDepthCents: 3,
	})
	const sr = 44100
	in := audiotest.Sine(440, 0.8, sr, 4*sr)
	outL := make([]float64, len(in))
	outR := make([]float64, len(in))
	for i, s := range in {
		outL[i], outR[i] = wf.Tick(s, s)
	}
	// Skip warm-up. RMS should be unchanged within 0.5 dB.
	stableL := outL[2048:]
	stableR := outR[2048:]
	stableIn := in[2048:]
	audiotest.AssertRMSDB(t, stableL, audiotest.ToDB(audiotest.RMS(stableIn)), 0.5)
	audiotest.AssertRMSDB(t, stableR, audiotest.ToDB(audiotest.RMS(stableIn)), 0.5)
}

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
