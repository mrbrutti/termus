package gen

import "testing"

func TestApplyMovementBiasShapesState(t *testing.T) {
	state := LongHorizonState{
		Movement:      MovementLift,
		HarmonyFamily: "minor-haze",
		MotifFamily:   "hover",
		TextureScene:  "dusty",
		DensityBias:   -1,
		RegisterBias:  -1,
	}
	applyMovementBias(&state, "lofi")
	if state.HarmonyFamily != "modal-wander" {
		t.Fatalf("lift harmony = %q, want %q", state.HarmonyFamily, "modal-wander")
	}
	if state.DensityBias <= 0 {
		t.Fatalf("lift density = %d, want positive", state.DensityBias)
	}
	if state.RegisterBias <= 0 {
		t.Fatalf("lift register = %d, want positive", state.RegisterBias)
	}

	state = LongHorizonState{
		Movement:      MovementBreathe,
		HarmonyFamily: "dominant-chain",
		MotifFamily:   "pickup-line",
		TextureScene:  "horn-forward",
		DensityBias:   1,
		RegisterBias:  1,
	}
	applyMovementBias(&state, "jazz")
	if state.HarmonyFamily != "modal-minor" {
		t.Fatalf("breathe harmony = %q, want %q", state.HarmonyFamily, "modal-minor")
	}
	if state.DensityBias >= 0 {
		t.Fatalf("breathe density = %d, want negative", state.DensityBias)
	}
	if state.RegisterBias > 0 {
		t.Fatalf("breathe register = %d, want non-positive", state.RegisterBias)
	}
}
