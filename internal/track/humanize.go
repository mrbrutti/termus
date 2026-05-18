package track

import (
	"math"
	"math/rand"
	"strings"
)

// DefaultHumanizeForFamily returns the genre-tuned default HumanizeSpec for
// the given Role family. The defaults err on the gentle side so any role
// receives at least some jitter — purely mechanical timing is rare in real
// performances.
//
// SP16: empty family or unknown family → drum-ish defaults (no harm done).
func DefaultHumanizeForFamily(family string) HumanizeSpec {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "drum", "drums", "percussion":
		return HumanizeSpec{TimingMs: 3, Velocity: 5, Accent: "clean"}
	case "bass":
		return HumanizeSpec{TimingMs: 5, Velocity: 6, Accent: "clean"}
	case "piano", "rhodes", "keys", "wurli":
		return HumanizeSpec{TimingMs: 6, Velocity: 8, Accent: "clean"}
	case "lead", "sax", "melody", "trumpet", "guitar":
		return HumanizeSpec{TimingMs: 10, Velocity: 10, Accent: "phrase_arc"}
	case "pad", "drone", "ambient", "strings":
		return HumanizeSpec{TimingMs: 2, Velocity: 4, Accent: "clean"}
	}
	return HumanizeSpec{TimingMs: 3, Velocity: 5, Accent: "clean"}
}

// HumanizeForRole applies Humanize then re-applies the "dilla" accent
// using the role name as the drum classifier. This is the integration entry
// point used by authored_compile.go: when authors omit the Pitch string on
// drum events (the common case), only the role name identifies the hit.
//
// Calling Humanize directly with kind information embedded in the Pitch
// string is also supported.
func HumanizeForRole(events []NoteEvent, spec HumanizeSpec, seed int64, sectionBeats, bpm float64, roleName string) []NoteEvent {
	events = Humanize(events, spec, seed, sectionBeats, bpm)
	if strings.EqualFold(strings.TrimSpace(spec.Accent), "dilla") {
		// If accent is dilla and events lack explicit pitches (drum case),
		// apply the role-name-based nudge as a second pass.
		applyDillaAccentByRole(events, roleName, bpm)
	}
	return events
}

func applyDillaAccentByRole(events []NoteEvent, roleName string, bpm float64) {
	if bpm <= 0 {
		bpm = 120
	}
	msPerBeat := 60000.0 / bpm
	kind := strings.ToLower(strings.TrimSpace(roleName))
	switch kind {
	case "snare", "clap":
		nudge := 10.0 / msPerBeat
		for i := range events {
			if classifyDrumPitch(events[i].Pitch) != "" {
				continue // already covered by pitch-classifier path
			}
			events[i].Beat += nudge
		}
	case "kick":
		nudge := 5.0 / msPerBeat
		for i := range events {
			if classifyDrumPitch(events[i].Pitch) != "" {
				continue
			}
			events[i].Beat -= nudge
			if events[i].Beat < 0 {
				events[i].Beat = 0
			}
		}
	}
}

// Humanize applies the per-role humanization spec to a sequence of NoteEvent.
// The events are mutated in place — beats may be nudged later, velocities
// jittered, accent profiles applied, and a phrase-level dynamic shape
// imposed. Returns the same slice for fluency.
//
// Determinism: all randomness uses a local rand.Rand seeded from seed; no
// global rand state is touched. Calling Humanize with the same inputs always
// produces the same outputs.
//
// bpm is required so the timing-jitter ms can be converted to beats. When
// bpm <= 0 the function uses 120 BPM as a safe fallback.
//
// sectionBeats is the total length of the section in beats; phrase-arc and
// crescendo/decrescendo shapes use this to position each event within the
// section. When sectionBeats <= 0 the function falls back to the max event
// beat so the shapes still apply over the captured range.
//
// Accent profiles ("dilla", "swing_accent", "phrase_arc", "clean") and
// PhraseShape ("crescendo", "decrescendo", "arc", "steady") are applied
// after the per-event jitter so the deterministic structure dominates the
// random noise.
func Humanize(events []NoteEvent, spec HumanizeSpec, seed int64, sectionBeats, bpm float64) []NoteEvent {
	if len(events) == 0 {
		return events
	}
	if bpm <= 0 {
		bpm = 120
	}
	if sectionBeats <= 0 {
		for _, ev := range events {
			if ev.Beat > sectionBeats {
				sectionBeats = ev.Beat
			}
		}
		if sectionBeats <= 0 {
			sectionBeats = 1
		}
	}

	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // not security-sensitive

	// Convert ms jitter to beats: beats = ms / (60000 / bpm) where 60000/bpm
	// is the ms-per-beat at the current tempo.
	msPerBeat := 60000.0 / bpm
	maxBeatJitter := 0.0
	if spec.TimingMs > 0 {
		maxBeatJitter = spec.TimingMs / msPerBeat
	}

	accent := strings.ToLower(strings.TrimSpace(spec.Accent))
	phraseShape := strings.ToLower(strings.TrimSpace(spec.PhraseShape))

	// First pass: per-event timing + velocity jitter, plus accent-profile
	// tweaks that depend on each event's role (e.g. "snare" gets late on
	// dilla). Articulation hints stay untouched.
	for i := range events {
		ev := &events[i]

		// Timing jitter.
		if maxBeatJitter > 0 {
			jitter := (rng.Float64()*2 - 1) * maxBeatJitter
			ev.Beat += jitter
			if ev.Beat < 0 {
				ev.Beat = 0
			}
		}

		// Velocity jitter.
		if spec.Velocity > 0 {
			vel := ev.Vel
			if vel <= 0 {
				vel = 80
			}
			j := rng.Intn(2*spec.Velocity+1) - spec.Velocity
			vel += j
			ev.Vel = clampVel(vel)
		}
	}

	// Second pass: accent profile. These are deterministic adjustments on
	// top of the per-event jitter and are seeded from the same RNG so they
	// remain reproducible.
	switch accent {
	case "dilla":
		applyDillaAccent(events, msPerBeat)
	case "swing_accent":
		applySwingAccent(events)
	case "phrase_arc":
		applyHumanizePhraseArc(events)
	case "clean", "":
		// no-op
	}

	// Third pass: section-wide phrase shape — adjust velocity according to
	// position within the section.
	switch phraseShape {
	case "crescendo":
		applyVelocityCurve(events, sectionBeats, func(t float64) float64 {
			// +15% from start to end.
			return 1.0 + 0.15*t
		})
	case "decrescendo":
		applyVelocityCurve(events, sectionBeats, func(t float64) float64 {
			return 1.0 + 0.15*(1.0-t)
		})
	case "arc":
		applyVelocityCurve(events, sectionBeats, func(t float64) float64 {
			// Peak at 60% through the section.
			peak := 0.6
			distance := math.Abs(t - peak)
			// Symmetric arc, +12% at the peak.
			return 1.0 + 0.12*(1.0-distance/math.Max(peak, 1.0-peak))
		})
	}

	return events
}

func clampVel(v int) int {
	switch {
	case v < 1:
		return 1
	case v > 127:
		return 127
	}
	return v
}

// applyDillaAccent applies J Dilla-style timing: snare hits land ~10ms late
// on beats 2 & 4; kick hits land ~5ms early on beat 1. The role isn't on
// the event itself but the standard drum-MIDI pitches give it away; we also
// inspect the Pitch text in case authors used named drum tokens.
func applyDillaAccent(events []NoteEvent, msPerBeat float64) {
	// 10ms late → +10/msPerBeat beats. 5ms early → -5/msPerBeat beats.
	const snareLateMs = 10.0
	const kickEarlyMs = 5.0
	snareNudge := snareLateMs / msPerBeat
	kickNudge := kickEarlyMs / msPerBeat
	for i := range events {
		ev := &events[i]
		kind := classifyDrumPitch(ev.Pitch)
		switch kind {
		case "snare":
			ev.Beat += snareNudge
		case "kick":
			ev.Beat -= kickNudge
			if ev.Beat < 0 {
				ev.Beat = 0
			}
		}
	}
}

// applySwingAccent boosts beats 1 and 3 by +6 velocity (ride hits) and trims
// off-beats by -3. Off-beats are recognised by their fractional position.
func applySwingAccent(events []NoteEvent) {
	for i := range events {
		ev := &events[i]
		// Beats are 1-indexed in NoteEvent. Beat 1, 3, 5, 7 = strong; 2,4,6,8 = weak.
		// Treat anything within 0.1 beats of an odd integer as "on beat 1/3".
		beatPos := ev.Beat
		// Fractional offset from nearest integer.
		nearest := math.Round(beatPos)
		distance := math.Abs(beatPos - nearest)
		if distance < 0.1 {
			ib := int(nearest)
			if ib%2 == 1 {
				ev.Vel = clampVel(ev.Vel + 6)
			} else {
				ev.Vel = clampVel(ev.Vel - 3)
			}
		}
	}
}

// applyHumanizePhraseArc adds +6 velocity to the first event in each 4-bar (=16-beat)
// group and -8 to the last. Useful for melody lines.
func applyHumanizePhraseArc(events []NoteEvent) {
	if len(events) == 0 {
		return
	}
	// Group events by 16-beat windows; find the earliest and latest event in
	// each window.
	type bound struct {
		first, last int
	}
	groups := map[int]*bound{}
	for i, ev := range events {
		grp := int(math.Floor(ev.Beat / 16))
		b, ok := groups[grp]
		if !ok {
			groups[grp] = &bound{first: i, last: i}
			continue
		}
		if events[b.first].Beat > ev.Beat {
			b.first = i
		}
		if events[b.last].Beat < ev.Beat {
			b.last = i
		}
	}
	for _, b := range groups {
		events[b.first].Vel = clampVel(events[b.first].Vel + 6)
		events[b.last].Vel = clampVel(events[b.last].Vel - 8)
	}
}

// applyVelocityCurve scales each event's velocity by curve(t) where t is the
// 0..1 position of the event within the section.
func applyVelocityCurve(events []NoteEvent, sectionBeats float64, curve func(float64) float64) {
	if sectionBeats <= 0 {
		return
	}
	for i := range events {
		ev := &events[i]
		t := ev.Beat / sectionBeats
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		mult := curve(t)
		vel := ev.Vel
		if vel <= 0 {
			vel = 80
		}
		ev.Vel = clampVel(int(math.Round(float64(vel) * mult)))
	}
}

// classifyDrumPitch returns "kick", "snare", "hat", "" depending on the
// event's Pitch string. Numeric pitches map via the General MIDI percussion
// table; named tokens use Termus's drum-role conventions. Empty string
// returns "" — the caller can't tell from pitch alone.
func classifyDrumPitch(pitch string) string {
	p := strings.ToLower(strings.TrimSpace(pitch))
	if p == "" {
		return ""
	}
	switch p {
	case "kick", "bd", "k":
		return "kick"
	case "snare", "sd", "sn":
		return "snare"
	case "hat", "hh", "h", "closed", "openhat":
		return "hat"
	}
	if n, ok := parsePositiveInt(p); ok {
		switch n {
		case 35, 36:
			return "kick"
		case 38, 39, 40:
			return "snare"
		case 42, 44, 46:
			return "hat"
		}
	}
	return ""
}
