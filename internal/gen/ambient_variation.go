package gen

import "math/rand"

func variedContour(rng *rand.Rand, minLen, maxLen int) []int {
	if minLen < 1 {
		minLen = 1
	}
	if maxLen < minLen {
		maxLen = minLen
	}
	base := pickMelodicPhrase(rng)
	n := minLen
	if span := maxLen - minLen; span > 0 {
		n += rng.Intn(span + 1)
	}
	if n > len(base) {
		return trimOrRepeatPhrase(base, n, 0)
	}
	out := make([]int, n)
	copy(out, base[:n])
	return out
}

func variedRegisterShift(rng *rand.Rand) int {
	choices := []int{-12, 0, 12}
	return choices[rng.Intn(len(choices))]
}
