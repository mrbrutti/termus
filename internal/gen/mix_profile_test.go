package gen

import (
	"math"
	"testing"
)

type mixStubAlgo struct {
	name        string
	sectionGain float64
}

func (a mixStubAlgo) Name() string { return a.name }
func (a mixStubAlgo) Seed(int64)   {}
func (a mixStubAlgo) Next(left, right []float64) {
}
func (a mixStubAlgo) SectionGain() float64 { return a.sectionGain }

func TestEffectiveOutputGainUsesStaticTrimAndSectionGain(t *testing.T) {
	algo := mixStubAlgo{name: "glass-fm", sectionGain: 1.1}
	got := EffectiveOutputGain(algo)
	want := 0.55 * 1.1
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("EffectiveOutputGain = %v, want %v", got, want)
	}
}

func TestSectionMixProfileForCadenceAndOutro(t *testing.T) {
	cadence := SectionMixProfileFor(FormSection{Kind: FormCadence})
	outro := SectionMixProfileFor(FormSection{Kind: FormOutro})
	if cadence.Gain <= 1.0 {
		t.Fatalf("cadence gain = %v, want > 1", cadence.Gain)
	}
	if outro.Gain >= 1.0 {
		t.Fatalf("outro gain = %v, want < 1", outro.Gain)
	}
	if cadence.ExpressionDelta <= 0 || outro.ExpressionDelta >= 0 {
		t.Fatalf("unexpected expression deltas: cadence=%d outro=%d", cadence.ExpressionDelta, outro.ExpressionDelta)
	}
}
