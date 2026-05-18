package track

import (
	"strings"
)

// generateMotifEvents (SP18) returns NoteEvents derived from a section's
// reference into the file's MotifLibrary. Only emits events for melody-ish
// roles (lead, sax, melody, rhodes, keys, piano) and only when:
//   - Section.Motif names a motif in file.MotifLibrary
//   - The role is melodic per inferFamilyFromName / explicit Family
//
// The motif pattern is transformed according to Section.MotifTreatment using
// ApplyMotifTreatment. The resulting note tokens are distributed evenly over
// the section's beat span. The author-supplied per-role events take
// precedence on collision (handled by mergeEvents).
//
// For non-melody roles or when the motif library is absent, returns nil.
func generateMotifEvents(file *File, section Section, role Role, roleName string, _ []authoredHarmonyBar, beatsPerSection float64) []NoteEvent {
	if file == nil {
		return nil
	}
	if beatsPerSection <= 0 {
		return nil
	}
	motifName := strings.TrimSpace(section.Motif)
	if motifName == "" {
		return nil
	}
	if !isMelodyRoleForMotif(role, roleName) {
		return nil
	}
	def, ok := file.MotifLibrary[motifName]
	if !ok {
		return nil
	}
	pattern := ParseMotifPattern(def.Pattern)
	if len(pattern.Tokens) == 0 {
		return nil
	}
	treatment := strings.TrimSpace(section.MotifTreatment)
	transformed := ApplyMotifTreatment(pattern, treatment)
	// Layer the phrase-structure plan on top: if the section has a phrase
	// structure, emit one transformed pass per phrase using the phrase's
	// suggested treatment.
	plans := expandPhraseStructure(section.PhraseStructure, beatsPerSection)
	if len(plans) <= 1 {
		return motifToEventsForBeats(transformed, 1.0, beatsPerSection, 84)
	}
	out := []NoteEvent{}
	for _, p := range plans {
		// Compose treatments: section-level treatment then phrase-level.
		phraseMotif := ApplyMotifTreatment(transformed, p.MotifTreatment)
		out = append(out, motifToEventsForBeats(phraseMotif, p.StartBeat, p.Beats, 84)...)
	}
	return out
}

// isMelodyRoleForMotif returns true when the role should receive
// motif-derived events. We look at the role family and the role name.
// Bass and drum roles never get motif injection.
func isMelodyRoleForMotif(role Role, name string) bool {
	family := strings.ToLower(strings.TrimSpace(role.Family))
	switch family {
	case "drums", "percussion", "bass", "synth_bass", "pad", "drone":
		return false
	case "lead", "melody", "sax", "trumpet", "flute", "horn", "alto", "tenor", "clarinet":
		return true
	}
	// Family unspecified — try the name.
	lc := strings.ToLower(strings.TrimSpace(name))
	switch lc {
	case "lead", "melody", "sax", "trumpet", "flute", "horn", "alto", "tenor", "clarinet", "rhodes", "keys", "piano", "comp", "ep":
		return true
	}
	if strings.Contains(lc, "lead") || strings.Contains(lc, "melody") {
		return true
	}
	return false
}

// applySP18SectionTransforms applies arrangement gating, dynamic curve, and
// transition shaping in order across all roles' events for a section. Each
// transform mutates eventsByRole in place.
//
// Order matters:
//  1. Arrangement gating — drops events outside role windows, ramps fades
//  2. Dynamic curve — section-wide velocity envelope
//  3. Transition — last-bars treatment (turnaround/swell/fill/...)
//
// Arrangement runs first so role gating doesn't get undone by transitions
// adding events back. Dynamic runs before transition so the transition's
// own velocity bumps land on top of the section envelope, not the other
// way round.
func applySP18SectionTransforms(_ *File, section Section, eventsByRole map[string][]NoteEvent, beatsPerSection float64) {
	if len(eventsByRole) == 0 || beatsPerSection <= 0 {
		return
	}
	// 1. Arrangement gating.
	windows := computeArrangementWindows(section, beatsPerSection)
	if len(windows) > 0 {
		for role, evs := range eventsByRole {
			win, ok := windows[role]
			if !ok {
				continue
			}
			eventsByRole[role] = applyArrangementToEvents(evs, win)
		}
	}
	// 2. Dynamic curve.
	if strings.TrimSpace(section.DynamicCurve) != "" {
		for role, evs := range eventsByRole {
			applyDynamicCurveToEvents(evs, section.DynamicCurve, beatsPerSection)
			eventsByRole[role] = evs
		}
	}
	// 2b. Phrase-level dynamics (SP19-A). Layered on top of the section
	// envelope. Default on for sections with a PhraseStructure; opt-out via
	// `phrase_dynamics: off`.
	if phraseDynamicsEnabled(section) {
		for role, evs := range eventsByRole {
			applyPhraseDynamicsToEvents(evs, section.PhraseStructure, beatsPerSection)
			eventsByRole[role] = evs
		}
	}
	// 3. Transition.
	if strings.TrimSpace(section.TransitionToNext) != "" {
		applyTransition(TransitionStyle(strings.ToLower(strings.TrimSpace(section.TransitionToNext))), eventsByRole, beatsPerSection)
	}
	// 4. SP19-C pickup / anacrusis. Overlays pickup events on the section's
	// tail so the lead role glides into the next section's downbeat.
	if section.PickupBeats > 0 {
		spec := resolvePickupSpec(section, eventsByRole)
		applyPickupToSectionTail(spec, eventsByRole, beatsPerSection)
	}
}

// phraseDynamicsEnabled returns true when the section should receive the
// SP19-A per-phrase velocity envelope. Default: on when the section declares
// a PhraseStructure. Authors can disable with `phrase_dynamics: off`.
func phraseDynamicsEnabled(section Section) bool {
	if strings.TrimSpace(section.PhraseStructure) == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(section.PhraseDynamics)) {
	case "off", "false", "no", "disabled":
		return false
	}
	return true
}
