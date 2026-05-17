package gen

import (
	"math"
	"sort"
	"strings"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

type AuthoredChordSpan struct {
	StartSlot int
	EndSlot   int
	Label     string
}

type AuthoredRenderTrack struct {
	Name            string
	Family          string
	Tone            []string
	Articulation    string
	Register        string
	Prominence      string
	Channel         int32
	Program         int32
	Velocity        int32
	Pan             int32
	Reverb          int32
	Chorus          int32
	Brightness      int32
	Notes           []int
	VelocityPattern []int32
	TimingOffsets   []float64
	Gate            float64
	SwingAmount     float64
	Legato          bool
	TieRepeats      bool
	OverlapSec      float64
	FireProbability float64
}

type AuthoredTrackPlan struct {
	Style       string
	Section     string
	Key         string
	Tempo       string
	BPM         float64
	DurationSec float64
	BarCount    int
	SlotCount   int
	ChordSpans  []AuthoredChordSpan
	Tracks      []AuthoredRenderTrack
}

type AuthoredTrack struct {
	spec           AlgoSpec
	sf             *meltysynth.SoundFont
	plan           AuthoredTrackPlan
	profile        ControlProfile
	core           *sf2Core
	samplesElapsed int64
	barSamples     int64
}

func NewAuthoredTrack(spec AlgoSpec, sf *meltysynth.SoundFont, plan AuthoredTrackPlan) Algorithm {
	return &AuthoredTrack{
		spec:    spec,
		sf:      sf,
		plan:    plan,
		profile: DefaultControlProfile(),
	}
}

func (a *AuthoredTrack) Name() string { return a.spec.Name }

func (a *AuthoredTrack) ApplyControlProfile(profile ControlProfile) {
	a.profile = profileOrDefault(profile)
}

func (a *AuthoredTrack) Seed(seed int64) {
	a.samplesElapsed = 0
	if a.sf == nil {
		a.core = nil
		return
	}
	core, err := newSF2Core(a.sf, authoredMasterGain(a.spec.Name), seed)
	if err != nil {
		a.core = nil
		return
	}
	applyMaxSF2Palette(core, a.spec.Name)
	a.configureMasterBus(core)
	for _, setup := range a.uniqueChannelSetups() {
		core.setProgram(setup.Channel, setup.Program)
		core.setPan(setup.Channel, setup.Pan)
		core.setReverbSend(setup.Channel, SectionCC(setup.Reverb, ReverbDelta(a.profile)))
		core.setChorusSend(setup.Channel, setup.Chorus)
		core.setChannelCutoff(setup.Channel, SectionCC(setup.Brightness, BrightnessDelta(a.profile)))
	}
	for _, track := range a.plan.Tracks {
		velocityPattern := append([]int32(nil), track.VelocityPattern...)
		timingOffsets := append([]float64(nil), track.TimingOffsets...)
		cfg := SF2Track{
			Channel:         track.Channel,
			Velocity:        authoredVelocity(track.Velocity, a.profile),
			Notes:           append([]int(nil), track.Notes...),
			PeriodSec:       a.plan.DurationSec,
			Gate:            authoredGate(track.Gate, a.profile),
			SwingAmount:     authoredSwing(track.SwingAmount, a.profile),
			Legato:          track.Legato,
			TieRepeats:      track.TieRepeats,
			OverlapSec:      track.OverlapSec,
			FireProbability: authoredFireProbability(track.FireProbability, a.profile),
		}
		if len(velocityPattern) > 0 {
			cfg.ResolveVelocity = func(slot int, key int, base int32) int32 {
				if len(velocityPattern) == 0 {
					return base
				}
				idx := slot % len(velocityPattern)
				v := base + velocityPattern[idx]
				if v < 18 {
					return 18
				}
				if v > 127 {
					return 127
				}
				return v
			}
		}
		if len(timingOffsets) > 0 {
			cfg.ResolveTimingOffsetSec = func(slot int) float64 {
				if len(timingOffsets) == 0 {
					return 0
				}
				return timingOffsets[slot%len(timingOffsets)]
			}
		}
		core.addTrack(cfg)
	}
	a.core = core
	if a.plan.BarCount > 0 && a.plan.DurationSec > 0 {
		a.barSamples = secondsToSamples(a.plan.DurationSec / float64(a.plan.BarCount))
	}
}

func (a *AuthoredTrack) Next(left, right []float64) {
	if a.core == nil {
		for i := range left {
			left[i], right[i] = 0, 0
		}
		return
	}
	a.core.renderInto(left, right)
	a.samplesElapsed += int64(len(left))
}

func (a *AuthoredTrack) DebugStatus() DebugStatus {
	status := DebugStatus{
		Section: a.plan.Section,
		Chord:   a.currentChordLabel(),
	}
	if a.barSamples > 0 {
		status.Bar = sampleBarIndex(a.samplesElapsed, a.barSamples) + 1
	}
	return status
}

func (a *AuthoredTrack) SetReverbIR(ir []float64, wet float64) {
	if a.core == nil {
		return
	}
	a.core.setConvolutionIR(ir, wet)
}

func (a *AuthoredTrack) SectionGain() float64 {
	if kind, ok := authoredSectionKind(a.plan.Section, a.plan.Style); ok {
		return clampGain(SectionMixProfileFor(FormSection{Kind: kind}).Gain)
	}
	return 1.0
}

func (a *AuthoredTrack) uniqueChannelSetups() []AuthoredRenderTrack {
	seen := map[int32]AuthoredRenderTrack{}
	order := make([]int32, 0)
	for _, track := range a.plan.Tracks {
		if existing, ok := seen[track.Channel]; ok {
			if existing.Program == 0 && track.Program != 0 {
				seen[track.Channel] = track
			}
			continue
		}
		seen[track.Channel] = track
		order = append(order, track.Channel)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	out := make([]AuthoredRenderTrack, 0, len(order))
	for _, channel := range order {
		out = append(out, seen[channel])
	}
	return out
}

func (a *AuthoredTrack) currentChordLabel() string {
	if len(a.plan.ChordSpans) == 0 || a.plan.SlotCount <= 0 || a.plan.DurationSec <= 0 {
		return ""
	}
	totalSamples := secondsToSamples(a.plan.DurationSec)
	if totalSamples <= 0 {
		return ""
	}
	slot := int((a.samplesElapsed % totalSamples) * int64(a.plan.SlotCount) / totalSamples)
	for _, span := range a.plan.ChordSpans {
		if slot >= span.StartSlot && slot < span.EndSlot {
			return span.Label
		}
	}
	return a.plan.ChordSpans[len(a.plan.ChordSpans)-1].Label
}

func (a *AuthoredTrack) configureMasterBus(core *sf2Core) {
	switch a.spec.Name {
	case "lofi":
		core.setMasterEQ(180, 1.5, 4200, -3.5)
		core.setMasterLowpass(5200, 0.707)
		core.setTapeHiss(0.0025)
		core.setTapeSaturation(0.30)
		core.setVinylCrackle(6, 0.02, 0.8)
	case "jazz":
		core.setMasterEQ(140, 1.5, 6200, 0.8)
	case "bells":
		core.setMasterEQ(180, 0.5, 5600, -1.2)
	case "ambient", "drone", "phase":
		core.setMasterEQ(160, 1.0, 5000, -0.8)
	case "classical", "lullaby":
		core.setMasterEQ(180, 0.8, 5800, -0.3)
	}
}

func authoredVelocity(base int32, profile ControlProfile) int32 {
	v := base + int32(ProfileCentered(profile.Density))*4
	if v < 24 {
		v = 24
	}
	if v > 120 {
		v = 120
	}
	return v
}

func authoredGate(base float64, profile ControlProfile) float64 {
	if base <= 0 {
		base = 0.95
	}
	base += 0.04 * float64(ProfileCentered(profile.Phrase))
	if base < 0.18 {
		return 0.18
	}
	if base > 1.6 {
		return 1.6
	}
	return base
}

func authoredSwing(base float64, profile ControlProfile) float64 {
	base += SwingOffsetSeconds(profile, 0.05)
	if base < 0 {
		return 0
	}
	if base > 0.28 {
		return 0.28
	}
	return base
}

func authoredFireProbability(base float64, profile ControlProfile) float64 {
	if base <= 0 || base >= 1 {
		return 1
	}
	if ProfileCentered(profile.Density) > 0 {
		base += 0.08 * float64(ProfileCentered(profile.Density))
	}
	if base < 0.15 {
		return 0.15
	}
	if base > 1 {
		return 1
	}
	return base
}

func authoredMasterGain(style string) float64 {
	switch style {
	case "jazz":
		return 2.1
	case "lofi":
		return 2.5
	case "bells":
		return 3.0
	case "ambient":
		return 2.8
	case "drone":
		return 3.0
	case "classical":
		return 2.4
	case "phase":
		return 2.6
	case "lullaby":
		return 2.8
	default:
		return 2.4
	}
}

func authoredSectionKind(section, style string) (FormSectionKind, bool) {
	label := strings.ToLower(strings.TrimSpace(section))
	switch {
	case strings.Contains(label, "intro"), strings.Contains(label, "count-in"), strings.Contains(label, "threshold"):
		return FormIntro, true
	case strings.Contains(label, "bridge"), strings.Contains(label, "middle"), strings.Contains(label, "chorus"):
		return FormB, true
	case strings.Contains(label, "breakdown"), strings.Contains(label, "shadow"), strings.Contains(label, "interior"):
		return FormBreakdown, true
	case strings.Contains(label, "outro"), strings.Contains(label, "last"), strings.Contains(label, "close"), strings.Contains(label, "return"):
		return FormOutro, true
	case strings.Contains(label, "cadence"):
		return FormCadence, true
	case strings.Contains(label, "prime"), strings.Contains(label, "lift"), strings.Contains(label, "answer"):
		return FormAprime, true
	default:
		if style == "ambient" || style == "drone" {
			return FormA, true
		}
		return FormA, true
	}
}

func authoredRound(v float64) int {
	return int(math.Round(v))
}
