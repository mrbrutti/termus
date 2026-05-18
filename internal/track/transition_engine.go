package track

import (
	"strings"
)

// TransitionStyle (SP18) names a way of connecting one section to the next.
// Known values: turnaround, pickup, fill, breakdown, swell. Anything else is
// a no-op.
//
// Each style operates on the last 1-4 bars of the section's NoteEvent lists.
// The transition_engine modifies events in place but never crosses the
// section boundary — by design the existing seamless-section machinery
// (SP17) handles the actual section-to-section playback continuity.
type TransitionStyle string

const (
	TransTurnaround TransitionStyle = "turnaround"
	TransPickup     TransitionStyle = "pickup"
	TransFill       TransitionStyle = "fill"
	TransBreakdown  TransitionStyle = "breakdown"
	TransSwell      TransitionStyle = "swell"
)

// applyTransition mutates the per-role event map for the section so the last
// bars behave like the named transition style.
//
//	turnaround — last 2 bars: bass/keys add a ii–V–I-style step on the
//	             penultimate beat (we just emit accent events on the chord
//	             root at beat positions characteristic of jazz turnarounds).
//	pickup     — last 1 beat of the section: melody role plays a single
//	             leading note one diatonic step below the next-section's
//	             first chord root (anacrusis).
//	fill       — last 1 bar: drums add additional snare/hat hits at every
//	             16th-note position; we replicate hi-hat events on offbeats.
//	breakdown  — last 4 bars: everything except kick + bass + pad is silenced.
//	swell      — last 2 bars: every event's velocity ramps from 80% → 110%.
//
// Bpm is provided for resolving bars → beats. totalBeats is the section's
// full beat span (4/4 assumed).
func applyTransition(style TransitionStyle, eventsByRole map[string][]NoteEvent, totalBeats float64) {
	if len(eventsByRole) == 0 {
		return
	}
	const beatsPerBar = 4.0
	switch style {
	case TransTurnaround:
		applyTurnaround(eventsByRole, totalBeats)
	case TransPickup:
		applyPickup18(eventsByRole, totalBeats)
	case TransFill:
		applyFillDrum(eventsByRole, totalBeats, beatsPerBar)
	case TransBreakdown:
		applyBreakdown(eventsByRole, totalBeats, beatsPerBar*4)
	case TransSwell:
		applySwell18(eventsByRole, totalBeats, beatsPerBar*2)
	}
}

// applyTurnaround bumps velocity of bass and chordal events in the last bar
// to telegraph the impending section change. We don't synthesize new chord
// changes — the authored harmony already drives that — but we mark the
// turnaround sonically with an accent.
func applyTurnaround(eventsByRole map[string][]NoteEvent, totalBeats float64) {
	startBeat := totalBeats - 4.0
	if startBeat < 1.0 {
		startBeat = 1.0
	}
	for role, evs := range eventsByRole {
		if !isChordOrBassRole(role) {
			continue
		}
		for i, ev := range evs {
			if ev.Beat >= startBeat {
				v := ev.Vel
				if v == 0 {
					v = 80
				}
				v = int(float64(v) * 1.12)
				if v > 127 {
					v = 127
				}
				evs[i].Vel = v
				if evs[i].Art == "" {
					evs[i].Art = "accent"
				}
			}
		}
		eventsByRole[role] = evs
	}
}

// applyPickup18 adds a single anticipation event at the last 16th of the
// section. The anticipation lands on whatever role is the "lead" role; if
// none is found, the first melody-like role gets it.
func applyPickup18(eventsByRole map[string][]NoteEvent, totalBeats float64) {
	if totalBeats <= 1.0 {
		return
	}
	pickupBeat := totalBeats - 0.5
	role := pickLeadRoleName(eventsByRole)
	if role == "" {
		return
	}
	evs := eventsByRole[role]
	evs = append(evs, NoteEvent{
		Beat:  pickupBeat,
		Pitch: "7",
		Dur:   0.5,
		Vel:   90,
		Art:   "legato",
	})
	eventsByRole[role] = evs
}

// applyFillDrum injects extra snare/hat events in the last bar.
func applyFillDrum(eventsByRole map[string][]NoteEvent, totalBeats, beatsPerBar float64) {
	if totalBeats < beatsPerBar {
		return
	}
	fillStart := totalBeats - beatsPerBar + 1.0
	// Snare 16th-note triplet build in the last half-bar.
	if evs, ok := findDrumRole(eventsByRole, "snare"); ok {
		for i := 0; i < 4; i++ {
			beat := fillStart + 2.0 + float64(i)*0.5
			if beat >= totalBeats+1.0 {
				break
			}
			evs = append(evs, NoteEvent{
				Beat:  beat,
				Pitch: "",
				Dur:   0.2,
				Vel:   88 + i*4,
				Art:   "accent",
			})
		}
		setDrumRole(eventsByRole, "snare", evs)
	}
	// Hat: light closed-hat 16ths throughout the fill bar.
	if evs, ok := findDrumRole(eventsByRole, "hat"); ok {
		for i := 0; i < 8; i++ {
			beat := fillStart + float64(i)*0.5
			if beat >= totalBeats+1.0 {
				break
			}
			evs = append(evs, NoteEvent{
				Beat:  beat,
				Pitch: "",
				Dur:   0.1,
				Vel:   64,
			})
		}
		setDrumRole(eventsByRole, "hat", evs)
	}
}

// applyBreakdown silences melodic/chordal events in the last bars; kick,
// bass, and pad pass through. fadeBeats is the size of the breakdown window.
func applyBreakdown(eventsByRole map[string][]NoteEvent, totalBeats, fadeBeats float64) {
	start := totalBeats - fadeBeats
	if start < 1.0 {
		start = 1.0
	}
	for role, evs := range eventsByRole {
		if isKeepInBreakdown(role) {
			continue
		}
		filtered := make([]NoteEvent, 0, len(evs))
		for _, ev := range evs {
			if ev.Beat >= start && ev.Beat < totalBeats+1.0 {
				continue
			}
			filtered = append(filtered, ev)
		}
		eventsByRole[role] = filtered
	}
}

// applySwell18 ramps velocity 80% → 110% over the last `fadeBeats` beats
// for every role.
func applySwell18(eventsByRole map[string][]NoteEvent, totalBeats, fadeBeats float64) {
	start := totalBeats - fadeBeats
	if start < 1.0 {
		start = 1.0
	}
	for role, evs := range eventsByRole {
		for i, ev := range evs {
			if ev.Beat < start {
				continue
			}
			ratio := (ev.Beat - start) / fadeBeats
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			scale := 0.85 + 0.30*ratio
			v := ev.Vel
			if v == 0 {
				v = 80
			}
			v = int(float64(v) * scale)
			if v > 127 {
				v = 127
			}
			if v < 1 {
				v = 1
			}
			evs[i].Vel = v
		}
		eventsByRole[role] = evs
	}
}

func isChordOrBassRole(name string) bool {
	lc := strings.ToLower(name)
	if strings.Contains(lc, "bass") {
		return true
	}
	switch lc {
	case "rhodes", "keys", "piano", "ep", "comp", "guitar":
		return true
	}
	return false
}

func isKeepInBreakdown(name string) bool {
	lc := strings.ToLower(name)
	if strings.Contains(lc, "kick") || strings.Contains(lc, "bass") || strings.Contains(lc, "pad") || strings.Contains(lc, "drone") {
		return true
	}
	return false
}

func pickLeadRoleName(eventsByRole map[string][]NoteEvent) string {
	candidates := []string{"lead", "melody", "sax", "trumpet", "alto", "tenor", "flute", "clarinet", "horn", "rhodes", "keys", "piano"}
	for _, c := range candidates {
		if _, ok := eventsByRole[c]; ok {
			return c
		}
		for k := range eventsByRole {
			if strings.EqualFold(k, c) {
				return k
			}
		}
	}
	// Fallback: first non-drum, non-bass role.
	for k := range eventsByRole {
		lk := strings.ToLower(k)
		if strings.Contains(lk, "kick") || strings.Contains(lk, "snare") || strings.Contains(lk, "hat") {
			continue
		}
		if strings.Contains(lk, "bass") || strings.Contains(lk, "pad") || strings.Contains(lk, "drone") {
			continue
		}
		return k
	}
	return ""
}

func findDrumRole(eventsByRole map[string][]NoteEvent, kind string) ([]NoteEvent, bool) {
	if evs, ok := eventsByRole[kind]; ok {
		return evs, true
	}
	for k, evs := range eventsByRole {
		if strings.HasPrefix(strings.ToLower(k), kind) {
			return evs, true
		}
	}
	return nil, false
}

func setDrumRole(eventsByRole map[string][]NoteEvent, kind string, evs []NoteEvent) {
	if _, ok := eventsByRole[kind]; ok {
		eventsByRole[kind] = evs
		return
	}
	for k := range eventsByRole {
		if strings.HasPrefix(strings.ToLower(k), kind) {
			eventsByRole[k] = evs
			return
		}
	}
}
