package audiotest

import (
	"math"
	"testing"
)

func TestMeasureMonoOnUnitSine(t *testing.T) {
	s := Sine(1000, 1.0, 44100, 16384)
	m := MeasureMono(s, 44100)
	if math.Abs(m.RMSDb-(-3.01)) > 0.1 {
		t.Fatalf("RMSDb = %.2f, want -3.01 ± 0.1", m.RMSDb)
	}
	if math.Abs(m.PeakDb-0.0) > 0.5 {
		t.Fatalf("PeakDb = %.2f, want 0.0 ± 0.5", m.PeakDb)
	}
	if math.Abs(m.CentroidHz-1000) > 50 {
		t.Fatalf("CentroidHz = %.0f, want ~1000 ± 50", m.CentroidHz)
	}
	if m.Frames != len(s) {
		t.Fatalf("Frames = %d, want %d", m.Frames, len(s))
	}
}
