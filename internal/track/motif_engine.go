package track

import (
	"fmt"
	"strconv"
	"strings"
)

// MotifPattern (SP18) represents a parsed motif as an ordered list of tokens.
// Each token is either a scale-degree slot (digit, optionally with octave
// adornments) or a rest "." or a bar separator "|". Empty patterns are valid.
type MotifPattern struct {
	Tokens []string
}

// ParseMotifPattern splits the pattern string into whitespace tokens. Empty
// strings collapse to an empty MotifPattern.
func ParseMotifPattern(pattern string) MotifPattern {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return MotifPattern{}
	}
	return MotifPattern{Tokens: strings.Fields(p)}
}

// String returns the motif as a single space-joined string.
func (m MotifPattern) String() string {
	return strings.Join(m.Tokens, " ")
}

// Notes returns just the non-rest, non-bar tokens.
func (m MotifPattern) Notes() []string {
	out := make([]string, 0, len(m.Tokens))
	for _, t := range m.Tokens {
		if t == "." || t == "|" || t == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// MotifSequence shifts every scale-degree token in the motif by `steps` scale
// degrees (i.e. diatonic transposition by step). Non-digit tokens are passed
// through. Octave adornments ("<", ">") are preserved.
func MotifSequence(m MotifPattern, steps int) MotifPattern {
	if steps == 0 || len(m.Tokens) == 0 {
		return m
	}
	out := make([]string, len(m.Tokens))
	for i, t := range m.Tokens {
		out[i] = shiftDegreeToken(t, steps)
	}
	return MotifPattern{Tokens: out}
}

// MotifAugment doubles the duration of each note token by inserting a rest
// "." after each note. Factor 2 = doubled; factor 3 = tripled. Anything else
// is treated as 2.
func MotifAugment(m MotifPattern, factor int) MotifPattern {
	if factor < 2 {
		factor = 2
	}
	if len(m.Tokens) == 0 {
		return m
	}
	out := make([]string, 0, len(m.Tokens)*factor)
	for _, t := range m.Tokens {
		out = append(out, t)
		if t == "|" {
			continue
		}
		for k := 1; k < factor; k++ {
			out = append(out, ".")
		}
	}
	return MotifPattern{Tokens: out}
}

// MotifDiminish halves the duration of each note by dropping every other
// rest token. factor 2 drops every other rest, factor 3 drops two out of
// every three. Bar markers are preserved.
func MotifDiminish(m MotifPattern, factor int) MotifPattern {
	if factor < 2 {
		factor = 2
	}
	if len(m.Tokens) == 0 {
		return m
	}
	out := make([]string, 0, len(m.Tokens))
	dropped := 0
	for _, t := range m.Tokens {
		if t == "." {
			if dropped < factor-1 {
				dropped++
				continue
			}
			dropped = 0
		}
		out = append(out, t)
	}
	return MotifPattern{Tokens: out}
}

// MotifRetrograde returns the motif with the token order reversed, but bar
// markers held in their original positions (so the bar grid stays sensible).
func MotifRetrograde(m MotifPattern) MotifPattern {
	if len(m.Tokens) == 0 {
		return m
	}
	// Separate non-bar tokens, reverse them, then re-thread bars at their
	// original indices.
	nonBars := make([]string, 0, len(m.Tokens))
	barAt := make([]bool, len(m.Tokens))
	for i, t := range m.Tokens {
		if t == "|" {
			barAt[i] = true
			continue
		}
		nonBars = append(nonBars, t)
	}
	for i, j := 0, len(nonBars)-1; i < j; i, j = i+1, j-1 {
		nonBars[i], nonBars[j] = nonBars[j], nonBars[i]
	}
	out := make([]string, len(m.Tokens))
	k := 0
	for i := range m.Tokens {
		if barAt[i] {
			out[i] = "|"
			continue
		}
		out[i] = nonBars[k]
		k++
	}
	return MotifPattern{Tokens: out}
}

// MotifInvert mirrors scale-degree tokens around a pivot. pivot is a scale
// degree (1..7). For each note token, the inverted degree = 2*pivot - degree.
// Octave adornments are preserved. Non-digit tokens are passed through.
func MotifInvert(m MotifPattern, pivot int) MotifPattern {
	if pivot < 1 {
		pivot = 5
	}
	out := make([]string, len(m.Tokens))
	for i, t := range m.Tokens {
		out[i] = invertDegreeToken(t, pivot)
	}
	return MotifPattern{Tokens: out}
}

// MotifFragment returns only the first `keep` note tokens of the motif (rest
// tokens and bar markers preserved between them). keep <= 0 returns empty.
// Useful for "fragment" treatment — a short echo of the motif.
func MotifFragment(m MotifPattern, keepNotes int) MotifPattern {
	if keepNotes <= 0 {
		return MotifPattern{}
	}
	out := make([]string, 0, len(m.Tokens))
	notesSeen := 0
	for _, t := range m.Tokens {
		if t == "." || t == "|" || t == "" {
			out = append(out, t)
			continue
		}
		if notesSeen >= keepNotes {
			out = append(out, ".")
			continue
		}
		out = append(out, t)
		notesSeen++
	}
	return MotifPattern{Tokens: out}
}

// MotifOrnament adds one diatonic neighbour note before each note token. The
// neighbour is +1 scale degree (upper neighbour). When density <= 1, only
// every Nth note gets an ornament. density 1 = every note, 2 = every other.
func MotifOrnament(m MotifPattern, density int) MotifPattern {
	if density < 1 {
		density = 1
	}
	out := make([]string, 0, len(m.Tokens)*2)
	noteIdx := 0
	for _, t := range m.Tokens {
		if t == "." || t == "|" || t == "" {
			out = append(out, t)
			continue
		}
		if noteIdx%density == 0 {
			out = append(out, shiftDegreeToken(t, 1))
		}
		out = append(out, t)
		noteIdx++
	}
	return MotifPattern{Tokens: out}
}

// ApplyMotifTreatment returns the motif transformed according to the
// treatment label. Unknown treatments return the motif unchanged.
//
// Treatments and their default transformations:
//
//	introduce: identity (the motif as-authored)
//	hint:      fragment to the first 2 notes
//	vary:      sequence ±1 scale degree
//	develop:   sequence +2 then ornament every other note
//	fragment:  fragment to the first 3 notes
//	return:    identity but with ornament density 2 (slight elaboration)
//	retrograde:  reverse the motif
//	invert:    invert around the 5th degree
//	augment:   double durations
//	diminish:  halve durations
func ApplyMotifTreatment(m MotifPattern, treatment string) MotifPattern {
	switch strings.ToLower(strings.TrimSpace(treatment)) {
	case "", "introduce", "statement", "head":
		return m
	case "hint":
		return MotifFragment(m, 2)
	case "vary":
		return MotifSequence(m, 1)
	case "develop":
		return MotifOrnament(MotifSequence(m, 2), 2)
	case "fragment":
		return MotifFragment(m, 3)
	case "return":
		return MotifOrnament(m, 2)
	case "retrograde":
		return MotifRetrograde(m)
	case "invert":
		return MotifInvert(m, 5)
	case "augment":
		return MotifAugment(m, 2)
	case "diminish":
		return MotifDiminish(m, 2)
	}
	return m
}

// MotifTreatmentNames returns the treatment labels recognised by
// ApplyMotifTreatment.
func MotifTreatmentNames() []string {
	return []string{
		"introduce", "hint", "vary", "develop", "fragment", "return",
		"retrograde", "invert", "augment", "diminish",
	}
}

// shiftDegreeToken returns the token's scale degree shifted by `steps`
// diatonic steps. Octave adornments (>, <) are preserved on the result.
// Accidentals (b, #) are stripped during the shift (we treat the digit as
// the scale degree); the caller is responsible for picking accidentals
// appropriate to the harmony — the engine does this when resolving to MIDI.
func shiftDegreeToken(t string, steps int) string {
	if t == "" || t == "." || t == "|" {
		return t
	}
	prefix, body, suffix := splitDegreeToken(t)
	deg, err := strconv.Atoi(body)
	if err != nil {
		return t
	}
	shifted := deg + steps
	// Wrap into 1..7 (with octave adjustment).
	octaveShift := 0
	for shifted > 7 {
		shifted -= 7
		octaveShift++
	}
	for shifted < 1 {
		shifted += 7
		octaveShift--
	}
	out := fmt.Sprintf("%s%d%s", prefix, shifted, suffix)
	// Octave shift: prepend ">" or "<" markers per step.
	for k := 0; k < octaveShift; k++ {
		out = ">" + out
	}
	for k := 0; k < -octaveShift; k++ {
		out = "<" + out
	}
	return out
}

func invertDegreeToken(t string, pivot int) string {
	if t == "" || t == "." || t == "|" {
		return t
	}
	prefix, body, suffix := splitDegreeToken(t)
	deg, err := strconv.Atoi(body)
	if err != nil {
		return t
	}
	mirrored := 2*pivot - deg
	// Wrap to 1..7.
	octaveShift := 0
	for mirrored > 7 {
		mirrored -= 7
		octaveShift--
	}
	for mirrored < 1 {
		mirrored += 7
		octaveShift++
	}
	out := fmt.Sprintf("%s%d%s", prefix, mirrored, suffix)
	for k := 0; k < octaveShift; k++ {
		out = ">" + out
	}
	for k := 0; k < -octaveShift; k++ {
		out = "<" + out
	}
	return out
}

// splitDegreeToken splits a token like ">b5" into prefix ">b", body "5",
// suffix "" — accidentals stay in the prefix, octave markers also in prefix
// (preserved by the upstream caller). When body parsing fails, body is
// returned as the entire token and prefix/suffix are empty.
func splitDegreeToken(t string) (string, string, string) {
	// Strip leading octave markers and accidentals.
	prefix := ""
	i := 0
	for i < len(t) {
		c := t[i]
		if c == '<' || c == '>' || c == 'b' || c == '#' {
			prefix += string(c)
			i++
			continue
		}
		break
	}
	body := ""
	for i < len(t) && t[i] >= '0' && t[i] <= '9' {
		body += string(t[i])
		i++
	}
	suffix := t[i:]
	if body == "" {
		return "", t, ""
	}
	return prefix, body, suffix
}

// motifToEventsForBeats expands a motif pattern into NoteEvents spaced across
// the given beat range. The motif's note tokens are distributed evenly: total
// span / number-of-positions, starting at startBeat. Rest tokens contribute
// silence (no event emitted). Each note is given duration matching the slot.
//
// pitchPrefix is currently unused but reserved for future per-section pitch
// resolution; today the events emit the raw degree tokens and the existing
// pitch resolver translates them at resolve time.
//
// vel is the base velocity for each event (typically 80-90).
func motifToEventsForBeats(m MotifPattern, startBeat, beats float64, vel int) []NoteEvent {
	positions := 0
	for _, t := range m.Tokens {
		if t == "|" {
			continue
		}
		positions++
	}
	if positions == 0 || beats <= 0 {
		return nil
	}
	step := beats / float64(positions)
	if step <= 0 {
		return nil
	}
	out := make([]NoteEvent, 0, positions)
	idx := 0
	for _, t := range m.Tokens {
		if t == "|" {
			continue
		}
		beat := startBeat + step*float64(idx)
		idx++
		if t == "." || t == "" {
			continue
		}
		dur := step * 0.85
		if dur > 2.0 {
			dur = 2.0
		}
		out = append(out, NoteEvent{
			Beat:  beat,
			Pitch: t,
			Dur:   dur,
			Vel:   vel,
		})
	}
	return out
}
