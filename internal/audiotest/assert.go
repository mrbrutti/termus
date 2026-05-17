// internal/audiotest/assert.go
//
// Test assertions wrapping the DSP helpers in signal.go. All assertions
// call t.Helper() so test failures point at the caller line.
package audiotest

import (
	"math"
	"testing"
)

// AssertRMSDB fails the test if the RMS level of buf differs from wantDB by
// more than tolDB.
func AssertRMSDB(t testing.TB, buf []float64, wantDB, tolDB float64) {
	t.Helper()
	got := ToDB(RMS(buf))
	if math.Abs(got-wantDB) > tolDB {
		t.Errorf("RMS = %.2f dB, want %.2f ± %.2f dB", got, wantDB, tolDB)
	}
}

// AssertPeakDB fails the test if the peak amplitude of buf (in dBFS) differs
// from wantDB by more than tolDB.
func AssertPeakDB(t testing.TB, buf []float64, wantDB, tolDB float64) {
	t.Helper()
	got := ToDB(Peak(buf))
	if math.Abs(got-wantDB) > tolDB {
		t.Errorf("Peak = %.2f dBFS, want %.2f ± %.2f dB", got, wantDB, tolDB)
	}
}
