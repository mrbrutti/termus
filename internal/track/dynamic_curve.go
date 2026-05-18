package track

import (
	"math"
	"strings"
)

// DynamicCurve (SP18) is a per-section velocity envelope shape applied to
// every event in the section. The curve modulates velocity by up to ±20%
// of the base value across the section's beat span.
//
// Recognised shapes:
//
//	arc          — rise to peak at 60% of section, then fall (most musical)
//	crescendo    — linear rise from 80% to 110%
//	decrescendo  — linear fall from 110% to 80%
//	wave         — single sine cycle ±15%
//	steady       — flat (no modulation)
//
// Unknown shapes default to "steady".
type DynamicCurve string

// dynamicCurveScale returns the velocity scale factor at a given progress
// position (0..1 fraction of the way through the section).
func dynamicCurveScale(curve string, progress float64) float64 {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	switch strings.ToLower(strings.TrimSpace(curve)) {
	case "arc":
		// Peak at 0.6, base 0.85, peak 1.15.
		peak := 0.6
		dist := math.Abs(progress - peak)
		// Triangular falloff scaled so distance 0.4 (one of the edges) yields 0.85.
		maxDist := math.Max(peak, 1.0-peak)
		ratio := dist / maxDist
		return 1.15 - 0.30*ratio
	case "crescendo":
		return 0.85 + 0.25*progress
	case "decrescendo":
		return 1.10 - 0.25*progress
	case "wave":
		// One full sine cycle, ±0.15.
		return 1.0 + 0.15*math.Sin(progress*math.Pi*2)
	case "steady", "":
		return 1.0
	}
	return 1.0
}

// applyDynamicCurveToEvents scales every event's velocity according to the
// section's dynamic curve. totalBeats is the section's total beat span. The
// curve is sampled at each event's beat position. Original slice is mutated
// in place.
func applyDynamicCurveToEvents(events []NoteEvent, curve string, totalBeats float64) {
	if curve == "" || len(events) == 0 || totalBeats <= 0 {
		return
	}
	for i, ev := range events {
		progress := (ev.Beat - 1.0) / totalBeats
		scale := dynamicCurveScale(curve, progress)
		v := ev.Vel
		if v == 0 {
			v = 80
		}
		nv := int(float64(v)*scale + 0.5)
		if nv < 1 {
			nv = 1
		}
		if nv > 127 {
			nv = 127
		}
		events[i].Vel = nv
	}
}

// SP19-A: phrase-level dynamic curves.
//
// Real 8-bar phrases breathe inside a section. We layer a per-phrase envelope
// on top of the section curve. Phrase positions:
//
//	first phrase   → crescendo (introduce)
//	middle phrases → arc       (rise / fall)
//	last phrase    → decrescendo (close)
//
// The phrase envelope modulates velocity by ±10% (smaller than the section
// envelope's ±20%) — multiplicative on top of whatever applyDynamicCurveToEvents
// already produced.
//
// Section.PhraseStructure drives which phrase letter each phrase carries.
// If empty / unknown, we treat the section as a single phrase and skip the
// per-phrase pass.

// phraseShapeForPosition returns the phrase-level dynamic shape for a phrase
// at position phraseIdx out of totalPhrases. First = crescendo, last =
// decrescendo, anything else = arc.
func phraseShapeForPosition(phraseIdx, totalPhrases int) string {
	if totalPhrases <= 1 {
		return "arc"
	}
	switch phraseIdx {
	case 0:
		return "crescendo"
	case totalPhrases - 1:
		return "decrescendo"
	default:
		return "arc"
	}
}

// phraseLevelScale returns the per-phrase velocity scale factor at the given
// progress (0..1) inside the phrase using the per-phrase shape. The
// modulation depth is ±10% (so values land between 0.90 and 1.10).
func phraseLevelScale(shape string, progress float64) float64 {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	switch strings.ToLower(strings.TrimSpace(shape)) {
	case "arc":
		peak := 0.55
		dist := math.Abs(progress - peak)
		maxDist := math.Max(peak, 1.0-peak)
		ratio := dist / maxDist
		return 1.10 - 0.20*ratio
	case "crescendo":
		return 0.92 + 0.16*progress
	case "decrescendo":
		return 1.08 - 0.16*progress
	case "wave":
		return 1.0 + 0.08*math.Sin(progress*math.Pi*2)
	case "steady", "":
		return 1.0
	}
	return 1.0
}

// applyPhraseDynamicsToEvents scales velocities by an additional per-phrase
// envelope. The events are bucketed into phrase plans (from expandPhraseStructure)
// and within each phrase the shape from phraseShapeForPosition is sampled at
// the event's position inside the phrase span.
//
// Mutates the slice in place. Safe to call when structureName is empty or
// totalBeats is non-positive — in that case it returns without changes.
//
// Behaviour interacts with applyDynamicCurveToEvents: that one should be
// called first (section envelope), then this one composes on top. The two
// scale factors multiply, so combined modulation is up to ±30%.
func applyPhraseDynamicsToEvents(events []NoteEvent, structureName string, totalBeats float64) {
	if len(events) == 0 || totalBeats <= 0 {
		return
	}
	plans := expandPhraseStructure(structureName, totalBeats)
	if len(plans) <= 1 {
		return
	}
	for i, ev := range events {
		// Find the phrase containing this event's beat. Events use Beat
		// 1-indexed (Beat=1.0 is first beat of section).
		var phraseIdx int = -1
		var p PhrasePlan
		for j, candidate := range plans {
			start := candidate.StartBeat
			end := candidate.StartBeat + candidate.Beats
			if ev.Beat >= start && ev.Beat < end {
				phraseIdx = j
				p = candidate
				break
			}
		}
		if phraseIdx < 0 {
			// Past the end (event beat == total+1). Use last phrase.
			phraseIdx = len(plans) - 1
			p = plans[phraseIdx]
		}
		if p.Beats <= 0 {
			continue
		}
		shape := phraseShapeForPosition(phraseIdx, len(plans))
		progress := (ev.Beat - p.StartBeat) / p.Beats
		scale := phraseLevelScale(shape, progress)
		v := ev.Vel
		if v == 0 {
			v = 80
		}
		nv := int(float64(v)*scale + 0.5)
		if nv < 1 {
			nv = 1
		}
		if nv > 127 {
			nv = 127
		}
		events[i].Vel = nv
	}
}
