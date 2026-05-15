package gen

// MotifMemory stores a small phrase family that higher-level form sections can
// recall and vary. The slices usually contain plan codes rather than concrete
// MIDI notes.
type MotifMemory struct {
	A       []int
	Aprime  []int
	B       []int
	Cadence []int
	Outro   []int
}

func (m MotifMemory) PhraseFor(kind FormSectionKind) []int {
	switch kind {
	case FormAprime:
		return firstNonEmpty(m.Aprime, m.A)
	case FormB:
		return firstNonEmpty(m.B, m.A)
	case FormCadence:
		return firstNonEmpty(m.Cadence, m.A)
	case FormOutro:
		return firstNonEmpty(m.Outro, m.Cadence, m.A)
	default:
		return m.A
	}
}

func firstNonEmpty(seq ...[]int) []int {
	for _, part := range seq {
		if len(part) > 0 {
			return part
		}
	}
	return nil
}

func copyPhrase(in []int) []int {
	if len(in) == 0 {
		return nil
	}
	out := make([]int, len(in))
	copy(out, in)
	return out
}

func sequencePhrase(in []int, substitutions map[int]int) []int {
	out := copyPhrase(in)
	for i, v := range out {
		if next, ok := substitutions[v]; ok {
			out[i] = next
		}
	}
	return out
}

func stitchPhrase(parts ...[]int) []int {
	total := 0
	for _, part := range parts {
		total += len(part)
	}
	out := make([]int, 0, total)
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}

func trimOrRepeatPhrase(src []int, n int, fill int) []int {
	if n <= 0 {
		return nil
	}
	if len(src) == 0 {
		out := make([]int, n)
		for i := range out {
			out[i] = fill
		}
		return out
	}
	out := make([]int, n)
	for i := range out {
		out[i] = src[i%len(src)]
	}
	return out
}

func reversePhrase(src []int) []int {
	if len(src) == 0 {
		return nil
	}
	out := make([]int, len(src))
	for i := range src {
		out[len(src)-1-i] = src[i]
	}
	return out
}

func rotatePhrase(src []int, shift int) []int {
	if len(src) == 0 {
		return nil
	}
	n := len(src)
	shift = ((shift % n) + n) % n
	out := make([]int, n)
	for i := range src {
		out[i] = src[(i+shift)%n]
	}
	return out
}
