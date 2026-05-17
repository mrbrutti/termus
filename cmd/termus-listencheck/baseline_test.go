// cmd/termus-listencheck/baseline_test.go
package main

import (
	"path/filepath"
	"testing"
)

func TestBaselineRoundTrip(t *testing.T) {
	dir := t.TempDir()
	entries := []baselineEntry{
		{Name: "x", Algo: "ambient-synth", Seed: 1, Seconds: 0.5,
			Measurement: measurement{RMSDb: -12, PeakDb: -3, CentroidHz: 1200, Frames: 22050, SampleRate: 44100}},
	}
	path := filepath.Join(dir, "baseline.json")
	if err := writeBaseline(path, entries); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := readBaseline(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 1 || got[0].Name != "x" || got[0].Measurement.RMSDb != -12 {
		t.Fatalf("round-trip mismatch; got %+v", got)
	}
}

func TestCompareBaselineDetectsRMSDrift(t *testing.T) {
	base := []baselineEntry{
		{Name: "x", Measurement: measurement{RMSDb: -12, PeakDb: -3, CentroidHz: 1000}},
	}
	current := []baselineEntry{
		{Name: "x", Measurement: measurement{RMSDb: -10, PeakDb: -3, CentroidHz: 1000}},
	}
	drifts := compareBaselines(base, current)
	if len(drifts) != 1 {
		t.Fatalf("expected 1 drift; got %d (%v)", len(drifts), drifts)
	}
}

func TestCompareBaselineAcceptsCentroidWithin10Percent(t *testing.T) {
	base := []baselineEntry{
		{Name: "x", Measurement: measurement{RMSDb: -12, PeakDb: -3, CentroidHz: 1000}},
	}
	current := []baselineEntry{
		{Name: "x", Measurement: measurement{RMSDb: -12, PeakDb: -3, CentroidHz: 1080}},
	}
	drifts := compareBaselines(base, current)
	if len(drifts) != 0 {
		t.Fatalf("expected no drift; got %v", drifts)
	}
}
