// internal/audiotest/ir.go
//
// Naive time-domain convolution and an assertion that compares an actual
// processor output against the mathematical convolution of input × IR.
// Used to verify per-bus reverb wiring in SP2.
package audiotest

import (
	"math"
	"testing"
)

// Convolve returns y[n] = sum_k x[k]·h[n-k]. O(len(x)·len(h)); fine for
// short test IRs (≤ a few thousand samples). Returns nil if either input
// is empty.
func Convolve(x, h []float64) []float64 {
	if len(x) == 0 || len(h) == 0 {
		return nil
	}
	y := make([]float64, len(x)+len(h)-1)
	for i := range x {
		xi := x[i]
		if xi == 0 {
			continue
		}
		for j := range h {
			y[i+j] += xi * h[j]
		}
	}
	return y
}

// AssertConvolutionResponse fails the test if any sample of output diverges
// from the mathematical convolution of input × ir by more than tol. Also
// fails if output is shorter than the expected convolution (a short output
// usually means the processor truncated its tail).
func AssertConvolutionResponse(t testing.TB, input, ir, output []float64, tol float64) {
	t.Helper()
	want := Convolve(input, ir)
	if len(output) < len(want) {
		t.Errorf("output is %d samples, want at least %d (convolution length)",
			len(output), len(want))
		return
	}
	var maxErr float64
	maxIdx := -1
	for i := 0; i < len(want); i++ {
		d := math.Abs(want[i] - output[i])
		if d > maxErr {
			maxErr = d
			maxIdx = i
		}
	}
	if maxErr > tol {
		t.Errorf("convolution residual max = %.6g at sample %d (want ≤ %.6g)",
			maxErr, maxIdx, tol)
	}
}
