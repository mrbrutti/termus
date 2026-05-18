package track

import "fmt"

// EuclideanRhythm generates a Euclidean (Bjorklund) rhythm pattern with k
// onsets distributed as evenly as possible across n steps, optionally rotated
// by r positions. Returns a boolean slice of length n where true = onset.
// Returns an error if k > n or either value is <= 0.
func EuclideanRhythm(k, n, rotate int) ([]bool, error) {
	if n <= 0 {
		return nil, fmt.Errorf("euclidean: n must be > 0, got %d", n)
	}
	if k < 0 {
		return nil, fmt.Errorf("euclidean: k must be >= 0, got %d", k)
	}
	if k > n {
		return nil, fmt.Errorf("euclidean: k (%d) must not exceed n (%d)", k, n)
	}

	// Bjorklund algorithm via Euclidean recursion.
	pattern := bjorklund(k, n)

	// Apply rotation (positive = shift right / delay by r steps).
	if rotate != 0 {
		r := ((rotate % n) + n) % n
		rotated := make([]bool, n)
		for i, v := range pattern {
			rotated[(i+r)%n] = v
		}
		pattern = rotated
	}
	return pattern, nil
}

// bjorklund implements the Bjorklund / Euclidean rhythm algorithm.
// It distributes k pulses into n steps as evenly as possible.
func bjorklund(k, n int) []bool {
	if k == 0 {
		return make([]bool, n)
	}
	if k == n {
		out := make([]bool, n)
		for i := range out {
			out[i] = true
		}
		return out
	}

	// Each "group" is a slice of bool. We start with k groups of [true]
	// and (n-k) groups of [false], then iteratively distribute remainders.
	ones := k
	zeros := n - k

	// groups[i] holds the pattern for one sub-group.
	groups := make([][]bool, ones+zeros)
	for i := 0; i < ones; i++ {
		groups[i] = []bool{true}
	}
	for i := ones; i < ones+zeros; i++ {
		groups[i] = []bool{false}
	}

	// Distribute the smaller set into the larger set until done.
	for {
		smaller := zeros
		larger := ones
		if smaller > larger {
			smaller, larger = larger, smaller
		}
		if smaller <= 1 {
			break
		}
		// Pair each group from the smaller set with a group from the larger set.
		newGroups := make([][]bool, 0, larger)
		for i := 0; i < smaller; i++ {
			merged := append(append([]bool(nil), groups[i]...), groups[larger+i]...)
			newGroups = append(newGroups, merged)
		}
		// Remainder groups (from the larger set that weren't paired).
		for i := smaller; i < larger; i++ {
			newGroups = append(newGroups, groups[i])
		}
		groups = newGroups
		ones = smaller
		zeros = larger - smaller
	}

	// Flatten all groups into a single pattern.
	out := make([]bool, 0, n)
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// euclideanPatternString converts a Euclidean rhythm to an x/. string.
func euclideanPatternString(k, n, rotate int) (string, error) {
	pattern, err := EuclideanRhythm(k, n, rotate)
	if err != nil {
		return "", err
	}
	buf := make([]byte, n)
	for i, hit := range pattern {
		if hit {
			buf[i] = 'x'
		} else {
			buf[i] = '.'
		}
	}
	return string(buf), nil
}
