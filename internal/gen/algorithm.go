// Package gen exposes the generative-music algorithm interface used by termus.
package gen

// Algorithm produces stereo PCM samples on demand. Implementations must be
// wait-free in Next (no locks, no allocations).
type Algorithm interface {
	// Name returns a short identifier, e.g. "eno-drift".
	Name() string
	// Seed (re)initializes the algorithm deterministically.
	Seed(s int64)
	// Next fills left and right with the next block of samples.
	// len(left) == len(right) is guaranteed by the caller.
	Next(left, right []float64)
}

// IterationApplier (SP19-B) is an optional capability that lets a section's
// algorithm receive an iteration counter when the host loops the section
// schedule. The host (TUI / playlist) calls ApplyIteration before the next
// section's Seed; algorithms that implement it can rewrite their plan to add
// drum-fill density, alternate voicings, activate extra roles, etc.
type IterationApplier interface {
	ApplyIteration(iter int)
}
