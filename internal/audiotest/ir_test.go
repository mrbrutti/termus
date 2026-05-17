// internal/audiotest/ir_test.go
package audiotest

import (
	"math"
	"testing"
)

func TestConvolveImpulseReturnsIR(t *testing.T) {
	ir := []float64{0.5, 0.4, 0.3, 0.2, 0.1}
	x := Click(0, 8, 1.0)
	y := Convolve(x, ir)
	for i, v := range ir {
		if math.Abs(y[i]-v) > 1e-9 {
			t.Fatalf("y[%d] = %g, want %g", i, y[i], v)
		}
	}
}

func TestConvolveLengthIsSum(t *testing.T) {
	x := make([]float64, 10)
	h := make([]float64, 5)
	y := Convolve(x, h)
	if len(y) != 14 {
		t.Fatalf("len(y) = %d, want 14", len(y))
	}
}

func TestAssertConvolutionResponseAcceptsMatchingOutput(t *testing.T) {
	ir := []float64{0.5, 0.4, 0.3}
	x := Click(0, 8, 1.0)
	y := Convolve(x, ir)
	AssertConvolutionResponse(t, x, ir, y, 1e-9)
}

func TestAssertConvolutionResponseRejectsMismatchedOutput(t *testing.T) {
	ir := []float64{0.5, 0.4, 0.3}
	x := Click(0, 8, 1.0)
	y := Convolve(x, ir)
	y[2] += 0.5 // corrupt
	stub := &testing.T{}
	AssertConvolutionResponse(stub, x, ir, y, 1e-9)
	if !stub.Failed() {
		t.Fatal("expected assertion to fail on corrupted output")
	}
}
