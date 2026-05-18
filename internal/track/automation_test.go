package track

import (
	"math"
	"testing"
)

// TestAutomationLane_NoBreakpoints verifies that a lane with no breakpoints
// returns 0 at any progress.
func TestAutomationLane_NoBreakpoints(t *testing.T) {
	lane := AutomationLane{Param: "cutoff"}
	for _, p := range []float64{0, 0.5, 1.0} {
		if got := lane.ValueAt(p); got != 0 {
			t.Fatalf("ValueAt(%v) = %v, want 0", p, got)
		}
	}
}

// TestAutomationLane_SingleBreakpoint verifies that a single breakpoint
// returns its value at all progress values.
func TestAutomationLane_SingleBreakpoint(t *testing.T) {
	lane := AutomationLane{
		Param:       "pan",
		Breakpoints: []Bkpt{{AtPercent: 50, Value: 0.75}},
	}
	for _, p := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
		if got := lane.ValueAt(p); got != 0.75 {
			t.Fatalf("ValueAt(%v) = %v, want 0.75", p, got)
		}
	}
}

// TestAutomationLane_Interpolates verifies linear interpolation between
// breakpoints: [(0%,0), (50%,1), (100%,0)] → 25% gives 0.5, 75% gives 0.5.
func TestAutomationLane_Interpolates(t *testing.T) {
	lane := AutomationLane{
		Param: "expression",
		Breakpoints: []Bkpt{
			{AtPercent: 0, Value: 0},
			{AtPercent: 50, Value: 1},
			{AtPercent: 100, Value: 0},
		},
	}
	const eps = 1e-9
	cases := []struct {
		progress float64
		want     float64
	}{
		{0.25, 0.5},
		{0.75, 0.5},
		{0.0, 0.0},
		{0.5, 1.0},
		{1.0, 0.0},
	}
	for _, c := range cases {
		got := lane.ValueAt(c.progress)
		if math.Abs(got-c.want) > eps {
			t.Fatalf("ValueAt(%v) = %v, want %v", c.progress, got, c.want)
		}
	}
}

// TestAutomationLane_ClampsToEnds verifies that progress < 0 returns the
// first breakpoint's value and progress > 1 returns the last.
func TestAutomationLane_ClampsToEnds(t *testing.T) {
	lane := AutomationLane{
		Param: "cutoff",
		Breakpoints: []Bkpt{
			{AtPercent: 0, Value: 0.2},
			{AtPercent: 100, Value: 0.8},
		},
	}
	if got := lane.ValueAt(-0.5); got != 0.2 {
		t.Fatalf("ValueAt(-0.5) = %v, want 0.2", got)
	}
	if got := lane.ValueAt(1.5); got != 0.8 {
		t.Fatalf("ValueAt(1.5) = %v, want 0.8", got)
	}
}
