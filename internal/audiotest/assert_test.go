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
	func() {
		defer func() { recover() }()
		AssertRMSDB(stub, s, -12.0, 0.1)
	}()
	if !stub.Failed() {
		t.Fatal("expected AssertRMSDB to fail when actual RMS is far from target")
	}
}
