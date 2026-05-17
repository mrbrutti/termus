// cmd/termus-listencheck/baseline_test.go
package main

import (
	"path/filepath"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
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

// TestCommittedBaselineMatchesCurrentRender renders the default non-SF2
// corpus end-to-end and asserts no drift against the committed baseline.
// This is the regression net that SP1+ sub-plans rely on.
func TestCommittedBaselineMatchesCurrentRender(t *testing.T) {
	base, err := readBaseline("../../testdata/listencheck/baseline.json")
	if err != nil {
		t.Skipf("no committed baseline: %v", err)
	}
	corpus := []corpusCase{
		{Name: "ambient-synth-42", Algo: "ambient-synth", Seed: 42, Seconds: 12},
		{Name: "classical-synth-99", Algo: "classical-synth", Seed: 99, Seconds: 14},
	}
	current := make([]baselineEntry, 0, len(corpus))
	for _, item := range corpus {
		if _, ok := gen.Resolve(item.Algo); !ok {
			t.Skipf("algorithm %q not registered", item.Algo)
		}
		m, err := measureCorpusItem(item)
		if err != nil {
			t.Fatalf("measure %s: %v", item.Name, err)
		}
		current = append(current, baselineEntry{
			Name:        item.Name,
			Algo:        item.Algo,
			Seed:        item.Seed,
			Seconds:     item.Seconds,
			Measurement: measurementFromAudiotest(m),
		})
	}
	drifts := compareBaselines(base, current)
	if len(drifts) > 0 {
		for _, d := range drifts {
			t.Errorf("drift: %s", d)
		}
	}
}
