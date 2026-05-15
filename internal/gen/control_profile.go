package gen

import (
	"math"

	"github.com/mrbrutti/termus/internal/synth"
)

// ControlProfile stores the shared user-facing music macros. Values are 0..4
// with 2 as the neutral midpoint.
type ControlProfile struct {
	Density    int `json:"density"`
	Brightness int `json:"brightness"`
	Motion     int `json:"motion"`
	Reverb     int `json:"reverb"`
	Swing      int `json:"swing"`
	DroneDepth int `json:"drone_depth"`
	Tempo      int `json:"tempo"`
	Phrase     int `json:"phrase"`
}

type ControlProfileApplier interface {
	ApplyControlProfile(ControlProfile)
}

func DefaultControlProfile() ControlProfile {
	return ControlProfile{
		Density:    2,
		Brightness: 2,
		Motion:     2,
		Reverb:     2,
		Swing:      2,
		DroneDepth: 2,
		Tempo:      2,
		Phrase:     2,
	}
}

func clampProfileValue(v int) int {
	if v < 0 {
		return 0
	}
	if v > 4 {
		return 4
	}
	return v
}

func normalizeProfile(profile ControlProfile) ControlProfile {
	profile.Density = clampProfileValue(profile.Density)
	profile.Brightness = clampProfileValue(profile.Brightness)
	profile.Motion = clampProfileValue(profile.Motion)
	profile.Reverb = clampProfileValue(profile.Reverb)
	profile.Swing = clampProfileValue(profile.Swing)
	profile.DroneDepth = clampProfileValue(profile.DroneDepth)
	profile.Tempo = clampProfileValue(profile.Tempo)
	profile.Phrase = clampProfileValue(profile.Phrase)
	return profile
}

func profileOrDefault(profile ControlProfile) ControlProfile {
	if profile == (ControlProfile{}) {
		return DefaultControlProfile()
	}
	return normalizeProfile(profile)
}

func ProfileCentered(v int) int {
	return clampProfileValue(v) - 2
}

func DensityGain(profile ControlProfile) float64 {
	return 1.0 + 0.08*float64(ProfileCentered(profile.Density))
}

func BrightnessDelta(profile ControlProfile) int32 {
	return int32(ProfileCentered(profile.Brightness) * 8)
}

func ReverbDelta(profile ControlProfile) int32 {
	return int32(ProfileCentered(profile.Reverb) * 10)
}

func DroneDepthDelta(profile ControlProfile) int32 {
	return int32(ProfileCentered(profile.DroneDepth) * 10)
}

func SwingOffsetSeconds(profile ControlProfile, scale float64) float64 {
	return float64(ProfileCentered(profile.Swing)) * scale
}

func TempoScale(profile ControlProfile) float64 {
	return math.Pow(1.06, float64(ProfileCentered(profile.Tempo)))
}

func PhraseScale(profile ControlProfile) float64 {
	return math.Pow(1.12, float64(ProfileCentered(profile.Phrase)))
}

// ApplyControlProfile applies a profile to a fresh algorithm and returns an
// optionally wrapped version with generic post-effects.
func ApplyControlProfile(algo Algorithm, profile ControlProfile) Algorithm {
	if algo == nil {
		return nil
	}
	algo = ConfigureControlProfile(algo, profile)
	profile = normalizeProfile(profile)
	if profile.Brightness == 2 && profile.Motion == 2 && profile.Reverb == 2 {
		return algo
	}
	return newControlFXAlgorithm(algo, profile)
}

// ConfigureControlProfile applies algorithm-specific control hooks without
// adding any generic post-FX wrappers. Export paths use this so type
// assertions like TuningExporter still work on the concrete algorithm.
func ConfigureControlProfile(algo Algorithm, profile ControlProfile) Algorithm {
	if algo == nil {
		return nil
	}
	profile = normalizeProfile(profile)
	if applier, ok := algo.(ControlProfileApplier); ok {
		applier.ApplyControlProfile(profile)
	}
	return algo
}

type controlFXAlgorithm struct {
	Algorithm
	profile ControlProfile
	highL   *synth.HighShelf
	highR   *synth.HighShelf
	revL    *synth.Reverb
	revR    *synth.Reverb
	left    []float64
	right   []float64
	phase   float64
}

func newControlFXAlgorithm(algo Algorithm, profile ControlProfile) Algorithm {
	profile = normalizeProfile(profile)
	brightnessDB := float64(ProfileCentered(profile.Brightness)) * 2.2
	reverbWet := clampUnit(0.10 + 0.07*float64(profile.Reverb))
	return &controlFXAlgorithm{
		Algorithm: algo,
		profile:   profile,
		highL:     synth.NewHighShelf(3800, brightnessDB, 0.707),
		highR:     synth.NewHighShelf(3800, brightnessDB, 0.707),
		revL:      synth.NewReverb(reverbWet * 0.18),
		revR:      synth.NewReverbRight(reverbWet * 0.18),
	}
}

func (a *controlFXAlgorithm) Next(left, right []float64) {
	if cap(a.left) < len(left) {
		a.left = make([]float64, len(left))
		a.right = make([]float64, len(right))
	}
	a.left = a.left[:len(left)]
	a.right = a.right[:len(right)]
	a.Algorithm.Next(a.left, a.right)
	motionDepth := 0.04 + 0.03*float64(a.profile.Motion)
	motionRate := 0.00018 + 0.00005*float64(a.profile.Motion)
	for i := range left {
		lv := a.highL.Tick(a.left[i])
		rv := a.highR.Tick(a.right[i])
		lv = a.revL.Tick(lv)
		rv = a.revR.Tick(rv)
		pan := math.Sin(a.phase) * motionDepth
		widthL := clampUnit(0.5 - pan)
		widthR := clampUnit(0.5 + pan)
		mid := (lv + rv) * 0.5
		side := (lv - rv) * 0.5
		left[i] = mid + side*widthL*2.0
		right[i] = mid - side*widthR*2.0
		a.phase += motionRate
	}
}

func (a *controlFXAlgorithm) DebugStatus() DebugStatus {
	status := SnapshotDebugStatus(a.Algorithm)
	return status
}

func (a *controlFXAlgorithm) SectionGain() float64 {
	if provider, ok := a.Algorithm.(SectionGainProvider); ok {
		return provider.SectionGain()
	}
	return 1.0
}

func (a *controlFXAlgorithm) SetReverbIR(ir []float64, wet float64) {
	if rev, ok := a.Algorithm.(SF2Reverberator); ok {
		rev.SetReverbIR(ir, wet)
	}
}

func clampUnit(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}
