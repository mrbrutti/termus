package gen

import "math/rand"

func invertNumericPhrase(src []int) []int {
	if len(src) == 0 {
		return nil
	}
	pivot := src[0]
	out := make([]int, len(src))
	for i, v := range src {
		out[i] = pivot - (v - pivot)
	}
	return out
}

func transposeNumericPhrase(src []int, delta int) []int {
	out := copyPhrase(src)
	for i := range out {
		out[i] += delta
	}
	return out
}

func stretchNumericPhrase(src []int, factor int) []int {
	if len(src) == 0 || factor <= 1 {
		return copyPhrase(src)
	}
	out := make([]int, 0, len(src)*factor)
	for _, v := range src {
		for i := 0; i < factor; i++ {
			out = append(out, v)
		}
	}
	return out
}

func transformNumericPhrase(rng *rand.Rand, src []int) []int {
	if len(src) == 0 {
		return nil
	}
	var out []int
	switch rng.Intn(5) {
	case 0:
		out = transposeNumericPhrase(src, []int{-2, -1, 1, 2}[rng.Intn(4)])
	case 1:
		out = invertNumericPhrase(src)
	case 2:
		out = rotatePhrase(src, 1+rng.Intn(maxInt(1, len(src)-1)))
	case 3:
		out = reversePhrase(src)
	default:
		out = stretchNumericPhrase(src, 2)
	}
	return trimOrRepeatPhrase(out, len(src), 0)
}

func transformNumericMotifMemory(rng *rand.Rand, base MotifMemory) MotifMemory {
	return MotifMemory{
		A:       transformNumericPhrase(rng, base.A),
		Aprime:  transformNumericPhrase(rng, firstNonEmpty(base.Aprime, base.A)),
		B:       transformNumericPhrase(rng, firstNonEmpty(base.B, base.A)),
		Cadence: transformNumericPhrase(rng, firstNonEmpty(base.Cadence, base.A)),
		Outro:   transformNumericPhrase(rng, firstNonEmpty(base.Outro, base.Cadence, base.A)),
	}
}
