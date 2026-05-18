package track

import (
	"math"
	"strings"
)

// VoiceContext provides the harmonic + temporal context the voicing engine
// needs to expand one chord region into a sequence of NoteEvents.
//
// Beats are 1-indexed (matching NoteEvent.Beat). DurationBeats is the
// total length of the chord region in beats; Tempo is bpm. Register is one
// of "low", "mid", "high" — the engine picks an octave for chord tones using
// this hint. Key is the section key string (e.g. "Dmin", "C") and is used to
// help the engine decide on chromatic approach notes for walking bass.
//
// BassPresent, when true, tells the voicing engine that some other role
// (typically a dedicated bass) is already covering the chord root, so the
// generated voicing should omit the root to keep the harmony uncluttered.
type VoiceContext struct {
	Chord         string
	NextChord     string
	StartBeat     float64
	DurationBeats float64
	Tempo         float64
	Register      string
	Key           string
	BassPresent   bool
}

// GenerateVoicing returns the NoteEvents for one chord region using the
// given voicing style. Recognised style names:
//
//	walking_bass, walking_with_anticipation, pedal_root,
//	rhodes_comp, jazz_rootless_a, jazz_rootless_b,
//	drop2, drop3, shell_voicing,
//	pad_sustain, pad_crossfade, bell_arpeggio
//
// An empty / unknown style returns nil. The caller can decide whether that
// is a soft (warn) or hard (error) failure. Pitches are emitted as absolute
// MIDI note names (e.g. "F#4") so the downstream resolver can pass them
// through unchanged.
func GenerateVoicing(style string, ctx VoiceContext) []NoteEvent {
	style = strings.ToLower(strings.TrimSpace(style))
	if style == "" {
		return nil
	}
	if ctx.DurationBeats <= 0 {
		return nil
	}
	tones, ok := chordToneSemis(ctx.Chord)
	if !ok {
		return nil
	}
	rootPC := chordRootPC(ctx.Chord)
	if rootPC < 0 {
		return nil
	}
	nextRootPC := -1
	if ctx.NextChord != "" {
		nextRootPC = chordRootPC(ctx.NextChord)
	}
	switch style {
	case "walking_bass":
		return walkingBass(rootPC, tones, nextRootPC, ctx, false)
	case "walking_with_anticipation":
		return walkingBass(rootPC, tones, nextRootPC, ctx, true)
	case "pedal_root":
		return pedalRoot(rootPC, ctx)
	case "rhodes_comp":
		return rhodesComp(rootPC, tones, ctx)
	case "jazz_rootless_a":
		return rootlessVoicing(rootPC, tones, ctx, "a")
	case "jazz_rootless_b":
		return rootlessVoicing(rootPC, tones, ctx, "b")
	case "drop2":
		return drop2Voicing(rootPC, tones, ctx)
	case "drop3":
		return drop3Voicing(rootPC, tones, ctx)
	case "shell_voicing", "shell":
		return shellVoicing(rootPC, tones, ctx)
	case "pad_sustain":
		return padSustain(rootPC, tones, ctx, false)
	case "pad_crossfade":
		return padSustain(rootPC, tones, ctx, true)
	case "bell_arpeggio":
		return bellArpeggio(rootPC, tones, ctx)
	}
	return nil
}

// chordRootPC returns the pitch-class (0..11, C=0) of the chord's root, or
// -1 on parse failure.
func chordRootPC(chord string) int {
	chord = strings.TrimSpace(chord)
	if chord == "" {
		return -1
	}
	pc, _, ok := parseRootToken(chord)
	if !ok {
		return -1
	}
	return pc
}

// chordSuffix returns the chord-symbol suffix following the root (e.g.
// "Dm7b5" → "m7b5"). Empty string on parse failure.
func chordSuffix(chord string) string {
	chord = strings.TrimSpace(chord)
	if chord == "" {
		return ""
	}
	_, rest, ok := parseRootToken(chord)
	if !ok {
		return ""
	}
	return rest
}

// chordToneSemis returns the chord's tones as semitone offsets above the
// root (e.g. major-seventh chord → [0, 4, 7, 11]). Handles the common cases
// described in the SP16 brief: maj7, m7, 7, m7b5, dim7, sus4, 6, 9, 11, 13,
// b9, #9, b13, b5, #5.
func chordToneSemis(chord string) ([]int, bool) {
	chord = strings.TrimSpace(chord)
	if chord == "" {
		return nil, false
	}
	suffix := strings.ToLower(chordSuffix(chord))

	root, third, fifth, seventh := 0, 4, 7, 10
	hasSeventh := false

	switch {
	case strings.Contains(suffix, "m7b5") || strings.Contains(suffix, "ø"):
		third = 3
		fifth = 6
		seventh = 10
		hasSeventh = true
	case strings.Contains(suffix, "dim7"):
		third = 3
		fifth = 6
		seventh = 9
		hasSeventh = true
	case strings.Contains(suffix, "dim"):
		third = 3
		fifth = 6
		hasSeventh = false
	case strings.Contains(suffix, "sus2"):
		third = 2
		hasSeventh = strings.Contains(suffix, "7")
	case strings.Contains(suffix, "sus"):
		third = 5 // sus4
		hasSeventh = strings.Contains(suffix, "7")
	case strings.Contains(suffix, "maj"):
		third = 4
		seventh = 11
		hasSeventh = strings.Contains(suffix, "7") || strings.Contains(suffix, "9") || strings.Contains(suffix, "11") || strings.Contains(suffix, "13")
	case strings.HasPrefix(suffix, "m"), strings.HasPrefix(suffix, "min"):
		third = 3
		seventh = 10
		hasSeventh = strings.Contains(suffix, "7") || strings.Contains(suffix, "9") || strings.Contains(suffix, "11") || strings.Contains(suffix, "13")
	default:
		// Dominant family — implicit 7 when the symbol carries any extension.
		third = 4
		seventh = 10
		hasSeventh = strings.Contains(suffix, "7") || strings.Contains(suffix, "9") || strings.Contains(suffix, "11") || strings.Contains(suffix, "13")
	}

	if strings.Contains(suffix, "#5") || strings.Contains(suffix, "aug") {
		fifth = 8
	}
	if strings.Contains(suffix, "b5") && !strings.Contains(suffix, "m7b5") {
		fifth = 6
	}

	out := []int{root, third, fifth}
	if hasSeventh {
		out = append(out, seventh)
	}

	// 6th (not part of the 7/9/11/13 stack): "C6" → +9 semis. The "6" must
	// be a standalone digit, not part of "16", "13" or "60".
	if hasStandaloneDigit(suffix, "6") {
		out = append(out, 9)
	}

	if hasStandaloneDigit(suffix, "9") || strings.Contains(suffix, "11") || strings.Contains(suffix, "13") || strings.Contains(suffix, "add9") {
		// 9th = 14, but reflect b9/#9.
		switch {
		case strings.Contains(suffix, "b9"):
			out = append(out, 13)
		case strings.Contains(suffix, "#9"):
			out = append(out, 15)
		default:
			out = append(out, 14)
		}
	}

	if strings.Contains(suffix, "11") {
		// 11th = 17, #11 = 18, b11 = 16 (rare).
		switch {
		case strings.Contains(suffix, "#11"):
			out = append(out, 18)
		case strings.Contains(suffix, "b11"):
			out = append(out, 16)
		default:
			out = append(out, 17)
		}
	}

	if strings.Contains(suffix, "13") {
		switch {
		case strings.Contains(suffix, "b13"):
			out = append(out, 20)
		default:
			out = append(out, 21)
		}
	}

	return out, true
}

// hasStandaloneDigit reports whether suffix contains the digit `d` as an
// unaccompanied numeric token (not part of "11", "13", etc.). Caller passes
// lowercased input.
func hasStandaloneDigit(suffix, d string) bool {
	if !strings.Contains(suffix, d) {
		return false
	}
	// Walk every occurrence of the digit and check both neighbours.
	for i := 0; i+len(d) <= len(suffix); i++ {
		if suffix[i:i+len(d)] != d {
			continue
		}
		prevDigit := i > 0 && suffix[i-1] >= '0' && suffix[i-1] <= '9'
		nextDigit := i+len(d) < len(suffix) && suffix[i+len(d)] >= '0' && suffix[i+len(d)] <= '9'
		if !prevDigit && !nextDigit {
			return true
		}
	}
	return false
}

// pcToMidi finds the nearest MIDI key in [low,high] with the given pitch
// class, closest to anchorMidi. Used to place chord tones in a register.
func pcToMidi(pc int, anchorMidi, low, high int) int {
	pc = ((pc % 12) + 12) % 12
	// Search octave by octave around the anchor.
	best := -1
	bestDelta := math.MaxInt32
	for octave := 0; octave <= 10; octave++ {
		m := pc + octave*12
		if m < low || m > high {
			continue
		}
		d := abs(m - anchorMidi)
		if d < bestDelta {
			best = m
			bestDelta = d
		}
	}
	if best < 0 {
		// Fallback: nearest MIDI key in the broader range.
		for octave := 0; octave <= 10; octave++ {
			m := pc + octave*12
			if m >= 0 && m <= 127 {
				return m
			}
		}
		return 60
	}
	return best
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// registerAnchor returns a typical anchor MIDI key for the given Register hint.
func registerAnchor(register string) (anchor, low, high int) {
	switch strings.ToLower(strings.TrimSpace(register)) {
	case "low", "bass":
		return 40, 24, 55
	case "high", "lead":
		return 76, 60, 96
	default: // "mid", ""
		return 60, 48, 84
	}
}

// midiToName converts a MIDI key to a note name like "F#4". Sharps preferred.
var sharpNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func midiToName(midi int) string {
	if midi < 0 {
		midi = 0
	}
	if midi > 127 {
		midi = 127
	}
	pc := ((midi % 12) + 12) % 12
	octave := midi/12 - 1
	return sharpNames[pc] + itoa(octave)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	s := string(digits)
	if negative {
		s = "-" + s
	}
	return s
}

// ============================================================
// Voicing implementations
// ============================================================

func walkingBass(rootPC int, tones []int, nextRootPC int, ctx VoiceContext, anticipate bool) []NoteEvent {
	// One quarter note per beat. Beat 1: root. Beat 2: third (or fifth).
	// Beat 3: fifth (or third). Beat 4: chromatic approach to next root.
	anchor, low, high := registerAnchor("low")
	if ctx.Register == "" {
		anchor, low, high = 40, 28, 55
	}
	rootMidi := pcToMidi(rootPC, anchor, low, high)
	thirdSemi, fifthSemi := 4, 7
	for _, t := range tones {
		switch t {
		case 3:
			thirdSemi = 3
		case 4:
			thirdSemi = 4
		case 6:
			fifthSemi = 6
		case 7:
			fifthSemi = 7
		case 8:
			fifthSemi = 8
		}
	}
	thirdMidi := pcToMidi((rootPC+thirdSemi)%12, rootMidi+7, low, high)
	fifthMidi := pcToMidi((rootPC+fifthSemi)%12, rootMidi+9, low, high)

	// Chromatic approach to next root.
	approachTarget := rootMidi
	if nextRootPC >= 0 {
		approachTarget = pcToMidi(nextRootPC, rootMidi, low, high)
	}
	approachMidi := approachTarget - 1
	if approachMidi < low {
		approachMidi = approachTarget + 1
	}

	// Total beats: at minimum 4 (a bar). If the chord spans fewer beats, we
	// halve the figure (use just root + fifth). If it spans more, repeat.
	beats := int(math.Round(ctx.DurationBeats))
	if beats < 1 {
		beats = 1
	}
	events := make([]NoteEvent, 0, beats)
	startBeat := ctx.StartBeat
	if startBeat <= 0 {
		startBeat = 1.0
	}

	makePitches := func() []int {
		switch beats {
		case 1:
			return []int{rootMidi}
		case 2:
			return []int{rootMidi, fifthMidi}
		case 3:
			return []int{rootMidi, thirdMidi, fifthMidi}
		case 4:
			return []int{rootMidi, thirdMidi, fifthMidi, approachMidi}
		default:
			// Loop a 4-beat pattern for chord regions > 4 beats.
			out := make([]int, beats)
			pat := []int{rootMidi, thirdMidi, fifthMidi, approachMidi}
			for i := 0; i < beats; i++ {
				out[i] = pat[i%4]
			}
			return out
		}
	}
	pitches := makePitches()
	for i, m := range pitches {
		beat := startBeat + float64(i)
		dur := 1.0
		vel := 88
		if i == 0 {
			vel = 95
		}
		if anticipate && i == beats-1 && nextRootPC >= 0 {
			// Beat 4 of the bar anticipates the next root by half a beat:
			// move start half a beat earlier and play the next-root pitch.
			beat -= 0.5
			m = pcToMidi(nextRootPC, rootMidi, low, high)
			dur = 0.5
		}
		events = append(events, NoteEvent{
			Beat:  beat,
			Pitch: midiToName(m),
			Dur:   dur,
			Vel:   vel,
		})
	}
	return events
}

func pedalRoot(rootPC int, ctx VoiceContext) []NoteEvent {
	anchor, low, high := 40, 28, 55
	if r := strings.ToLower(strings.TrimSpace(ctx.Register)); r != "" {
		anchor, low, high = registerAnchor(r)
	}
	rootMidi := pcToMidi(rootPC, anchor, low, high)
	start := ctx.StartBeat
	if start <= 0 {
		start = 1
	}
	return []NoteEvent{{
		Beat:  start,
		Pitch: midiToName(rootMidi),
		Dur:   ctx.DurationBeats,
		Vel:   85,
	}}
}

// rhodesComp emits 4–6 chord stabs per 4 beats of chord duration. Each stab
// places 3–4 simultaneous chord tones (rendered as multiple NoteEvents at
// the same beat). When BassPresent, the root is omitted.
func rhodesComp(rootPC int, tones []int, ctx VoiceContext) []NoteEvent {
	anchor, low, high := registerAnchor(strFallback(ctx.Register, "mid"))
	if ctx.Register == "" {
		anchor, low, high = 60, 48, 78
	}

	// Pick the chord-tone palette: skip the root if a bass is present.
	semis := append([]int(nil), tones...)
	if ctx.BassPresent {
		filtered := semis[:0]
		for _, s := range semis {
			if s != 0 {
				filtered = append(filtered, s)
			}
		}
		semis = filtered
	}
	if len(semis) < 3 {
		// Need at least 3 chord tones; ensure the third and fifth at minimum.
		// Recover from filtering by re-adding the third + fifth from `tones`.
		want := map[int]struct{}{}
		for _, s := range semis {
			want[s] = struct{}{}
		}
		for _, s := range tones {
			if s == 0 && ctx.BassPresent {
				continue
			}
			if _, ok := want[s]; !ok {
				semis = append(semis, s)
				want[s] = struct{}{}
			}
			if len(semis) >= 4 {
				break
			}
		}
	}

	// Build the voiced chord (3–4 stacked tones).
	chord := []int{}
	for _, s := range semis {
		m := pcToMidi((rootPC+s)%12, anchor, low, high)
		chord = append(chord, m)
		if len(chord) >= 4 {
			break
		}
	}
	if len(chord) < 3 {
		return nil
	}

	// 4–6 stabs across DurationBeats. Use a syncopated pattern: hit on
	// beats 1, &2, 3, &4 (4 stabs over a 4-beat region) — or scale for the
	// region's actual length.
	beatsTotal := ctx.DurationBeats
	if beatsTotal <= 0 {
		return nil
	}
	// Stab positions for a 4-beat bar: 0, 1.5, 2, 3.5 — gives 4 stabs.
	// Use 0, 0.75, 1.5, 2, 3.5 for a denser 5-stab figure. We pick 4-stab.
	pattern := []float64{0, 1.5, 2, 3.5}
	startBeat := ctx.StartBeat
	if startBeat <= 0 {
		startBeat = 1
	}

	events := []NoteEvent{}
	cycles := int(math.Ceil(beatsTotal / 4.0))
	if cycles < 1 {
		cycles = 1
	}
	for c := 0; c < cycles; c++ {
		offset := float64(c) * 4
		for stab, p := range pattern {
			beat := startBeat + offset + p
			if beat-startBeat >= beatsTotal-0.05 {
				break
			}
			// Slight velocity variation per stab (beat 1 strongest).
			vel := 78
			switch stab {
			case 0:
				vel = 86
			case 1:
				vel = 72
			case 2:
				vel = 80
			case 3:
				vel = 70
			}
			// Vary inversion across bars: rotate the chord by `c`.
			rot := c % len(chord)
			rotated := append([]int{}, chord[rot:]...)
			rotated = append(rotated, chord[:rot]...)
			for _, m := range rotated {
				events = append(events, NoteEvent{
					Beat:  beat,
					Pitch: midiToName(m),
					Dur:   0.4,
					Vel:   vel,
					Art:   "",
				})
			}
		}
	}
	return events
}

// rootlessVoicing emits 4-note rootless chord voicings: 3-5-7-9 for type A,
// 7-9-3-5 for type B. Comping pattern: chord on beat 1 (held 2 beats), ghost
// stab on the "and" of beat 3.
func rootlessVoicing(rootPC int, tones []int, ctx VoiceContext, variant string) []NoteEvent {
	anchor, low, high := registerAnchor(strFallback(ctx.Register, "mid"))
	if ctx.Register == "" {
		anchor, low, high = 64, 52, 80
	}

	thirdSemi, fifthSemi, seventhSemi, ninthSemi := 4, 7, 10, 14
	for _, s := range tones {
		switch {
		case s == 3 || s == 4:
			thirdSemi = s
		case s == 6 || s == 7 || s == 8:
			fifthSemi = s
		case s == 9: // 6 (the "6" extension), used in major-6 voicings
			if seventhSemi == 10 && !chordHas(tones, 10) && !chordHas(tones, 11) {
				seventhSemi = 9
			}
		case s == 10 || s == 11:
			seventhSemi = s
		case s == 13 || s == 14 || s == 15:
			ninthSemi = s
		}
	}

	var stack []int
	switch variant {
	case "b":
		stack = []int{seventhSemi, ninthSemi, thirdSemi + 12, fifthSemi + 12}
	default: // "a"
		stack = []int{thirdSemi, fifthSemi, seventhSemi, ninthSemi}
	}

	voiced := make([]int, 0, 4)
	prev := anchor
	for _, s := range stack {
		m := pcToMidi((rootPC+s)%12, prev, low, high)
		// Ensure ascending order.
		for m < prev {
			m += 12
		}
		if m > high {
			m -= 12
		}
		voiced = append(voiced, m)
		prev = m
	}

	startBeat := ctx.StartBeat
	if startBeat <= 0 {
		startBeat = 1
	}
	events := []NoteEvent{}
	// Sustained chord on beat 1, dur = 2.5 beats.
	for _, m := range voiced {
		events = append(events, NoteEvent{
			Beat:  startBeat,
			Pitch: midiToName(m),
			Dur:   math.Min(2.5, ctx.DurationBeats),
			Vel:   78,
		})
	}
	// Ghost stab on the "and" of beat 3 (beat 3.5).
	if ctx.DurationBeats >= 4 {
		for _, m := range voiced {
			events = append(events, NoteEvent{
				Beat:  startBeat + 2.5,
				Pitch: midiToName(m),
				Dur:   0.4,
				Vel:   58,
				Art:   "ghost",
			})
		}
	}
	return events
}

func drop2Voicing(rootPC int, tones []int, ctx VoiceContext) []NoteEvent {
	return dropVoicingN(rootPC, tones, ctx, 2)
}

func drop3Voicing(rootPC int, tones []int, ctx VoiceContext) []NoteEvent {
	return dropVoicingN(rootPC, tones, ctx, 3)
}

func dropVoicingN(rootPC int, tones []int, ctx VoiceContext, drop int) []NoteEvent {
	anchor, low, high := registerAnchor(strFallback(ctx.Register, "mid"))
	if ctx.Register == "" {
		anchor, low, high = 64, 52, 84
	}
	// Pick top 4 chord tones: root, 3, 5, 7 (or a richer subset).
	picks := []int{0}
	for _, s := range tones {
		if s == 0 {
			continue
		}
		picks = append(picks, s)
		if len(picks) >= 4 {
			break
		}
	}
	if len(picks) < 4 {
		return nil
	}
	// Close position: rising stack from anchor.
	close := []int{}
	prev := anchor - 12
	for _, s := range picks {
		m := pcToMidi((rootPC+s)%12, prev, low, high)
		for m < prev {
			m += 12
		}
		if m > high {
			m -= 12
		}
		close = append(close, m)
		prev = m
	}
	if len(close) < 4 {
		return nil
	}
	// Drop the Nth from the top (1=top, 2=second-from-top, etc.) down an octave.
	idx := len(close) - drop
	if idx < 0 || idx >= len(close) {
		return nil
	}
	close[idx] -= 12

	startBeat := ctx.StartBeat
	if startBeat <= 0 {
		startBeat = 1
	}
	events := make([]NoteEvent, 0, len(close))
	for _, m := range close {
		events = append(events, NoteEvent{
			Beat:  startBeat,
			Pitch: midiToName(m),
			Dur:   ctx.DurationBeats,
			Vel:   80,
		})
	}
	return events
}

func shellVoicing(rootPC int, tones []int, ctx VoiceContext) []NoteEvent {
	// Root + 3 + 7.
	anchor, low, high := registerAnchor(strFallback(ctx.Register, "mid"))
	if ctx.Register == "" {
		anchor, low, high = 56, 44, 74
	}
	thirdSemi := 4
	seventhSemi := 10
	for _, s := range tones {
		switch s {
		case 3:
			thirdSemi = 3
		case 4:
			thirdSemi = 4
		case 10:
			seventhSemi = 10
		case 11:
			seventhSemi = 11
		}
	}
	rootMidi := pcToMidi(rootPC, anchor, low, high)
	thirdMidi := pcToMidi((rootPC+thirdSemi)%12, rootMidi+thirdSemi, low, high)
	seventhMidi := pcToMidi((rootPC+seventhSemi)%12, thirdMidi+5, low, high)

	startBeat := ctx.StartBeat
	if startBeat <= 0 {
		startBeat = 1
	}
	return []NoteEvent{
		{Beat: startBeat, Pitch: midiToName(rootMidi), Dur: ctx.DurationBeats, Vel: 80},
		{Beat: startBeat, Pitch: midiToName(thirdMidi), Dur: ctx.DurationBeats, Vel: 76},
		{Beat: startBeat, Pitch: midiToName(seventhMidi), Dur: ctx.DurationBeats, Vel: 78},
	}
}

func padSustain(rootPC int, tones []int, ctx VoiceContext, crossfade bool) []NoteEvent {
	// 3–4 chord tones, omit root. Multiple simultaneous pitches.
	anchor, low, high := registerAnchor(strFallback(ctx.Register, "mid"))
	if ctx.Register == "" {
		anchor, low, high = 60, 50, 78
	}
	picks := []int{}
	for _, s := range tones {
		if s == 0 {
			continue
		}
		picks = append(picks, s)
		if len(picks) >= 4 {
			break
		}
	}
	if len(picks) < 2 {
		// Fallback: include root.
		picks = append(picks, 0)
	}
	startBeat := ctx.StartBeat
	if startBeat <= 0 {
		startBeat = 1
	}
	dur := ctx.DurationBeats
	if crossfade {
		// Hold 0.5 beats past the chord boundary.
		dur += 0.5
	}
	out := []NoteEvent{}
	prev := anchor - 8
	for _, s := range picks {
		m := pcToMidi((rootPC+s)%12, prev, low, high)
		for m < prev {
			m += 12
		}
		if m > high {
			m -= 12
		}
		out = append(out, NoteEvent{
			Beat:  startBeat,
			Pitch: midiToName(m),
			Dur:   dur,
			Vel:   72,
		})
		prev = m
	}
	return out
}

// bellArpeggio arpeggiates chord tones across the chord duration, multi-octave.
func bellArpeggio(rootPC int, tones []int, ctx VoiceContext) []NoteEvent {
	anchor, low, high := registerAnchor(strFallback(ctx.Register, "high"))
	if ctx.Register == "" {
		anchor, low, high = 76, 60, 96
	}
	if len(tones) == 0 {
		return nil
	}
	dur := ctx.DurationBeats
	if dur <= 0 {
		return nil
	}
	spacing := dur / float64(len(tones)*2) // 2 octaves
	if spacing < 0.125 {
		spacing = 0.125
	}
	startBeat := ctx.StartBeat
	if startBeat <= 0 {
		startBeat = 1
	}
	events := []NoteEvent{}
	// Two octave arpeggio.
	prev := anchor
	idx := 0
	for t := 0.0; t+spacing <= dur+0.001; t += spacing {
		semis := tones[idx%len(tones)]
		octaveBump := 0
		if idx >= len(tones) {
			octaveBump = 12
		}
		m := pcToMidi((rootPC+semis)%12, prev+octaveBump, low, high)
		events = append(events, NoteEvent{
			Beat:  startBeat + t,
			Pitch: midiToName(m),
			Dur:   spacing * 0.9,
			Vel:   70,
		})
		prev = m
		idx++
	}
	return events
}

// ============================================================
// helpers
// ============================================================

func chordHas(tones []int, want int) bool {
	for _, t := range tones {
		if t == want {
			return true
		}
	}
	return false
}

func strFallback(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
