package gen

import (
	"math/rand"
	"testing"
)

func TestEpisodePlanFormStatus(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	plan := NewEpisodePlan(rng, 1000, "jazz")
	movement, episode, chain, idx := plan.FormStatus(0)
	if movement != string(MovementEstablish) {
		t.Fatalf("movement = %q, want %q", movement, MovementEstablish)
	}
	if episode != 1 {
		t.Fatalf("episode = %d, want 1", episode)
	}
	if len(chain) == 0 || chain[0] != string(FormIntro) {
		t.Fatalf("chain = %v, want first section %q", chain, FormIntro)
	}
	if idx != 0 {
		t.Fatalf("idx = %d, want 0", idx)
	}
	// Move into the second section of episode 1: intro is 4 bars in the
	// jazz profile, so bar 5 (samples = 4*barSamples) is inside section A.
	_, _, _, idx2 := plan.FormStatus(4 * 1000)
	if idx2 != 1 {
		t.Fatalf("idx after intro = %d, want 1", idx2)
	}
}

func TestEpisodePlanFormStatusNilSafe(t *testing.T) {
	var plan *EpisodePlan
	movement, episode, chain, idx := plan.FormStatus(0)
	if movement != "" || episode != 0 || chain != nil || idx != 0 {
		t.Fatalf("nil plan should return zero values, got %q %d %v %d", movement, episode, chain, idx)
	}
}
