package track

import (
	"math"
	"testing"
)

func cloneEvents(events []NoteEvent) []NoteEvent {
	out := make([]NoteEvent, len(events))
	copy(out, events)
	return out
}

func TestHumanize_DeterministicWithSeed(t *testing.T) {
	base := []NoteEvent{
		{Beat: 1.0, Pitch: "1", Dur: 0.5, Vel: 90},
		{Beat: 1.5, Pitch: "3", Dur: 0.5, Vel: 80},
		{Beat: 2.0, Pitch: "5", Dur: 0.5, Vel: 70},
		{Beat: 2.5, Pitch: "7", Dur: 0.5, Vel: 60},
	}
	spec := HumanizeSpec{TimingMs: 10, Velocity: 6, Accent: "clean"}
	a := Humanize(cloneEvents(base), spec, 42, 16, 120)
	b := Humanize(cloneEvents(base), spec, 42, 16, 120)
	if len(a) != len(b) {
		t.Fatalf("length mismatch: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Beat != b[i].Beat || a[i].Vel != b[i].Vel {
			t.Errorf("non-deterministic at %d: a=%+v b=%+v", i, a[i], b[i])
		}
	}
}

func TestHumanize_ZeroSpec_Identity(t *testing.T) {
	base := []NoteEvent{
		{Beat: 1.0, Pitch: "1", Dur: 1, Vel: 80},
		{Beat: 3.0, Pitch: "5", Dur: 1, Vel: 80},
	}
	out := Humanize(cloneEvents(base), HumanizeSpec{}, 7, 16, 120)
	for i := range out {
		if out[i].Beat != base[i].Beat || out[i].Vel != base[i].Vel {
			t.Errorf("zero spec changed event %d: got %+v want %+v", i, out[i], base[i])
		}
	}
}

func TestHumanize_TimingWithinBounds(t *testing.T) {
	base := []NoteEvent{
		{Beat: 1.0, Pitch: "1", Dur: 0.5, Vel: 80},
		{Beat: 2.5, Pitch: "5", Dur: 0.5, Vel: 80},
		{Beat: 4.0, Pitch: "7", Dur: 0.5, Vel: 80},
	}
	spec := HumanizeSpec{TimingMs: 12, Velocity: 0, Accent: "clean"}
	const bpm = 120.0
	maxBeats := 12.0 / (60000.0 / bpm)
	out := Humanize(cloneEvents(base), spec, 99, 16, bpm)
	for i, ev := range out {
		delta := math.Abs(ev.Beat - base[i].Beat)
		if delta > maxBeats+1e-9 {
			t.Errorf("event %d timing jitter %f exceeds bound %f", i, delta, maxBeats)
		}
	}
}

func TestHumanize_DillaAccent_SnareLate(t *testing.T) {
	// Two events: one snare on beat 2, one hat on beat 1.5.
	// Use seed 0 and zero jitter so we can isolate the dilla nudge.
	base := []NoteEvent{
		{Beat: 1.0, Pitch: "kick", Dur: 0.25, Vel: 100},
		{Beat: 2.0, Pitch: "snare", Dur: 0.25, Vel: 100},
		{Beat: 3.0, Pitch: "kick", Dur: 0.25, Vel: 100},
		{Beat: 4.0, Pitch: "snare", Dur: 0.25, Vel: 100},
	}
	spec := HumanizeSpec{TimingMs: 0, Velocity: 0, Accent: "dilla"}
	out := Humanize(cloneEvents(base), spec, 0, 16, 120)
	// Snare must shift later than its original beat.
	if out[1].Beat <= base[1].Beat {
		t.Errorf("snare 2 not nudged later: %f vs %f", out[1].Beat, base[1].Beat)
	}
	if out[3].Beat <= base[3].Beat {
		t.Errorf("snare 4 not nudged later: %f vs %f", out[3].Beat, base[3].Beat)
	}
	// Kick must shift earlier than its original beat.
	if out[0].Beat >= base[0].Beat && base[0].Beat > 0 {
		t.Errorf("kick 1 not nudged earlier: %f vs %f", out[0].Beat, base[0].Beat)
	}
}

func TestHumanize_PhraseArc_VelocityRises(t *testing.T) {
	// Linear ramp over 16 beats — phrase shape "arc" should give middle a
	// higher avg velocity than the endpoints.
	base := make([]NoteEvent, 17)
	for i := 0; i <= 16; i++ {
		base[i] = NoteEvent{Beat: float64(i), Pitch: "1", Dur: 0.5, Vel: 80}
	}
	spec := HumanizeSpec{TimingMs: 0, Velocity: 0, PhraseShape: "arc"}
	out := Humanize(cloneEvents(base), spec, 0, 16, 120)
	// Endpoints
	avgEnds := float64(out[0].Vel+out[16].Vel) / 2.0
	// Mid (events 7,8,9,10 around 0.5..0.6 of section)
	avgMid := float64(out[7].Vel+out[8].Vel+out[9].Vel+out[10].Vel) / 4.0
	if avgMid <= avgEnds {
		t.Errorf("arc shape did not lift mid: mid=%f ends=%f", avgMid, avgEnds)
	}
}

func TestDefaultHumanizeForFamily(t *testing.T) {
	cases := map[string]string{
		"drums": "clean",
		"bass":  "clean",
		"piano": "clean",
		"lead":  "phrase_arc",
		"pad":   "clean",
	}
	for fam, want := range cases {
		got := DefaultHumanizeForFamily(fam)
		if got.Accent != want {
			t.Errorf("family %q: accent=%q want %q", fam, got.Accent, want)
		}
		if got.TimingMs <= 0 {
			t.Errorf("family %q: timing_ms=%f want positive", fam, got.TimingMs)
		}
	}
}
