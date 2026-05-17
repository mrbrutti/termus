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

func TestConvolveEmptyInputReturnsNil(t *testing.T) {
	if y := Convolve(nil, []float64{1, 2, 3}); y != nil {
		t.Fatalf("Convolve(nil, h) = %v, want nil", y)
	}
	if y := Convolve([]float64{1, 2, 3}, nil); y != nil {
		t.Fatalf("Convolve(x, nil) = %v, want nil", y)
	}
	if y := Convolve(nil, nil); y != nil {
		t.Fatalf("Convolve(nil, nil) = %v, want nil", y)
	}
}

func TestAssertConvolutionResponseFlagsShortOutput(t *testing.T) {
	ir := []float64{0.5, 0.4, 0.3}
	x := Click(0, 8, 1.0)
	full := Convolve(x, ir)
	truncated := full[:len(full)-1] // missing one sample at the tail
	stub := &testing.T{}
	AssertConvolutionResponse(stub, x, ir, truncated, 1e-9)
	if !stub.Failed() {
		t.Fatal("expected assertion to fail when output is shorter than convolution")
	}
}
