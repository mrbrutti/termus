package track

import (
	"strings"
)

// SP19-C pickup / anacrusis support.
//
// A pickup (anacrusis) is one or more "leading-in" notes that prepare the
// downbeat of the NEXT section. Real music uses these constantly: the
// saxophone glides into the first chord on the "and" of beat 4 of the bar
// before the head starts.
//
// In SP19's pragmatic implementation the pickup is encoded as part of the
// CURRENT section's tail. When Section N declares pickup_beats > 0 the engine
// overlays pickup events on Section N's PickupRole at the end of its event
// timeline. Conceptually this is "the last N beats of Section N lead into
// Section N+1." The cross-section visual is preserved because seamless
// playback (SP17) joins the two sections without a gap.
//
// Section.PickupBeats names the size of the pickup window (1..4 beats).
// Section.PickupRole selects the role that plays it (default: first lead
// role available). Section.PickupMotif optionally overrides the default
// stepwise ascent pattern.

// pickupSpec is the resolved per-incoming-section pickup directive.
type pickupSpec struct {
	Beats float64 // how many beats of pickup
	Role  string  // role that plays it (resolved)
	Motif string  // optional explicit pattern
}

// resolvePickupSpec returns the resolved pickup spec for a section, or
// zero-value (Beats == 0) when no pickup is requested.
func resolvePickupSpec(section Section, eventsByRole map[string][]NoteEvent) pickupSpec {
	beats := float64(section.PickupBeats)
	if beats <= 0 {
		return pickupSpec{}
	}
	if beats > 4 {
		beats = 4
	}
	role := strings.TrimSpace(section.PickupRole)
	if role == "" {
		role = pickLeadRoleName(eventsByRole)
	}
	if role == "" {
		return pickupSpec{}
	}
	return pickupSpec{
		Beats: beats,
		Role:  role,
		Motif: strings.TrimSpace(section.PickupMotif),
	}
}

// applyPickupToSectionTail overlays pickup events on this section's PickupRole
// at the very end of its event timeline. totalBeats is the section's beat
// span. The pickup window lies at (totalBeats - pickupBeats + 1.0 ... totalBeats).
//
// Default motif: a stepwise approach using diatonic scale degrees that lead
// into the next section's downbeat. We use ascending degrees 5–6–7 across the
// pickup span to give a "rising into the next section" feel. The engine does
// the actual key resolution later in pitch_resolver.
func applyPickupToSectionTail(spec pickupSpec, eventsByRole map[string][]NoteEvent, totalBeats float64) {
	if spec.Beats <= 0 || spec.Role == "" || totalBeats <= 0 {
		return
	}
	startBeat := totalBeats - spec.Beats + 1.0
	if startBeat < 1.0 {
		startBeat = 1.0
	}
	tokens := defaultPickupTokens(int(spec.Beats))
	if spec.Motif != "" {
		tokens = pickupTokensFromMotif(spec.Motif, int(spec.Beats))
	}
	if len(tokens) == 0 {
		return
	}
	// Distribute tokens evenly across pickup beats. Each token lasts beat_span/len.
	span := spec.Beats
	step := span / float64(len(tokens))
	if step <= 0 {
		step = 0.5
	}
	evs := eventsByRole[spec.Role]
	// Build a fresh list of pickup events.
	pickups := make([]NoteEvent, 0, len(tokens))
	for i, tok := range tokens {
		if strings.TrimSpace(tok) == "" || tok == "." || tok == "-" || tok == "r" {
			continue
		}
		pickups = append(pickups, NoteEvent{
			Beat:  startBeat + float64(i)*step,
			Pitch: tok,
			Dur:   step * 0.9,
			Vel:   88,
			Art:   "legato",
		})
	}
	if len(pickups) == 0 {
		return
	}
	// Remove any existing events on the pickup role within the pickup
	// window so the anacrusis isn't crowded.
	endBeat := totalBeats + 1.0
	filtered := make([]NoteEvent, 0, len(evs))
	for _, e := range evs {
		if e.Beat >= startBeat && e.Beat < endBeat {
			continue
		}
		filtered = append(filtered, e)
	}
	filtered = append(filtered, pickups...)
	eventsByRole[spec.Role] = filtered
}

// defaultPickupTokens produces a default anacrusis line that ascends through
// scale degrees 5, 6, 7 leading into the section's first chord root (degree 1).
// The returned tokens are scale-degree strings.
func defaultPickupTokens(beats int) []string {
	switch beats {
	case 1:
		return []string{"7"}
	case 2:
		return []string{"5", "7"}
	case 3:
		return []string{"5", "6", "7"}
	default:
		return []string{"5", "6", "7", "5"}
	}
}

// pickupTokensFromMotif splits an explicit motif pattern into per-step tokens.
// We split on whitespace and drop bar dividers ("|").
func pickupTokensFromMotif(motif string, beats int) []string {
	parts := strings.Fields(strings.ReplaceAll(motif, "|", " "))
	if len(parts) == 0 {
		return defaultPickupTokens(beats)
	}
	return parts
}
