package track

import (
	"math"
	"sort"
	"strings"

	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/synth"
)

// generateIntentEvents (SP16) returns the events generated from the role's
// auto_voice/auto_phrase directives across the section's harmony. Each chord
// region in harmonyBars is fed to GenerateVoicing; the resulting events are
// concatenated and returned in beat order.
//
// When the role has no AutoVoice/AutoPhrase set, the function returns nil so
// the caller can skip the merge step entirely.
//
// bassPresent is true when some other role in the same section already has
// an authored bass line — used so chord voicings can omit the chord root.
func generateIntentEvents(role Role, harmonyBars []authoredHarmonyBar, section Section, bpm float64, bassPresent bool) []NoteEvent {
	autoVoice := strings.TrimSpace(role.AutoVoice)
	autoPhrase := strings.TrimSpace(role.AutoPhrase)
	if autoVoice == "" && autoPhrase == "" {
		return nil
	}
	if len(harmonyBars) == 0 {
		return nil
	}
	beatsPerSection := totalBeatsForSection(section, bpm)
	if beatsPerSection <= 0 {
		return nil
	}
	const beatsPerBar = 4.0
	out := []NoteEvent{}
	for barIdx, bar := range harmonyBars {
		if len(bar.chords) == 0 {
			continue
		}
		perChordBeats := beatsPerBar / float64(len(bar.chords))
		barStartBeat := float64(barIdx)*beatsPerBar + 1.0
		for ci, chord := range bar.chords {
			startBeat := barStartBeat + float64(ci)*perChordBeats
			if startBeat-1 >= beatsPerSection {
				break
			}
			nextLabel := nextChordLabel(harmonyBars, barIdx, ci)
			ctx := VoiceContext{
				Chord:         chord.Label,
				NextChord:     nextLabel,
				StartBeat:     startBeat,
				DurationBeats: perChordBeats,
				Tempo:         bpm,
				Register:      role.Register,
				Key:           "",
				BassPresent:   bassPresent,
			}
			if autoVoice != "" {
				out = append(out, GenerateVoicing(autoVoice, ctx)...)
			}
			// AutoPhrase is intentionally a no-op in this pass — the brief
			// lists names but the engine focuses on AutoVoice for now. We
			// still accept the field for forward compatibility.
			_ = autoPhrase
		}
	}
	// Sort by beat for stable downstream behaviour.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Beat == out[j].Beat {
			return out[i].Pitch < out[j].Pitch
		}
		return out[i].Beat < out[j].Beat
	})
	return out
}

// nextChordLabel returns the chord label following the chord at
// harmonyBars[barIdx].chords[chordIdx], wrapping across bar boundaries.
// Returns "" when the next chord can't be found.
func nextChordLabel(harmonyBars []authoredHarmonyBar, barIdx, chordIdx int) string {
	bar := harmonyBars[barIdx]
	if chordIdx+1 < len(bar.chords) {
		return bar.chords[chordIdx+1].Label
	}
	if barIdx+1 < len(harmonyBars) {
		next := harmonyBars[barIdx+1]
		if len(next.chords) > 0 {
			return next.chords[0].Label
		}
	}
	return ""
}

// mergeEvents combines generator-produced events with author-supplied
// events. The author wins on collisions: when an explicit event lands on
// the same (beat, pitch) as a generated one, the explicit event replaces
// the generated one. When the explicit event lands on a beat the generator
// produced nothing for, the explicit event is added.
//
// Beat collisions use a tolerance of 1/16th note (0.0625 beats) to absorb
// authoring imprecision. Pitch collisions are exact-string-match.
func mergeEvents(generated, authored []NoteEvent) []NoteEvent {
	if len(authored) == 0 {
		return generated
	}
	if len(generated) == 0 {
		return authored
	}
	const beatTolerance = 0.0625
	out := make([]NoteEvent, 0, len(generated)+len(authored))
	used := make([]bool, len(generated))
	for _, ev := range authored {
		// Look for a near-match in the generated list to override.
		overrode := false
		for i, gen := range generated {
			if used[i] {
				continue
			}
			if math.Abs(gen.Beat-ev.Beat) < beatTolerance && strings.EqualFold(gen.Pitch, ev.Pitch) {
				used[i] = true
				overrode = true
				break
			}
		}
		_ = overrode
		out = append(out, ev)
	}
	for i, gen := range generated {
		if used[i] {
			continue
		}
		out = append(out, gen)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Beat == out[j].Beat {
			return out[i].Pitch < out[j].Pitch
		}
		return out[i].Beat < out[j].Beat
	})
	return out
}

// hasBassRole reports whether any role in the roles map has Family=bass or
// a name that suggests a bass role. Used to flag VoiceContext.BassPresent.
func hasBassRole(roles map[string]Role) bool {
	for name, role := range roles {
		fam := strings.ToLower(strings.TrimSpace(role.Family))
		ln := strings.ToLower(strings.TrimSpace(name))
		if fam == "bass" || fam == "synth_bass" || strings.Contains(ln, "bass") {
			return true
		}
	}
	return false
}

// resolveHumanizeSpec returns the humanization spec to use for a role. If
// role.Humanize is the zero value, the family default is used.
func resolveHumanizeSpec(role Role, roleName string) HumanizeSpec {
	if !role.Humanize.IsZero() {
		return role.Humanize
	}
	family := role.Family
	if family == "" {
		family = inferFamilyFromName(roleName)
	}
	return DefaultHumanizeForFamily(family)
}

// inferFamilyFromName recognises common role names and returns a family
// hint. Used when Role.Family is empty and we still want a sensible
// humanize default.
func inferFamilyFromName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "kick", "snare", "hat", "hihat", "openhat", "ride", "crash", "clap", "rim", "tom", "perc", "drums":
		return "drums"
	case "bass", "sub", "walk":
		return "bass"
	case "keys", "rhodes", "ep", "chords", "piano", "comp":
		return "piano"
	case "lead", "sax", "alto", "tenor", "clarinet", "trumpet", "horn", "flute":
		return "lead"
	case "pad", "drone", "bed", "choir", "strings":
		return "pad"
	}
	return ""
}

// applyVoicePreset (SP16) applies the LP/HP cuts and envelope hints from a
// VoicePreset to the AuthoredRenderTrack template. Brightness CC74 is set
// to the LP-cutoff equivalent when LowCutHz or HighCutHz is provided.
// Returns the (possibly modified) template.
func applyVoicePreset(template authoredRoleTemplate, preset *synth.VoicePreset) authoredRoleTemplate {
	if preset == nil {
		return template
	}
	// Program override: the SF2 inventory loader will resolve preset.SF2PresetName
	// if the loaded font has it; in any case the FallbackProgram is a safe
	// GM number to use as the channel program.
	if preset.FallbackProgram > 0 {
		template.Program = int32(preset.FallbackProgram)
	}
	// HP/LP cuts: HighCutHz dominates Brightness CC74. LP cutoff (low pass
	// at high frequency) lowers brightness as the cutoff drops below ~5 kHz.
	if preset.HighCutHz > 0 {
		template.Brightness = synth.HzToMIDICutoff(preset.HighCutHz)
	}
	// AttackBoostDB: positive values push the channel velocity slightly up
	// (sharper transient); negative values pull it down for soft attack.
	if preset.AttackBoostDB != 0 {
		// Roughly 1 dB ≈ 5 velocity units.
		delta := int32(preset.AttackBoostDB * 5)
		template.Velocity = clampVelInt32(template.Velocity + delta)
	}
	return template
}

// applyChainSpec (SP16) applies the per-role MixChain to the AuthoredRenderTrack
// template, configuring reverb send (CC91 → template.Reverb), pan (CC10 →
// template.Pan). Compression style and tape drive are advisory hints that
// the audio engine reads via plan metadata in a future pass; today the
// SF2 engine does not have per-channel compressors. The pan offset is
// applied additively.
func applyChainSpec(template authoredRoleTemplate, chain gen.MixChain) authoredRoleTemplate {
	template.Reverb = chain.ReverbSendCC91()
	if chain.PanOffset != 0 {
		// Compose with the template's existing pan: convert template.Pan
		// (0..127) to -1..+1, add chain's offset, clamp.
		base := (float64(template.Pan) - 64.0) / 63.0
		combined := base + chain.PanOffset
		if combined < -1 {
			combined = -1
		}
		if combined > 1 {
			combined = 1
		}
		template.Pan = int32(64 + combined*63)
		if template.Pan < 0 {
			template.Pan = 0
		}
		if template.Pan > 127 {
			template.Pan = 127
		}
	}
	return template
}

// resolveMixChain merges the family default with the optional Role.Chain
// overrides. The roleName lets the function pick more specific drum
// defaults (kick/snare/hat) when the role's Family is generic "drums".
func resolveMixChain(role Role, roleName string) gen.MixChain {
	familyHint := strings.ToLower(strings.TrimSpace(role.Family))
	roleHint := strings.ToLower(strings.TrimSpace(roleName))
	// Prefer role-name specificity for drums.
	if familyHint == "drums" || familyHint == "percussion" {
		switch roleHint {
		case "kick", "snare", "hat", "hihat", "openhat", "ride", "crash", "clap":
			familyHint = roleHint
		}
	} else if familyHint == "" {
		// Try role name when family was omitted.
		if hint := inferFamilyFromName(roleName); hint != "" {
			familyHint = hint
		} else {
			familyHint = roleHint
		}
	}
	base := gen.DefaultChainForFamily(familyHint)
	return gen.MergeChain(base, role.Chain.ReverbSend, role.Chain.CompressStyle, role.Chain.TapeDriveDB, role.Chain.PanOffset)
}

// hashRoleName produces a small stable hash of a role name for seeding the
// per-role Humanize RNG. FNV-1a-style; collisions are unlikely across the
// typical 4-6 roles in a track but we don't need cryptographic strength.
func hashRoleName(name string) uint32 {
	var h uint32 = 2166136261
	for _, c := range name {
		h ^= uint32(c)
		h *= 16777619
	}
	return h
}

func clampVelInt32(v int32) int32 {
	if v < 1 {
		return 1
	}
	if v > 127 {
		return 127
	}
	return v
}
