// internal/audiotest/transient_test.go
package audiotest

import (
	"math"
	"math/rand"
	"testing"
)

func TestFindTransientsDetectsBurstAfterSilence(t *testing.T) {
	const n = 8000
	buf := make([]float64, n)
	rng := rand.New(rand.NewSource(2))
	// First half silent. Second half: loud noise.
	for i := n / 2; i < n; i++ {
		buf[i] = 0.5 * (2*rng.Float64() - 1)
	}
	trans := FindTransients(buf, 256, 12.0)
	if len(trans) == 0 {
		t.Fatal("expected at least one transient at the silence→noise boundary")
	}
	near := false
	for _, s := range trans {
		if math.Abs(float64(s-n/2)) < 512 {
			near = true
			break
		}
	}
	if !near {
		t.Fatalf("expected transient near sample %d; got %v", n/2, trans)
	}
}

func TestFindTransientsReturnsEmptyForSteadySignal(t *testing.T) {
	buf := Sine(440, 0.5, 44100, 8000)
	trans := FindTransients(buf, 256, 12.0)
	if len(trans) != 0 {
		t.Fatalf("expected no transients in steady sine; got %v", trans)
	}
}

func TestAssertHasTransientAtPasses(t *testing.T) {
	const n = 8000
	buf := make([]float64, n)
	rng := rand.New(rand.NewSource(3))
	for i := 4000; i < n; i++ {
		buf[i] = 0.5 * (2*rng.Float64() - 1)
	}
	AssertHasTransientAt(t, buf, 4000, 600)
}
