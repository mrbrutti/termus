package synth

import (
	"math"
	"testing"
)

func TestEnvelopePhases(t *testing.T) {
	e := NewEnvelope(0.1, 0.1, 0.5, 0.2) // A=100ms D=100ms S=0.5 R=200ms
	e.Gate(true)

	// At t=0 envelope is 0.
	if v := e.Tick(); v > 0.05 {
		t.Fatalf("t=0: env=%g, want near 0", v)
	}
	// End of attack (≈100ms): should be near 1.
	for i := 1; i < int(0.1*SampleRate); i++ {
		e.Tick()
	}
	if v := e.Tick(); math.Abs(v-1) > 0.05 {
		t.Fatalf("end of attack: env=%g, want near 1", v)
	}
	// After decay (≈200ms total): should be near sustain 0.5.
	for i := 0; i < int(0.1*SampleRate); i++ {
		e.Tick()
	}
	if v := e.Tick(); math.Abs(v-0.5) > 0.05 {
		t.Fatalf("end of decay: env=%g, want near 0.5", v)
	}
	// Release: gate off, should decay toward 0.
	e.Gate(false)
	for i := 0; i < int(0.2*SampleRate)+10; i++ {
		e.Tick()
	}
	if v := e.Tick(); v > 0.05 {
		t.Fatalf("after release: env=%g, want near 0", v)
	}
}
