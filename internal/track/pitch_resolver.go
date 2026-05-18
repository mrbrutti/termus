package track

import (
	"regexp"
	"strconv"
	"strings"
)

// ResolvePitch converts a Pitch string from a NoteEvent into a MIDI note
// number, given the section's key, the current chord at the requested beat,
// and the role's default register. Returns -1 on parse failure so the
// caller can fall back (e.g. by emitting a warning and skipping the note).
//
// Accepted forms:
//
//	"C4", "F#3", "Bb5"   — absolute MIDI note name
//	"1", "b3", "#5", "7" — scale degree relative to the section's key
//	"1>", "5<<"          — scale degree with octave shift(s)
//	"R", "3", "5", "7",  — chord-relative degree (R = root, 3 = third, ...)
//	  "9", "11", "13"
//
// The chord-vs-key distinction is determined by inspecting the first
// character: an uppercase letter A–G that does not match "R" indicates a
// note name, "R" indicates a chord root, a digit indicates a degree.
// Degrees default to the chord interpretation when chord is well-defined
// and the form is one of R/3/5/7/9/11/13; otherwise they resolve against
// the key.
//
// When chord is the zero value or its root/scale are missing, chord-relative
// requests fall back to the key.
func ResolvePitch(pitch, key string, chord authoredChord, register string) int {
	pitch = strings.TrimSpace(pitch)
	if pitch == "" {
		return -1
	}
	// Note name?
	if midi, ok := parsePitchNoteName(pitch); ok {
		return midi
	}
	// Chord-relative degree letter "R" or numeric chord degree.
	if strings.HasPrefix(strings.ToUpper(pitch), "R") {
		return resolveChordDegree("R", pitch, chord, register, key)
	}
	if num, _, _, ok := parseDegreeForm(pitch); ok {
		// Treat 9/11/13 as chord-tones if the chord is set; otherwise key.
		if chord.RootPC >= 0 && (num == 9 || num == 11 || num == 13) {
			return resolveChordDegree(strconv.Itoa(num), pitch, chord, register, key)
		}
		// Scale degree against the key.
		return resolveScaleDegree(pitch, key, register)
	}
	return -1
}

// parsePitchNoteName parses absolute note names like "C4", "F#3", "Bb5".
// Returns (midiNote, true) on success.
var noteNameRE = regexp.MustCompile(`^([A-Ga-g])([#b]?)(\d)$`)

func parsePitchNoteName(s string) (int, bool) {
	m := noteNameRE.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, false
	}
	letter := strings.ToUpper(m[1])
	accidental := m[2]
	octave, err := strconv.Atoi(m[3])
	if err != nil {
		return 0, false
	}
	pc := map[string]int{"C": 0, "D": 2, "E": 4, "F": 5, "G": 7, "A": 9, "B": 11}[letter]
	switch accidental {
	case "#":
		pc++
	case "b":
		pc--
	}
	// MIDI: C-1 = 0, C0 = 12, C4 = 60.
	midi := pc + (octave+1)*12
	if midi < 0 || midi > 127 {
		return 0, false
	}
	return midi, true
}

// parseDegreeForm parses things like "1", "b3", "#5", "7", "1>", "5<<".
// Returns (degreeNumber, accidentalSemitones, octaveShift, ok).
var degreeRE = regexp.MustCompile(`^([b#])?(\d{1,2})([<>]*)$`)

func parseDegreeForm(s string) (int, int, int, bool) {
	m := degreeRE.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, 0, 0, false
	}
	acc := 0
	switch m[1] {
	case "#":
		acc = 1
	case "b":
		acc = -1
	}
	deg, err := strconv.Atoi(m[2])
	if err != nil || deg < 1 || deg > 13 {
		return 0, 0, 0, false
	}
	octShift := 0
	for _, r := range m[3] {
		switch r {
		case '>':
			octShift++
		case '<':
			octShift--
		}
	}
	return deg, acc, octShift, true
}

// resolveScaleDegree resolves a degree like "1", "b3", "#5", "7" against
// the section's key (e.g. "Cmin"). The natural scale degree pitches are
// taken from the key's mode (major / natural minor / dorian / mixolydian
// etc.). The returned MIDI note is placed near the role's default register.
func resolveScaleDegree(s, key, register string) int {
	deg, acc, octShift, ok := parseDegreeForm(s)
	if !ok {
		return -1
	}
	rootPC, mode := parseKeyForDegree(key)
	scale := scaleForMode(mode)
	// Reduce degrees > 7 into octave-aware indices.
	octBump := 0
	for deg > 7 {
		deg -= 7
		octBump++
	}
	if deg < 1 {
		deg = 1
	}
	if deg > 7 {
		deg = 7
	}
	pcOffset := scale[deg-1] + acc
	baseOctave := registerCenterOctave(register)
	midi := (baseOctave+1)*12 + rootPC + pcOffset + 12*(octShift+octBump)
	if midi < 0 || midi > 127 {
		return clampMIDI(midi)
	}
	return midi
}

// resolveChordDegree resolves "R", "3", "5", "7", "9", "11", "13" against
// the current chord. When the chord is unset / unparseable, falls back to
// the key.
func resolveChordDegree(canonical, raw string, chord authoredChord, register, key string) int {
	octShift := 0
	for _, r := range raw {
		switch r {
		case '>':
			octShift++
		case '<':
			octShift--
		}
	}
	if chord.RootPC < 0 || len(chord.Scale) == 0 {
		return resolveScaleDegree(raw, key, register)
	}
	var pcOffset int
	switch strings.ToUpper(canonical) {
	case "R":
		pcOffset = 0
	case "3":
		pcOffset = chordToneInterval(chord, 3)
	case "5":
		pcOffset = chordToneInterval(chord, 5)
	case "7":
		pcOffset = chordToneInterval(chord, 7)
	case "9":
		pcOffset = chordToneInterval(chord, 9)
	case "11":
		pcOffset = chordToneInterval(chord, 11)
	case "13":
		pcOffset = chordToneInterval(chord, 13)
	default:
		return resolveScaleDegree(raw, key, register)
	}
	baseOctave := registerCenterOctave(register)
	midi := (baseOctave+1)*12 + chord.RootPC + pcOffset + 12*octShift
	return clampMIDI(midi)
}

// chordToneInterval returns the semitone interval above the chord root for
// scale-degree N (3, 5, 7, 9, 11, 13). It first checks chord.Degrees and
// chord.Interval (already-resolved chord-voicing data), then falls back to
// defaults appropriate for the chord Kind.
func chordToneInterval(chord authoredChord, deg int) int {
	if chord.Degrees != nil {
		if v, ok := chord.Degrees[deg]; ok {
			return v
		}
	}
	// Common defaults by kind. These are intentionally simple — the chord
	// parser produces richer voicings, but unset cases get a musical
	// fallback so events don't disappear.
	maj := chord.Kind == "" || chord.Kind == "maj" || chord.Kind == "maj7" || chord.Kind == "Maj" || strings.HasPrefix(chord.Kind, "maj")
	min := chord.Kind == "min" || chord.Kind == "m" || chord.Kind == "min7" || strings.HasPrefix(chord.Kind, "m")
	dom := chord.Kind == "7" || chord.Kind == "9" || chord.Kind == "13"
	switch deg {
	case 3:
		if min {
			return 3
		}
		return 4
	case 5:
		return 7
	case 7:
		if maj {
			return 11
		}
		if dom {
			return 10
		}
		// minor or unspecified non-major → minor 7
		return 10
	case 9:
		return 14
	case 11:
		return 17
	case 13:
		return 21
	}
	return 0
}

// parseKeyForDegree splits a key string like "Cmin", "Bbmaj", "G dorian"
// into (rootPC, mode). Mode defaults to "maj".
func parseKeyForDegree(key string) (int, string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, "maj"
	}
	// Strip leading root.
	root, rest, ok := parseRootToken(key)
	if !ok {
		return 0, "maj"
	}
	mode := strings.ToLower(strings.TrimSpace(rest))
	mode = strings.TrimPrefix(mode, " ")
	switch {
	case mode == "" || strings.HasPrefix(mode, "maj"):
		return root, "maj"
	case strings.HasPrefix(mode, "min") || mode == "m":
		return root, "min"
	case strings.HasPrefix(mode, "dor"):
		return root, "dorian"
	case strings.HasPrefix(mode, "phr"):
		return root, "phrygian"
	case strings.HasPrefix(mode, "lyd"):
		return root, "lydian"
	case strings.HasPrefix(mode, "mix"):
		return root, "mixolydian"
	case strings.HasPrefix(mode, "loc"):
		return root, "locrian"
	}
	return root, "maj"
}

// scaleForMode returns the semitone offsets from the tonic for one octave
// of the given mode, 7 entries.
func scaleForMode(mode string) [7]int {
	switch mode {
	case "min":
		return [7]int{0, 2, 3, 5, 7, 8, 10} // natural minor
	case "dorian":
		return [7]int{0, 2, 3, 5, 7, 9, 10}
	case "phrygian":
		return [7]int{0, 1, 3, 5, 7, 8, 10}
	case "lydian":
		return [7]int{0, 2, 4, 6, 7, 9, 11}
	case "mixolydian":
		return [7]int{0, 2, 4, 5, 7, 9, 10}
	case "locrian":
		return [7]int{0, 1, 3, 5, 6, 8, 10}
	default:
		return [7]int{0, 2, 4, 5, 7, 9, 11} // major / ionian
	}
}

// registerCenterOctave converts a register hint ("low" / "mid" / "high" /
// "sub") into a default MIDI octave number for placing scale-degree notes.
//
// MIDI octave convention here matches the standard one where C4 = 60, so
// octave 4 places C at MIDI 60.
func registerCenterOctave(register string) int {
	switch strings.ToLower(strings.TrimSpace(register)) {
	case "sub", "low-sub":
		return 1
	case "low", "lo":
		return 2
	case "low-mid":
		return 3
	case "mid", "":
		return 4
	case "high-mid":
		return 4
	case "high", "hi":
		return 5
	case "very-high", "top":
		return 6
	}
	return 4
}

func clampMIDI(midi int) int {
	for midi < 12 {
		midi += 12
	}
	for midi > 108 {
		midi -= 12
	}
	return midi
}
