// internal/audiotest/assert_test.go
package audiotest

import (
	"testing"
)

func TestAssertRMSDBPasses(t *testing.T) {
	// Unit-amplitude sine: RMS = 1/sqrt(2) ≈ -3.01 dB
	s := Sine(440, 1.0, 44100, 44100)
	AssertRMSDB(t, s, -3.01, 0.1)
}

func TestAssertRMSDBFails(t *testing.T) {
	s := Sine(440, 1.0, 44100, 44100)
	stub := &testing.T{}
	AssertRMSDB(stub, s, -12.0, 0.1)
	if !stub.Failed() {
		t.Fatal("expected AssertRMSDB to fail when actual RMS is far from target")
	}
}

func TestAssertPeakDBPasses(t *testing.T) {
	s := Sine(440, 0.5, 44100, 44100) // peak 0.5 → -6.02 dB
	AssertPeakDB(t, s, -6.02, 0.1)
}

func TestAssertPeakDBFails(t *testing.T) {
	s := Sine(440, 0.5, 44100, 44100)
	stub := &testing.T{}
	AssertPeakDB(stub, s, 0.0, 0.1)
	if !stub.Failed() {
		t.Fatal("expected AssertPeakDB to fail when peak is far from target")
	}
}
