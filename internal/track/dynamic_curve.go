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
