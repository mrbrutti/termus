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
// short test IRs (≤ a few thousand samples).
func Convolve(x, h []float64) []float64 {
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
// from the mathematical convolution of input × ir by more than tol.
func AssertConvolutionResponse(t testing.TB, input, ir, output []float64, tol float64) {
	t.Helper()
	want := Convolve(input, ir)
	n := len(want)
	if len(output) < n {
		n = len(output)
	}
	var maxErr float64
	maxIdx := -1
	for i := 0; i < n; i++ {
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
