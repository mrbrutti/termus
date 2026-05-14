package scope

import (
	"sync"
	"testing"
)

func TestRingSnapshotReturnsLatest(t *testing.T) {
	r := NewRing(8)
	for i := 0; i < 12; i++ {
		r.Write([]float64{float64(i)})
	}
	out := make([]float64, 4)
	r.Snapshot(out)
	// Expect the last 4 values written: 8, 9, 10, 11.
	want := []float64{8, 9, 10, 11}
	for i, v := range want {
		if out[i] != v {
			t.Fatalf("Snapshot[%d] = %v, want %v (full=%v)", i, out[i], v, out)
		}
	}
}

func TestRingConcurrent(t *testing.T) {
	r := NewRing(1024)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100_000; i++ {
			r.Write([]float64{float64(i)})
		}
	}()
	// Reader does many snapshots; the race detector will catch unsynchronized access.
	out := make([]float64, 64)
	for i := 0; i < 1000; i++ {
		r.Snapshot(out)
	}
	wg.Wait()
}
