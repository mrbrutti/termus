package synth_test

import (
	"testing"

	"github.com/mrbrutti/termus/internal/synth"
)

// staticSource is a trivial ModSource that always returns a fixed value.
type staticSource struct{ val float64 }

func (s *staticSource) Value() float64 { return s.val }

// TestModMatrixAppliesAmountAndCallsDest verifies that a single route
// multiplies source value by Amount and delivers it to the destination.
func TestModMatrixAppliesAmountAndCallsDest(t *testing.T) {
	m := synth.NewModMatrix()
	var received float64
	m.AddRoute(synth.ModRoute{
		Source: &staticSource{val: 0.5},
		Dest:   func(v float64) { received = v },
		Amount: 2.0,
	})
	m.Tick()
	const want = 1.0 // 0.5 × 2.0
	if received != want {
		t.Fatalf("Dest received %g, want %g", received, want)
	}
}

// TestModMatrixMultipleRoutes verifies that two independent routes each
// deliver the correct scaled value to their respective destinations.
func TestModMatrixMultipleRoutes(t *testing.T) {
	m := synth.NewModMatrix()
	var destA, destB float64
	m.AddRoute(synth.ModRoute{
		Source: &staticSource{val: 3.0},
		Dest:   func(v float64) { destA = v },
		Amount: 2.0,
	})
	m.AddRoute(synth.ModRoute{
		Source: &staticSource{val: 0.25},
		Dest:   func(v float64) { destB = v },
		Amount: -1.0,
	})
	m.Tick()
	if destA != 6.0 {
		t.Fatalf("destA = %g, want 6.0", destA)
	}
	if destB != -0.25 {
		t.Fatalf("destB = %g, want -0.25", destB)
	}
}

// TestModMatrixEmptyTickIsNoOp verifies that Tick on an empty matrix doesn't
// panic.
func TestModMatrixEmptyTickIsNoOp(t *testing.T) {
	m := synth.NewModMatrix()
	m.Tick() // must not panic
}
