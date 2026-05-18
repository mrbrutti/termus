package track

import (
	"math"
	"strings"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

// eventSlotsPerBeat is the grid resolution we use for explicit NoteEvent
// compilation. A grid of 16 slots per beat (64th-note resolution) accommodates
// all common rhythmic placements (8th, 16th, triplet — see below) with at
// most one slot of quantization. Sub-quantization is recovered via the
// AuthoredRenderTrack.TimingOffsets array, which adds a per-slot timing
// jitter in seconds.
//
// 16 slots per beat lands exactly on 4th, 8th, 16th, and 32nd-note positions.
// Triplet (1/12-beat) placements quantize to the nearest 1/16-beat with a
// timing offset of at most ~0.5 / (16 * BPS) seconds — well below the
// audibility threshold for typical 60–180 BPM material.
const eventSlotsPerBeat = 16

// roleEventList returns the active NoteEvent list for the given role in the
// given section. Resolution precedence (per SP14):
//
//	section.RoleEvents[roleName]  > role.Events  > nil
//
// Returning nil means: "no explicit events, fall through to the existing
// pattern / motif / algorithm path".
func roleEventList(roleName string, role Role, section Section) []NoteEvent {
	if section.RoleEvents != nil {
		if list, ok := section.RoleEvents[roleName]; ok && len(list) > 0 {
			return list
		}
	}
	if len(role.Events) > 0 {
		return role.Events
	}
	return nil
}

// compileRoleEventTrack converts a list of NoteEvents into a single
// AuthoredRenderTrack. The track uses a high-resolution slot grid
// (eventSlotsPerBeat per beat) with rests in unused slots; events are
// placed at the slot whose center time best matches the requested Beat
// position.
//
// The returned track always has its PeriodSec set to the full section
// duration; the engine cycles the Notes array once per section so the
// rhythm repeats predictably across section boundaries.
//
// When the events list is empty this returns the zero track and ok=false.
func compileRoleEventTrack(ctx authoredSectionContext, name string, role Role, events []NoteEvent, harmonyBars []authoredHarmonyBar, section Section, bpm float64) (gen.AuthoredRenderTrack, bool) {
	if len(events) == 0 {
		return gen.AuthoredRenderTrack{}, false
	}
	template := authoredTemplateFor(ctx.style, name, role)
	beatsPerSection := totalBeatsForSection(section, bpm)
	if beatsPerSection <= 0 {
		// Without a positive duration we can't lay out events deterministically.
		return gen.AuthoredRenderTrack{}, false
	}
	totalSlots := int(math.Ceil(beatsPerSection * eventSlotsPerBeat))
	if totalSlots < 1 {
		totalSlots = 1
	}

	notes := make([]int, totalSlots)
	velPattern := make([]int32, totalSlots)
	timingOff := make([]float64, totalSlots)
	gatePattern := make([]float64, totalSlots)
	for i := range notes {
		notes[i] = -1
	}

	keyStr := firstNonBlank(section.Key, ctx.style)
	if ctxKey := firstNonBlank(section.Key, ""); ctxKey != "" {
		keyStr = ctxKey
	}
	// Resolve via the file key when no section key was set.
	if strings.TrimSpace(section.Key) == "" {
		keyStr = ""
	}

	slotSec := 60.0 / (bpm * float64(eventSlotsPerBeat))
	kind := authoredRoleKind(name, role)
	isDrum := kind == "drum"

	for _, ev := range events {
		// Place beat → slot (rounded).
		beat := ev.Beat
		if beat < 1.0 {
			// Treat <1 as starting position from beat 1.
			beat = 1.0
		}
		// 0-based position within the section, in slots.
		pos := (beat - 1.0) * float64(eventSlotsPerBeat)
		slot := int(math.Round(pos))
		if slot < 0 || slot >= totalSlots {
			continue
		}
		// Sub-slot offset (in seconds) so the event fires at the exact
		// requested beat even if it doesn't land on the grid.
		residual := pos - float64(slot)
		timingOff[slot] = residual * slotSec

		// Pitch.
		var midi int
		if isDrum {
			midi = drumNoteForEvent(name, ev.Pitch)
		} else {
			chord := chordForEventBeat(harmonyBars, beat, bpm, beatsPerSection)
			midi = ResolvePitch(ev.Pitch, keyStr, chord, role.Register)
			if midi < 0 {
				// Fallback: middle C with a deterministic register shift so the
				// event remains audible but obviously off-spec for debugging.
				midi = 60
			}
		}
		notes[slot] = midi

		// Velocity + articulation.
		vel := ev.Vel
		if vel <= 0 {
			vel = 80
		}
		velOffset := int32(vel) - template.Velocity
		gate := 1.0
		switch strings.ToLower(strings.TrimSpace(ev.Art)) {
		case "ghost":
			velOffset -= 32
			gate = 0.35
		case "accent":
			velOffset += 15
		case "staccato":
			gate = 0.25
		case "tenuto":
			gate = 1.0
		case "legato":
			gate = 1.1
		}
		velPattern[slot] = velOffset

		// Duration. Convert beats → slots; gate is the proportion of the slot
		// the note holds. We translate event Dur to gate-per-slot by computing
		// the held duration in seconds and dividing by the slot duration.
		dur := ev.Dur
		if dur <= 0 {
			dur = 0.5
		}
		holdSec := dur * 60.0 / bpm
		// Apply articulation gate as a multiplier on top.
		holdSec *= gate
		gatePattern[slot] = holdSec / slotSec
		if gatePattern[slot] < 0.05 {
			gatePattern[slot] = 0.05
		}
		if gatePattern[slot] > 32.0 {
			gatePattern[slot] = 32.0
		}
	}

	out := gen.AuthoredRenderTrack{
		Name:            name,
		Family:          role.Family,
		Tone:            append([]string(nil), role.Tone...),
		Articulation:    role.Articulation,
		Register:        role.Register,
		Prominence:      role.Prominence,
		Channel:         template.Channel,
		Program:         template.Program,
		Velocity:        template.Velocity,
		Pan:             template.Pan,
		Reverb:          template.Reverb,
		Chorus:          template.Chorus,
		Brightness:      template.Brightness,
		Notes:           notes,
		VelocityPattern: velPattern,
		TimingOffsets:   timingOff,
		GatePattern:     gatePattern,
		Gate:            template.Gate,
		SwingAmount:     0, // events bypass groove swing entirely
		Legato:          template.Legato,
		TieRepeats:      template.TieRepeats,
		OverlapSec:      template.OverlapSec,
		FireProbability: 1,
	}
	return out, true
}

// totalBeatsForSection computes the total number of beats in a section
// given its duration string and the resolved tempo BPM. Returns 0 on parse
// failure.
func totalBeatsForSection(section Section, bpm float64) float64 {
	durSec := parseSectionDurationSeconds(section.Duration)
	if durSec <= 0 || bpm <= 0 {
		return 0
	}
	return durSec * bpm / 60.0
}

// parseSectionDurationSeconds parses a string like "8s", "1m30s", "2m" into
// seconds. Returns 0 on failure.
func parseSectionDurationSeconds(raw string) float64 {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d.Seconds()
}

// chordForEventBeat selects the chord active at a given beat position within
// the section. Beats are 1-indexed (1.0 = first beat).
func chordForEventBeat(harmonyBars []authoredHarmonyBar, beat, bpm, totalBeats float64) authoredChord {
	if len(harmonyBars) == 0 {
		return authoredChord{RootPC: 0, Kind: "maj", Scale: []int{0, 2, 4, 5, 7, 9, 11}}
	}
	// Assume each bar is 4 beats (4/4 throughout the codebase).
	const beatsPerBar = 4
	bar0 := int((beat - 1.0) / beatsPerBar)
	if bar0 < 0 {
		bar0 = 0
	}
	if bar0 >= len(harmonyBars) {
		bar0 = len(harmonyBars) - 1
	}
	chords := harmonyBars[bar0].chords
	if len(chords) == 0 {
		return authoredChord{RootPC: 0, Kind: "maj", Scale: []int{0, 2, 4, 5, 7, 9, 11}}
	}
	// Map position within the bar to the chord that owns it.
	posInBar := math.Mod(beat-1.0, beatsPerBar)
	perChord := float64(beatsPerBar) / float64(len(chords))
	idx := int(math.Floor(posInBar / perChord))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(chords) {
		idx = len(chords) - 1
	}
	return chords[idx]
}

// drumNoteForEvent maps a NoteEvent Pitch field on a drum role to a MIDI
// percussion key. Empty / "x" → the role's canonical hit. A bare integer
// → that MIDI note directly. Anything else falls back to the canonical hit.
func drumNoteForEvent(roleName, pitch string) int {
	pitch = strings.TrimSpace(pitch)
	if pitch == "" || pitch == "x" || pitch == "X" {
		return drumCanonicalNote(roleName)
	}
	if n, ok := parsePositiveInt(pitch); ok && n >= 0 && n < 128 {
		return n
	}
	return drumCanonicalNote(roleName)
}

func drumCanonicalNote(roleName string) int {
	switch strings.ToLower(strings.TrimSpace(roleName)) {
	case "kick":
		return 36
	case "snare":
		return 38
	case "clap":
		return 39
	case "rim":
		return 37
	case "hat", "hihat", "hat_closed":
		return 42
	case "openhat", "hat_open":
		return 46
	case "ride":
		return 51
	case "crash":
		return 49
	case "tom", "tom-low":
		return 45
	case "tom-mid":
		return 47
	case "tom-high":
		return 50
	case "cowbell":
		return 56
	case "shaker":
		return 70
	default:
		return 42
	}
}

func parsePositiveInt(s string) (int, bool) {
	n := 0
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
		if n > 1000 {
			return 0, false
		}
	}
	return n, true
}
