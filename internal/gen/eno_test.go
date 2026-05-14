package gen

import (
	"math"
	"testing"
)

func TestEnoDeterministic(t *testing.T) {
	a := NewEno()
	b := NewEno()
	a.Seed(42)
	b.Seed(42)

	const n = 4096
	la := make([]float64, n)
	ra := make([]float64, n)
	lb := make([]float64, n)
	rb := make([]float64, n)
	a.Next(la, ra)
	b.Next(lb, rb)

	for i := 0; i < n; i++ {
		if la[i] != lb[i] || ra[i] != rb[i] {
			t.Fatalf("non-deterministic at i=%d: a=(%g,%g) b=(%g,%g)",
				i, la[i], ra[i], lb[i], rb[i])
		}
	}

	// Run two more blocks to catch state-corruption bugs that only show up
	// after the first Next() returns.
	for round := 1; round <= 2; round++ {
		a.Next(la, ra)
		b.Next(lb, rb)
		for i := 0; i < n; i++ {
			if la[i] != lb[i] || ra[i] != rb[i] {
				t.Fatalf("non-deterministic at round=%d i=%d: a=(%g,%g) b=(%g,%g)",
					round, i, la[i], ra[i], lb[i], rb[i])
			}
		}
	}
}

func TestEnoProducesAudio(t *testing.T) {
	a := NewEno()
	a.Seed(1)
	// One second.
	l := make([]float64, 44100)
	r := make([]float64, 44100)
	a.Next(l, r)
	var sum float64
	for i := range l {
		sum += l[i]*l[i] + r[i]*r[i]
	}
	rms := math.Sqrt(sum / float64(2*len(l)))
	if rms < 0.01 {
		t.Fatalf("eno RMS=%g, want >= 0.01 (was the generator silent?)", rms)
	}
	if rms > 0.5 {
		t.Fatalf("eno RMS=%g, expected < 0.5 (gain regression?)", rms)
	}
}

func TestEnoName(t *testing.T) {
	if NewEno().Name() != "eno-drift" {
		t.Fatalf("Name() = %q", NewEno().Name())
	}
}
