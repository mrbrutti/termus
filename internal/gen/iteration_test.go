package gen

import "testing"

// SP19-B: mutatePlanForIteration is a no-op for iteration 0.
func TestSP19MutateIterationZeroIsNoop(t *testing.T) {
	plan := AuthoredTrackPlan{
		Tracks: []AuthoredRenderTrack{
			{Name: "hat", Family: "drums", FireProbability: 0.5},
		},
	}
	out := mutatePlanForIteration(plan, 0)
	if out.Tracks[0].FireProbability != 0.5 {
		t.Fatalf("expected iter=0 to leave plan unchanged, got %v", out.Tracks[0].FireProbability)
	}
}

// SP19-B: iter>=1 bumps drum fire probability.
func TestSP19MutateIterationBumpsDrumFill(t *testing.T) {
	plan := AuthoredTrackPlan{
		Tracks: []AuthoredRenderTrack{
			{Name: "hat", Family: "drums", FireProbability: 0.5},
		},
	}
	out := mutatePlanForIteration(plan, 1)
	if out.Tracks[0].FireProbability <= 0.5 {
		t.Fatalf("expected iter=1 to bump fire probability; got %v", out.Tracks[0].FireProbability)
	}
}

// SP19-B: iteration mutation alternates voicing on harmonic tracks.
func TestSP19MutateIterationOctaveShiftsRhodes(t *testing.T) {
	plan := AuthoredTrackPlan{
		Tracks: []AuthoredRenderTrack{
			{Name: "rhodes", Family: "rhodes", Notes: []int{60, 64, 67, 72, 60, 64, 67, 72}},
		},
	}
	out := mutatePlanForIteration(plan, 1)
	// Style 1 = iter%3==1 → shift +12 on every 4th index.
	// Index 0 should shift by +12.
	if out.Tracks[0].Notes[0] == 60 {
		t.Fatalf("expected iter=1 to octave-shift the first voicing note; notes=%v", out.Tracks[0].Notes)
	}
	// Base plan must be untouched.
	if plan.Tracks[0].Notes[0] != 60 {
		t.Fatalf("expected base plan untouched; notes=%v", plan.Tracks[0].Notes)
	}
}
