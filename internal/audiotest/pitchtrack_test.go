// internal/audiotest/pitchtrack_test.go
package audiotest

import (
	"math"
	"testing"
)

func TestPitchTrackOnSteadyToneReportsZeroCentsDrift(t *testing.T) {
	s := Sine(440, 1.0, 44100, 2*44100)
	cents := PitchTrack(s, 44100, 440)
	if len(cents) < 100 {
		t.Fatalf("too few zero-crossings: %d", len(cents))
	}
	depth := ModulationDepthCents(cents)
	if depth > 2 {
		t.Fatalf("steady tone modulation depth = %.2f cents, want < 2", depth)
	}
}

func TestPitchTrackOnModulatedToneRecoversDepthAndRate(t *testing.T) {
	const depthCents = 15.0
	const rateHz = 0.7
	const seconds = 8.0
	s := ModulatedSine(440, depthCents, rateHz, 44100, int(seconds*44100))
	cents := PitchTrack(s, 44100, 440)
	gotDepth := ModulationDepthCents(cents)
	if math.Abs(gotDepth-depthCents) > 1.0 {
		t.Fatalf("recovered depth = %.2f cents, want %.2f ± 1.0", gotDepth, depthCents)
	}
	perSecond := float64(len(cents)) / seconds
	gotRate := ModulationRateHz(cents, perSecond)
	if math.Abs(gotRate-rateHz) > 0.05 {
		t.Fatalf("recovered rate = %.3f Hz, want %.3f ± 0.05", gotRate, rateHz)
	}
}

func TestAssertPitchModulationCentsPasses(t *testing.T) {
	s := ModulatedSine(440, 15, 0.7, 44100, 8*44100)
	AssertPitchModulationCents(t, s, 44100, 440, 15, 0.7, 1.0, 0.05)
}

func TestAssertPitchModulationCentsFailsOnWrongDepth(t *testing.T) {
	s := ModulatedSine(440, 5, 0.7, 44100, 8*44100)
	stub := &testing.T{}
	AssertPitchModulationCents(stub, s, 44100, 440, 15, 0.7, 1.0, 0.05)
	if !stub.Failed() {
		t.Fatal("expected assertion to fail when depth doesn't match")
	}
}
