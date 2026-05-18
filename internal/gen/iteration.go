package gen

import (
	"strings"
)

// clonePlan returns a deep copy of an AuthoredTrackPlan so it can be mutated
// for per-iteration variations without affecting the base plan stored as
// basePlan on the AuthoredTrack.
func clonePlan(p AuthoredTrackPlan) AuthoredTrackPlan {
	out := p
	if len(p.ChordSpans) > 0 {
		out.ChordSpans = append([]AuthoredChordSpan(nil), p.ChordSpans...)
	}
	if len(p.PhraseSpans) > 0 {
		out.PhraseSpans = append([]AuthoredPhraseSpan(nil), p.PhraseSpans...)
	}
	if len(p.Tracks) > 0 {
		out.Tracks = make([]AuthoredRenderTrack, len(p.Tracks))
		for i, t := range p.Tracks {
			out.Tracks[i] = t
			if len(t.Tone) > 0 {
				out.Tracks[i].Tone = append([]string(nil), t.Tone...)
			}
			if len(t.Notes) > 0 {
				out.Tracks[i].Notes = append([]int(nil), t.Notes...)
			}
			if len(t.VelocityPattern) > 0 {
				out.Tracks[i].VelocityPattern = append([]int32(nil), t.VelocityPattern...)
			}
			if len(t.TimingOffsets) > 0 {
				out.Tracks[i].TimingOffsets = append([]float64(nil), t.TimingOffsets...)
			}
			if len(t.GatePattern) > 0 {
				out.Tracks[i].GatePattern = append([]float64(nil), t.GatePattern...)
			}
		}
	}
	if len(p.Automation) > 0 {
		out.Automation = make([]AuthoredAutomationLane, len(p.Automation))
		for i, a := range p.Automation {
			out.Automation[i] = a
			if len(a.Breakpoints) > 0 {
				out.Automation[i].Breakpoints = append([][2]float64(nil), a.Breakpoints...)
			}
		}
	}
	if len(p.RoleReverb) > 0 {
		out.RoleReverb = make(map[string]AuthoredRoleReverb, len(p.RoleReverb))
		for k, v := range p.RoleReverb {
			out.RoleReverb[k] = v
		}
	}
	if len(p.Textures) > 0 {
		out.Textures = append([]AuthoredTexture(nil), p.Textures...)
	}
	return out
}

// mutatePlanForIteration returns a copy of the plan with iteration-specific
// transformations applied:
//
//	iter == 0: unchanged
//	iter == 1: drum fill probability bumped (+50%, capped at 1.0)
//	iter == 2: above + first chord voicings octave-shifted up
//	iter >= 3: rotates through 1..2 styles
//
// The transformations are intentionally subtle so the looping listener hears
// variation without the section feeling like a different piece.
func mutatePlanForIteration(base AuthoredTrackPlan, iter int) AuthoredTrackPlan {
	if iter <= 0 {
		return clonePlan(base)
	}
	out := clonePlan(base)
	style := iter % 3 // 0, 1, 2

	for i := range out.Tracks {
		t := &out.Tracks[i]
		family := strings.ToLower(strings.TrimSpace(t.Family))
		name := strings.ToLower(strings.TrimSpace(t.Name))

		// Drum / percussion families: bump fill probability.
		if family == "drums" || family == "percussion" || strings.Contains(name, "hat") ||
			strings.Contains(name, "kick") || strings.Contains(name, "snare") ||
			strings.Contains(name, "ride") || strings.Contains(name, "perc") {
			// Multiply by 1.5 (capped at 1.0). When FireProbability is 0
			// we treat it as 1.0 (always fire).
			fp := t.FireProbability
			if fp <= 0 {
				fp = 1.0
			}
			bumped := fp * 1.5
			if bumped > 1.0 {
				bumped = 1.0
			}
			t.FireProbability = bumped
			// Add a small overall velocity bump (each iteration adds 3 to
			// the base velocity pattern offsets).
			if style == 2 {
				for j := range t.VelocityPattern {
					nv := t.VelocityPattern[j] + 3
					if nv > 12 {
						nv = 12
					}
					t.VelocityPattern[j] = nv
				}
			}
		}

		// Comping / pad / piano roles: alternate voicing inversion across
		// iterations. We approximate "rootless_a/rootless_b" inversion by
		// octave-shifting alternating notes within voicings.
		isHarmonic := family == "piano" || family == "acoustic_piano" || family == "electric_piano" ||
			family == "rhodes" || family == "pad" || family == "organ" ||
			strings.Contains(name, "rhodes") || strings.Contains(name, "keys") || strings.Contains(name, "comp") || strings.Contains(name, "pad")
		if isHarmonic && style != 0 && len(t.Notes) > 0 {
			// Style 1: drop the lowest note an octave (drop-2 substitute).
			// Style 2: lift it back / raise top note an octave.
			shift := 12
			if style == 2 {
				shift = -12
			}
			// Apply to every Nth note so the voicing alternation is audible
			// without flooding the register. We pick the first note in each
			// 8-slot group (one per bar at default 8 slots/bar).
			for j := 0; j < len(t.Notes); j++ {
				if j%4 == 0 {
					t.Notes[j] += shift
				}
			}
		}

		// Sub-bass role: activate at iter >= 2 (style != 0). We don't add
		// new tracks here, but we boost prominence of any role whose name
		// contains "sub".
		if (style == 2) && strings.Contains(name, "sub") {
			for j := range t.VelocityPattern {
				nv := t.VelocityPattern[j] + 6
				if nv > 18 {
					nv = 18
				}
				t.VelocityPattern[j] = nv
			}
		}
	}
	return out
}
