package track

import (
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestAnalyzeBuildsReviewMetrics(t *testing.T) {
	const src = `
title: Reviewable
style: jazz
substyle: trio-after-hours
roles:
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 9 . 7 3"
  piano:
    family: acoustic_piano
    pattern: "x..x .x.."
  bass:
    family: bass
    pattern: "x... x..."
  ride:
    family: drums
    pattern: "x..x.x.. | x..x.xx."
sections:
  - id: head
    duration: 32s
    harmony: "Dm9 G13 | Cmaj9 A7"
    scene: "head clipped"
    variation: "statement"
  - id: outro
    duration: 16s
    harmony: "Dm9 G13 | Cmaj9 Cmaj9"
    scene: "outro cadence"
    variation: "cadence"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 7, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	report := Analyze(file, compiled)
	if got, want := len(report.Sections), 2; got != want {
		t.Fatalf("section count = %d, want %d", got, want)
	}
	if report.Substyle != "trio-after-hours" {
		t.Fatalf("substyle = %q", report.Substyle)
	}
	checkMetric := func(name string, v float64) {
		if v < 0 || v > 1 {
			t.Fatalf("%s metric out of range: %.3f", name, v)
		}
	}
	checkMetric("repetition", report.Metrics.RepetitionFatigue)
	checkMetric("contrast", report.Metrics.SectionContrast)
	checkMetric("cadence", report.Metrics.CadenceSpacing)
	checkMetric("lead occupancy", report.Metrics.LeadOccupancy)
	checkMetric("register spread", report.Metrics.RegisterSpread)
	checkMetric("harmonic color", report.Metrics.HarmonicColorRetention)
	checkMetric("ensemble diversity", report.Metrics.EnsembleDiversity)
	if report.Metrics.HarmonicColorRetention < 0.99 {
		t.Fatalf("expected harmonic color retention near 1, got %.3f", report.Metrics.HarmonicColorRetention)
	}
	if len(report.Sections[0].RoleNames) == 0 {
		t.Fatal("expected section role names")
	}
}
