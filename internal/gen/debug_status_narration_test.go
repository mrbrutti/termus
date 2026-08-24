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

// TestEpisodePlanFormStatusMultiEpisode advances well past the first episode
// and checks that FormStatus keeps agreeing with SectionAt once episodes
// have rolled over — in particular that later episodes' chains no longer
// carry the "intro" section (planSections strips it for every movement
// after the first).
func TestEpisodePlanFormStatusMultiEpisode(t *testing.T) {
	const barSamples = 1000
	rng := rand.New(rand.NewSource(1))
	plan := NewEpisodePlan(rng, barSamples, "jazz")

	var samples int64
	movement, episode, chain, idx := plan.FormStatus(samples)
	for episode < 2 {
		samples += barSamples
		if samples > 10_000_000 {
			t.Fatalf("did not reach episode 2 within search bound (last episode=%d)", episode)
		}
		movement, episode, chain, idx = plan.FormStatus(samples)
	}
	if movement == "" {
		t.Fatalf("movement not populated at episode 2")
	}

	if episode != 2 {
		t.Fatalf("episode = %d, want 2", episode)
	}
	if len(chain) == 0 {
		t.Fatalf("chain is empty at episode 2")
	}
	if chain[0] == string(FormIntro) {
		t.Fatalf("chain[0] = %q, want non-intro: episodes after the first strip the intro section", chain[0])
	}
	if idx < 0 || idx >= len(chain) {
		t.Fatalf("idx = %d out of range [0, %d)", idx, len(chain))
	}
	wantSection := string(plan.SectionAt(samples).Kind)
	if chain[idx] != wantSection {
		t.Fatalf("chain[idx] = %q, want %q (from SectionAt)", chain[idx], wantSection)
	}
}

// TestChillDebugStatusNarrationWiring checks that Chill.DebugStatus wires
// the form-plan narration fields through. Chill requires a *meltysynth.
// SoundFont to Seed (which drives real synthesis setup), but DebugStatus
// itself only reads the form/section/progression/samplesElapsed fields, so
// we construct via the exported constructor with a nil SoundFont (never
// calling Seed) and set just those fields directly — a minimal in-package
// fake rather than a full algorithm bring-up.
func TestChillDebugStatusNarrationWiring(t *testing.T) {
	const barSamples = 1000
	rng := rand.New(rand.NewSource(1))
	algo := NewChill(nil)
	algo.barSamples = barSamples
	algo.form = NewEpisodePlan(rng, barSamples, "lofi")
	// Bar 10 lands past the (fixed-length) intro section for every profile
	// used across these three wiring tests, so FormIndex points at a
	// non-zero entry in FormChain — exercising the index, not just its
	// zero value.
	algo.samplesElapsed = 10 * barSamples
	algo.section = algo.form.SectionAt(algo.samplesElapsed)
	algo.progression = []chillChord{{label: "I"}, {label: "vi"}, {label: "IV"}, {label: "V"}}

	status := algo.DebugStatus()
	if status.Movement == "" {
		t.Fatalf("Movement not populated")
	}
	if status.Episode == 0 {
		t.Fatalf("Episode not populated")
	}
	if len(status.FormChain) == 0 {
		t.Fatalf("FormChain not populated")
	}
	if status.NextChord == "" {
		t.Fatalf("NextChord not populated")
	}
	if status.FormIndex < 0 || status.FormIndex >= len(status.FormChain) {
		t.Fatalf("FormIndex = %d out of range [0, %d)", status.FormIndex, len(status.FormChain))
	}
	if status.FormChain[status.FormIndex] != status.Section {
		t.Fatalf("FormChain[FormIndex] = %q, want Section %q", status.FormChain[status.FormIndex], status.Section)
	}
}

// TestJazzDebugStatusNarrationWiring mirrors
// TestChillDebugStatusNarrationWiring for Jazz. See that test for why a nil
// SoundFont + direct field assignment is used instead of Seed.
func TestJazzDebugStatusNarrationWiring(t *testing.T) {
	const barSamples = 1000
	rng := rand.New(rand.NewSource(1))
	algo := NewJazz(nil)
	algo.barSamples = barSamples
	algo.form = NewEpisodePlan(rng, barSamples, "jazz")
	// Bar 10 lands past the (fixed-length) intro section for every profile
	// used across these three wiring tests, so FormIndex points at a
	// non-zero entry in FormChain — exercising the index, not just its
	// zero value.
	algo.samplesElapsed = 10 * barSamples
	algo.section = algo.form.SectionAt(algo.samplesElapsed)
	algo.progression = []jazzChord{{label: "iim7"}, {label: "V7"}, {label: "Imaj7"}, {label: "VI7"}}

	status := algo.DebugStatus()
	if status.Movement == "" {
		t.Fatalf("Movement not populated")
	}
	if status.Episode == 0 {
		t.Fatalf("Episode not populated")
	}
	if len(status.FormChain) == 0 {
		t.Fatalf("FormChain not populated")
	}
	if status.NextChord == "" {
		t.Fatalf("NextChord not populated")
	}
	if status.FormIndex < 0 || status.FormIndex >= len(status.FormChain) {
		t.Fatalf("FormIndex = %d out of range [0, %d)", status.FormIndex, len(status.FormChain))
	}
	if status.FormChain[status.FormIndex] != status.Section {
		t.Fatalf("FormChain[FormIndex] = %q, want Section %q", status.FormChain[status.FormIndex], status.Section)
	}
}

// TestSF2MarkovDebugStatusNarrationWiring mirrors
// TestChillDebugStatusNarrationWiring for SF2Markov. See that test for why a
// nil SoundFont + direct field assignment is used instead of Seed.
func TestSF2MarkovDebugStatusNarrationWiring(t *testing.T) {
	const barSamples = 1000
	rng := rand.New(rand.NewSource(1))
	algo := NewSF2Markov(nil)
	algo.barSamples = barSamples
	algo.form = NewEpisodePlan(rng, barSamples, "classical")
	// Bar 10 lands past the (fixed-length) intro section for every profile
	// used across these three wiring tests, so FormIndex points at a
	// non-zero entry in FormChain — exercising the index, not just its
	// zero value.
	algo.samplesElapsed = 10 * barSamples
	algo.section = algo.form.SectionAt(algo.samplesElapsed)
	algo.progression = []classicalChord{{label: "I"}, {label: "IV"}, {label: "V"}, {label: "I"}}

	status := algo.DebugStatus()
	if status.Movement == "" {
		t.Fatalf("Movement not populated")
	}
	if status.Episode == 0 {
		t.Fatalf("Episode not populated")
	}
	if len(status.FormChain) == 0 {
		t.Fatalf("FormChain not populated")
	}
	if status.NextChord == "" {
		t.Fatalf("NextChord not populated")
	}
	if status.FormIndex < 0 || status.FormIndex >= len(status.FormChain) {
		t.Fatalf("FormIndex = %d out of range [0, %d)", status.FormIndex, len(status.FormChain))
	}
	if status.FormChain[status.FormIndex] != status.Section {
		t.Fatalf("FormChain[FormIndex] = %q, want Section %q", status.FormChain[status.FormIndex], status.Section)
	}
}
