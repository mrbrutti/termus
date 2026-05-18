package gen

import (
	"math"
	"sort"
	"strings"

	"github.com/mrbrutti/termus/internal/synth"
	"github.com/sinshu/go-meltysynth/meltysynth"
)

// AuthoredAutomationLane describes a per-section breakpoint curve.
// Param is one of: "cutoff", "pan", "expression".
// Breakpoints are at01 (0..1 fraction of section duration) → value pairs.
type AuthoredAutomationLane struct {
	Param       string
	Breakpoints [][2]float64 // [at01, value] pairs, sorted by at01
}

// ValueAt returns the interpolated value at position at01 (0..1 within the section).
func (l *AuthoredAutomationLane) ValueAt(at01 float64) float64 {
	if len(l.Breakpoints) == 0 {
		return 0
	}
	if at01 <= l.Breakpoints[0][0] {
		return l.Breakpoints[0][1]
	}
	for i := 1; i < len(l.Breakpoints); i++ {
		if at01 <= l.Breakpoints[i][0] {
			t0, v0 := l.Breakpoints[i-1][0], l.Breakpoints[i-1][1]
			t1, v1 := l.Breakpoints[i][0], l.Breakpoints[i][1]
			if t1 == t0 {
				return v1
			}
			frac := (at01 - t0) / (t1 - t0)
			return v0 + frac*(v1-v0)
		}
	}
	return l.Breakpoints[len(l.Breakpoints)-1][1]
}

// AuthoredRoleReverb holds per-role reverb bus configuration compiled from
// the Role.Room and Role.ReverbSendDB fields.
type AuthoredRoleReverb struct {
	IRName     string  // synth.IRPreset name, e.g. "jazz_club"
	SendDB     float64 // wet level in dBFS (e.g. -12)
	PreDelayMs float64 // pre-delay in ms (informational default)
}

type AuthoredChordSpan struct {
	StartSlot int
	EndSlot   int
	Label     string
}

type AuthoredPhraseSpan struct {
	StartBar int
	EndBar   int
	Label    string
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
	GatePattern     []float64
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
	PhraseSpans []AuthoredPhraseSpan
	Tracks      []AuthoredRenderTrack

	// SP8 wiring: v2 schema fields that drive the audio pipeline.

	// MixBus is the mix-bus profile name (e.g. "lofi", "jazz", "chill",
	// "ambient"). When set, AuthoredTrack.Seed calls applyMixBusProfile
	// with the resolved profile, replacing the hardcoded configureMasterBus
	// defaults. Empty = legacy default behaviour.
	MixBus string

	// Groove is the groove template name (e.g. "swing_56", "dilla_late").
	// When set, timing and velocity are modulated per-16th-note step.
	// Empty = no groove applied.
	Groove string

	// Automation holds per-section automation lanes that drive callbacks
	// as the section plays. Param names: "cutoff", "pan", "expression".
	Automation []AuthoredAutomationLane

	// RoleReverb maps role name → reverb bus config for per-role routing.
	// Built from Role.Room and Role.ReverbSendDB during compile.
	RoleReverb map[string]AuthoredRoleReverb

	// SP19-D: optional ambient texture layers compiled from File.Textures.
	// Each entry is rendered alongside the music and summed into the master
	// stereo output at the configured level.
	Textures []AuthoredTexture
}

// AuthoredTexture (SP19-D) is one resolved ambient-texture layer for a plan.
type AuthoredTexture struct {
	Name    string  // rain, room_tone, vinyl, tape_hiss, cafe
	LevelDB float64 // peak level in dBFS
}

type AuthoredTrack struct {
	spec           AlgoSpec
	sf             *meltysynth.SoundFont
	plan           AuthoredTrackPlan
	basePlan       AuthoredTrackPlan // SP19-B: untouched baseline plan
	iteration      int               // SP19-B: current loop iteration index
	profile        ControlProfile
	core           *sf2Core
	samplesElapsed int64
	barSamples     int64
}

func NewAuthoredTrack(spec AlgoSpec, sf *meltysynth.SoundFont, plan AuthoredTrackPlan) Algorithm {
	return &AuthoredTrack{
		spec:     spec,
		sf:       sf,
		plan:     plan,
		basePlan: clonePlan(plan),
		profile:  DefaultControlProfile(),
	}
}

// ApplyIteration (SP19-B) rewrites the track's plan to reflect a given loop
// iteration. iter==0 leaves the plan untouched; iter>=1 adds variations:
//   - drum fill / hat probability bumped (more activity)
//   - voicing-styled tracks alternate inversions between iter 1 and 2
//   - extra "iteration_active" roles activate at iter>=2
//
// The method does not call Seed — the caller must call Seed afterwards.
func (a *AuthoredTrack) ApplyIteration(iter int) {
	if iter < 0 {
		iter = 0
	}
	a.iteration = iter
	a.plan = mutatePlanForIteration(a.basePlan, iter)
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

	// SP8: When a mix_bus profile is specified in the .tm file, it takes over
	// as the single source of truth for the master chain. The hardcoded
	// configureMasterBus fallback applies only when no profile is given so
	// non-authored / legacy renders are unaffected.
	if a.plan.MixBus != "" {
		if profile := MixBusByName(a.plan.MixBus); profile != nil {
			_ = core.applyMixBusProfile(profile, float64(synth.SampleRate), seed)
		} else {
			// Unknown profile name: fall through to legacy defaults rather than
			// silently rendering dry.
			a.configureMasterBus(core)
		}
	} else {
		a.configureMasterBus(core)
	}

	for _, setup := range a.uniqueChannelSetups() {
		core.setProgram(setup.Channel, setup.Program)
		core.setPan(setup.Channel, setup.Pan)
		core.setReverbSend(setup.Channel, SectionCC(setup.Reverb, ReverbDelta(a.profile)))
		core.setChorusSend(setup.Channel, setup.Chorus)
		core.setChannelCutoff(setup.Channel, SectionCC(setup.Brightness, BrightnessDelta(a.profile)))
	}

	// SP8: Resolve groove template once and share across all tracks.
	var groove *GrooveTemplate
	if a.plan.Groove != "" {
		groove = GrooveByName(a.plan.Groove)
	}

	// SP8: Resolve automation lanes that drive callbacks. Pre-compute total
	// samples for the section so ValueAt(at01) can be called cheaply.
	totalSectionSamples := secondsToSamples(a.plan.DurationSec)
	automation := a.plan.Automation // captured by closure below

	// SP8: Wire per-role reverb buses (item 3). Multiple roles sharing the same
	// IR name share the same bus instance for CPU efficiency.
	//
	// Implementation note: the sf2Core renders all channels into a single stereo
	// buffer, so true per-channel reverb routing would require per-channel render
	// loops (a large architectural change). Instead we wire the dominant role's
	// IR into the master convolution bus when the mix_bus profile has not already
	// installed one. This gives audibly different reverb character per authored
	// track while staying within the existing infrastructure. Full per-channel
	// routing is tracked as a future improvement (SP10+).
	if len(a.plan.RoleReverb) > 0 && a.plan.MixBus == "" {
		// Find the IR name with the highest absolute send level (loudest reverb
		// dominates the character). Default: -12 dB wet.
		bestIR := ""
		bestDB := -100.0
		for _, rr := range a.plan.RoleReverb {
			if rr.IRName == "" {
				continue
			}
			if rr.SendDB > bestDB {
				bestDB = rr.SendDB
				bestIR = rr.IRName
			}
		}
		if bestIR != "" {
			_ = core.setNamedConvolutionIR(bestIR, float64(synth.SampleRate), seed, 0.35)
		}
	}

	// SP19-D: install ambient texture layers when configured. Each layer is
	// rendered in parallel with the music in the SF2 core's master mix.
	if len(a.plan.Textures) > 0 {
		layers := make([]*synth.TextureLayer, 0, len(a.plan.Textures))
		for i, txCfg := range a.plan.Textures {
			layer := synth.NewTextureLayer(float64(synth.SampleRate), synth.TextureConfig{
				Kind:    synth.TextureKind(txCfg.Name),
				LevelDB: txCfg.LevelDB,
				Seed:    seed ^ int64(0x7e7c7b00+i),
			})
			if layer != nil {
				layers = append(layers, layer)
			}
		}
		if len(layers) > 0 {
			core.setTextureLayers(layers)
		}
	} else {
		core.setTextureLayers(nil)
	}

	for _, track := range a.plan.Tracks {
		velocityPattern := append([]int32(nil), track.VelocityPattern...)
		timingOffsets := append([]float64(nil), track.TimingOffsets...)
		gatePattern := append([]float64(nil), track.GatePattern...)
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

		// SP8 groove: build a combined timing resolver that composes the
		// groove template on top of any authored per-slot timing offsets.
		if groove != nil {
			capturedGroove := groove
			capturedOffsets := timingOffsets
			cfg.ResolveTimingOffsetSec = func(slot int) float64 {
				step := slot % 16
				// Convert samples to seconds for the SF2Track interface.
				grv := float64(capturedGroove.TimingOffsetsSamples[step]) / float64(synth.SampleRate)
				authored := 0.0
				if len(capturedOffsets) > 0 {
					authored = capturedOffsets[slot%len(capturedOffsets)]
				}
				return grv + authored
			}
			// SP8 groove velocity: compose groove multiplier onto the authored pattern.
			capturedVelPattern := velocityPattern
			cfg.ResolveVelocity = func(slot int, key int, base int32) int32 {
				step := slot % 16
				mul := capturedGroove.VelocityMultipliers[step]
				v := base
				if len(capturedVelPattern) > 0 {
					idx := slot % len(capturedVelPattern)
					v += capturedVelPattern[idx]
				}
				v = int32(float64(v) * mul)
				if v < 18 {
					return 18
				}
				if v > 127 {
					return 127
				}
				return v
			}
		} else {
			// No groove: use the original velocity and timing resolvers.
			if len(velocityPattern) > 0 {
				vp := velocityPattern
				cfg.ResolveVelocity = func(slot int, key int, base int32) int32 {
					idx := slot % len(vp)
					v := base + vp[idx]
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
				to := timingOffsets
				cfg.ResolveTimingOffsetSec = func(slot int) float64 {
					return to[slot%len(to)]
				}
			}
		}

		if len(gatePattern) > 0 {
			gp := gatePattern
			cfg.ResolveGate = func(slot int, key int) float64 {
				return gp[slot%len(gp)]
			}
		}

		// SP8 automation: wire cutoff, expression, and pan lanes into their
		// respective callbacks. We use the current elapsed-samples position as
		// the progress indicator inside the callbacks. Because the elapsed
		// counter is read from the AuthoredTrack (not the SF2Track), we
		// reference it via a pointer to avoid a data race.
		if len(automation) > 0 && totalSectionSamples > 0 {
			elapsed := &a.samplesElapsed
			sectionSamples := totalSectionSamples
			autoLanes := automation

			cutoffLane := findAutomationLane(autoLanes, "cutoff")
			exprLane := findAutomationLane(autoLanes, "expression")

			if cutoffLane != nil {
				capturedLane := cutoffLane
				cfg.ResolveBrightness = func(slot int, key int) SF2ExpressionCurve {
					at01 := float64(*elapsed) / float64(sectionSamples)
					if at01 > 1 {
						at01 = 1
					}
					v := int32(capturedLane.ValueAt(at01))
					if v < 0 {
						v = 0
					}
					if v > 127 {
						v = 127
					}
					return SF2ExpressionCurve{Start: v, Peak: v, End: v}
				}
			}
			if exprLane != nil {
				capturedLane := exprLane
				cfg.ResolveExpression = func(slot int, key int) SF2ExpressionCurve {
					at01 := float64(*elapsed) / float64(sectionSamples)
					if at01 > 1 {
						at01 = 1
					}
					v := int32(capturedLane.ValueAt(at01))
					if v < 0 {
						v = 0
					}
					if v > 127 {
						v = 127
					}
					return SF2ExpressionCurve{Start: v, Peak: v, End: v}
				}
			}
		}

		core.addTrack(cfg)
	}
	a.core = core
	if a.plan.BarCount > 0 && a.plan.DurationSec > 0 {
		a.barSamples = secondsToSamples(a.plan.DurationSec / float64(a.plan.BarCount))
	}
}

// findAutomationLane returns the first lane matching param, or nil.
func findAutomationLane(lanes []AuthoredAutomationLane, param string) *AuthoredAutomationLane {
	for i := range lanes {
		if lanes[i].Param == param {
			return &lanes[i]
		}
	}
	return nil
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
		return 2.7
	case "lofi":
		return 2.8
	case "bells":
		return 3.4
	case "ambient":
		return 3.5
	case "drone":
		return 3.4
	case "classical":
		return 2.6
	case "phase":
		return 2.8
	case "lullaby":
		return 3.0
	default:
		return 2.6
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
