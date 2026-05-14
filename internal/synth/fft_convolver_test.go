package synth

import (
	"math"
	"math/rand"
	"testing"
)

// TestFFTConvolverMatchesDirect verifies that FFT-based convolution produces
// the same output as direct time-domain convolution (modulo the FFT
// convolver's blockSize-sample latency).
func TestFFTConvolverMatchesDirect(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	// IR with length not a multiple of blockSize, to exercise partial last
	// partition with zero-padding.
	ir := make([]float64, 1000)
	for i := range ir {
		// Exponentially decaying noise — looks IR-shaped.
		ir[i] = (2*rng.Float64() - 1) * math.Exp(-3.0*float64(i)/float64(len(ir)))
	}

	const blockSize = 256
	directConv := NewConvolver(ir)
	fftConv := NewFFTConvolver(ir, blockSize)
	if fftConv == nil {
		t.Fatal("NewFFTConvolver returned nil")
	}

	// Drive both with the same input, compare outputs accounting for the
	// FFT convolver's latency.
	const totalSamples = 3000
	input := make([]float64, totalSamples)
	for i := range input {
		input[i] = 2*rng.Float64() - 1
	}
	directOut := make([]float64, totalSamples)
	fftOut := make([]float64, totalSamples)
	for i, x := range input {
		directOut[i] = directConv.Tick(x)
		fftOut[i] = fftConv.Tick(x)
	}

	// FFT output is delayed by blockSize samples. Compare directOut[i] with
	// fftOut[i+blockSize] for i where both exist.
	const tol = 1e-6
	var maxErr float64
	for i := 0; i+blockSize < totalSamples; i++ {
		diff := math.Abs(directOut[i] - fftOut[i+blockSize])
		if diff > maxErr {
			maxErr = diff
		}
	}
	if maxErr > tol {
		t.Fatalf("FFT vs direct max error = %g (tolerance %g)", maxErr, tol)
	}
}

// TestFFTConvolverImpulsePreservesIR verifies that feeding a single impulse
// (1, 0, 0, ...) through the FFT convolver outputs the IR — delayed by the
// block-size latency.
func TestFFTConvolverImpulsePreservesIR(t *testing.T) {
	ir := []float64{0.5, 0.4, 0.3, 0.2, 0.1, 0, 0, 0, 0, 0, 0, 0}
	const blockSize = 8
	c := NewFFTConvolver(ir, blockSize)
	if c == nil {
		t.Fatal("NewFFTConvolver returned nil")
	}
	// Send impulse, then enough zeros to flush the IR + latency.
	got := make([]float64, blockSize+len(ir)+blockSize)
	got[0] = c.Tick(1.0)
	for i := 1; i < len(got); i++ {
		got[i] = c.Tick(0)
	}
	// First blockSize samples should be 0 (latency).
	for i := 0; i < blockSize; i++ {
		if math.Abs(got[i]) > 1e-9 {
			t.Fatalf("latency period: got[%d] = %g, want 0", i, got[i])
		}
	}
	// Next len(ir) samples should match IR.
	for i := 0; i < len(ir); i++ {
		want := ir[i]
		gotV := got[blockSize+i]
		if math.Abs(gotV-want) > 1e-9 {
			t.Fatalf("impulse response sample %d: got %g, want %g", i, gotV, want)
		}
	}
}
