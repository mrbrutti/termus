// Package scope provides a lock-free single-writer single-reader ring buffer
// of mono samples for the audio visualizer.
package scope

import (
	"math"
	"sync/atomic"
)

// Ring stores the last `cap` mono samples. The writer (audio goroutine) calls
// Write; the reader (UI goroutine) calls Snapshot. Concurrent Write and
// Snapshot are safe; concurrent Write or concurrent Snapshot are not.
type Ring struct {
	buf  []atomic.Uint64 // float64 bit representation stored as uint64
	cap  int
	wpos atomic.Uint64 // monotonically increasing write count
}

func NewRing(cap int) *Ring {
	if cap < 2 {
		cap = 2
	}
	buf := make([]atomic.Uint64, cap)
	return &Ring{buf: buf, cap: cap}
}

// Write appends samples to the ring. Wraps when full.
func (r *Ring) Write(samples []float64) {
	w := r.wpos.Load()
	for _, s := range samples {
		bits := math.Float64bits(s)
		r.buf[int(w%uint64(r.cap))].Store(bits)
		w++
	}
	r.wpos.Store(w)
}

// Snapshot copies the most recent len(dst) samples into dst in chronological
// order (oldest first). If fewer samples have been written than len(dst), the
// missing prefix is left as zeroes.
func (r *Ring) Snapshot(dst []float64) {
	w := r.wpos.Load()
	n := uint64(len(dst))
	if n > uint64(r.cap) {
		n = uint64(r.cap)
	}
	start := uint64(0)
	if w > n {
		start = w - n
	}
	for i := uint64(0); i < n; i++ {
		bits := r.buf[int((start+i)%uint64(r.cap))].Load()
		dst[i] = math.Float64frombits(bits)
	}
	// If dst is longer than what we wrote into it (n < len(dst) only when
	// requested > cap), zero the tail.
	for i := int(n); i < len(dst); i++ {
		dst[i] = 0
	}
}
