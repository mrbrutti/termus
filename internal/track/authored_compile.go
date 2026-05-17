package track

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mrbrutti/termus/internal/gen"
)

const authoredSlotsPerBar = 8

type authoredHarmonyBar struct {
	chords []authoredChord
}

type authoredChord struct {
	Label    string
	RootPC   int
	BassPC   int
	HasBass  bool
	Kind     string
	Scale    []int
	Degrees  map[int]int
	Interval []int
}

type authoredRoleTemplate struct {
	Channel    int32
	Program    int32
	Velocity   int32
	Pan        int32
	Reverb     int32
	Chorus     int32
	Brightness int32
	Gate       float64
	Swing      float64
	Legato     bool
	TieRepeats bool
	OverlapSec float64
}

type authoredSectionContext struct {
	style     string
	sectionID string
	variation string
	scene     string
	profile   gen.ControlProfile
	rng       *rand.Rand
}

func ctxForSection(style string, section Section, profile gen.ControlProfile, seed int64) authoredSectionContext {
	return authoredSectionContext{
		style:     style,
		sectionID: firstNonBlank(section.Title, section.ID),
		variation: strings.ToLower(strings.TrimSpace(section.Variation)),
		scene:     strings.ToLower(strings.TrimSpace(section.Scene)),
		profile:   profile,
		rng:       rand.New(rand.NewSource(seed)),
	}
}

func buildAuthoredPlan(spec gen.AlgoSpec, file *File, section Section, roles map[string]Role, dur time.Duration, profile gen.ControlProfile, seed int64) (gen.AuthoredTrackPlan, error) {
	bpm := resolveTempoBPM(firstNonBlank(section.Tempo, file.Tempo), spec.Name)
	barSec := 240.0 / bpm
	harmonyBars, err := parseHarmonyBars(section.Harmony)
	if err != nil {
		return gen.AuthoredTrackPlan{}, err
	}
	targetBars := maxInt(len(harmonyBars), int(math.Round(dur.Seconds()/barSec)))
	if targetBars <= 0 {
		targetBars = 1
	}
	if len(harmonyBars) == 0 {
		keyRoot := keyRootPitchClass(firstNonBlank(section.Key, file.Key))
		harmonyBars = []authoredHarmonyBar{{chords: []authoredChord{{Label: firstNonBlank(section.Harmony, firstNonBlank(section.Key, file.Key), "I"), RootPC: keyRoot, Kind: "maj", Scale: []int{0, 2, 4, 5, 7, 9, 11}}}}}
	}
	harmonyBars = repeatHarmonyBars(harmonyBars, targetBars)
	slotCount := targetBars * authoredSlotsPerBar
	phraseSpans := compilePhraseSpans(ctxForSection(spec.Name, section, profile, seed), targetBars)
	plan := gen.AuthoredTrackPlan{
		Style:       spec.Name,
		Section:     firstNonBlank(section.Title, section.ID),
		Key:         firstNonBlank(section.Key, file.Key),
		Tempo:       firstNonBlank(section.Tempo, file.Tempo),
		BPM:         bpm,
		DurationSec: dur.Seconds(),
		BarCount:    targetBars,
		SlotCount:   slotCount,
		ChordSpans:  compileChordSpans(harmonyBars),
		PhraseSpans: phraseSpans,
	}
	ctx := ctxForSection(spec.Name, section, profile, seed)
	roleNames := make([]string, 0, len(roles))
	for name, role := range roles {
		active := role.Active == nil || *role.Active
		if !active {
			continue
		}
		if strings.TrimSpace(role.Family) == "" && strings.TrimSpace(role.Pattern) == "" && strings.TrimSpace(role.Motif) == "" {
			continue
		}
		roleNames = append(roleNames, name)
	}
	sort.Strings(roleNames)
	for _, roleName := range roleNames {
		role := roles[roleName]
		rendered, err := compileRoleTracks(ctx, roleName, role, harmonyBars, targetBars, phraseSpans)
		if err != nil {
			return gen.AuthoredTrackPlan{}, fmt.Errorf("%s: %w", roleName, err)
		}
		plan.Tracks = append(plan.Tracks, rendered...)
	}
	applyPhraseDynamics(ctx, &plan)
	applySectionEvents(ctx, sectionEvents(section), harmonyBars, &plan)
	if len(plan.Tracks) == 0 {
		return gen.AuthoredTrackPlan{}, fmt.Errorf("no active authored role tracks compiled")
	}
	return plan, nil
}

func resolveTempoBPM(raw, style string) float64 {
	if bpm, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err == nil && bpm > 20 {
		return bpm
	}
	switch style {
	case "lofi":
		return 78
	case "jazz":
		return 126
	case "classical":
		return 92
	case "bells":
		return 54
	case "ambient":
		return 58
	case "drone":
		return 46
	case "phase":
		return 74
	case "lullaby":
		return 68
	default:
		return 80
	}
}

func (c authoredSectionContext) descriptor() string {
	return strings.TrimSpace(strings.Join([]string{c.variation, c.scene, c.sectionID}, " "))
}

func (c authoredSectionContext) has(parts ...string) bool {
	text := c.descriptor()
	for _, part := range parts {
		if part != "" && strings.Contains(text, strings.ToLower(strings.TrimSpace(part))) {
			return true
		}
	}
	return false
}

func (c authoredSectionContext) densityBias() int {
	bias := gen.ProfileCentered(c.profile.Density)
	switch {
	case c.has("sparse", "thin", "subtract", "break"):
		bias -= 2
	case c.has("busy", "lift", "open", "chorus", "drive"):
		bias += 1
	}
	return bias
}

func (c authoredSectionContext) motionBias() int {
	bias := gen.ProfileCentered(c.profile.Motion)
	switch {
	case c.has("still", "settle", "cadence", "outro"):
		bias -= 1
	case c.has("moving", "drive", "pulse", "sequence", "glide"):
		bias += 1
	}
	return bias
}

func (c authoredSectionContext) registerShift() int {
	switch {
	case c.has("open-register", "lift-register", "bright", "chorus", "air"):
		return 12
	case c.has("cadence", "outro", "settle", "home", "close"):
		return -12
	default:
		return 0
	}
}

func (c authoredSectionContext) shouldThin(slot int) bool {
	if c.densityBias() >= 0 {
		return false
	}
	if c.has("establish", "intro", "hush", "sparse") {
		return slot%4 == 1 || slot%8 == 6
	}
	return slot%4 == 3
}

func (c authoredSectionContext) shouldLift() bool {
	return c.registerShift() > 0
}

func compilePhraseSpans(ctx authoredSectionContext, totalBars int) []gen.AuthoredPhraseSpan {
	if totalBars <= 0 {
		return nil
	}
	phraseBars := phraseBlockSize(ctx.style, totalBars)
	if phraseBars <= 0 {
		phraseBars = totalBars
	}
	count := int(math.Ceil(float64(totalBars) / float64(phraseBars)))
	out := make([]gen.AuthoredPhraseSpan, 0, count)
	for startBar := 1; startBar <= totalBars; startBar += phraseBars {
		endBar := minInt(totalBars, startBar+phraseBars-1)
		idx := len(out)
		label := "statement"
		switch {
		case idx == 0:
			label = "statement"
		case endBar == totalBars && ctx.has("cadence", "outro", "settle", "close"):
			label = "cadence"
		case endBar == totalBars:
			label = "release"
		case idx%2 == 1:
			label = "answer"
		default:
			label = "sequence"
		}
		out = append(out, gen.AuthoredPhraseSpan{
			StartBar: startBar,
			EndBar:   endBar,
			Label:    label,
		})
	}
	return out
}

func phraseBlockSize(style string, totalBars int) int {
	if totalBars <= 2 {
		return totalBars
	}
	switch style {
	case "ambient", "drone":
		if totalBars >= 8 {
			return 4
		}
		return 2
	case "bells", "lullaby", "phase":
		return 2
	default:
		if totalBars >= 8 {
			return 4
		}
		return 2
	}
}

func phraseSpanForSlot(spans []gen.AuthoredPhraseSpan, slot int) (gen.AuthoredPhraseSpan, int) {
	bar := slot/authoredSlotsPerBar + 1
	for idx, span := range spans {
		if bar >= span.StartBar && bar <= span.EndBar {
			return span, idx
		}
	}
	if len(spans) == 0 {
		return gen.AuthoredPhraseSpan{StartBar: 1, EndBar: maxInt(1, bar), Label: "statement"}, 0
	}
	return spans[len(spans)-1], len(spans) - 1
}

func eventSlotRange(event Event, totalBars int) (int, int) {
	bar := event.Bar
	if bar <= 0 {
		bar = totalBars
	}
	if bar > totalBars {
		bar = totalBars
	}
	start := (bar - 1) * authoredSlotsPerBar
	if event.Slot > 0 {
		start += clampInt(event.Slot-1, 0, authoredSlotsPerBar-1)
	}
	spanBars := event.Bars
	if spanBars <= 0 {
		spanBars = 1
	}
	end := start + spanBars*authoredSlotsPerBar
	maxSlots := totalBars * authoredSlotsPerBar
	if end > maxSlots {
		end = maxSlots
	}
	if start < 0 {
		start = 0
	}
	return start, end
}

func eventRoleSelected(event Event, trackName string) bool {
	if len(event.Roles) == 0 {
		return true
	}
	base := baseRoleName(trackName)
	for _, role := range event.Roles {
		if strings.EqualFold(strings.TrimSpace(role), base) || strings.EqualFold(strings.TrimSpace(role), trackName) {
			return true
		}
	}
	return false
}

func muteSlots(track *gen.AuthoredRenderTrack, start, end int) {
	if track == nil {
		return
	}
	limit := minInt(len(track.Notes), end)
	for i := maxInt(0, start); i < limit; i++ {
		track.Notes[i] = -1
		if i < len(track.VelocityPattern) {
			track.VelocityPattern[i] = 0
		}
		if i < len(track.TimingOffsets) {
			track.TimingOffsets[i] = 0
		}
	}
}

func applyFill(track *gen.AuthoredRenderTrack, ctx authoredSectionContext, event Event, start, end int) {
	base := baseRoleName(track.Name)
	pattern := strings.TrimSpace(event.Pattern)
	if pattern == "" {
		pattern = defaultFillPattern(base)
	}
	notes := compileDrumPattern(ctx, base, Role{Family: "drums", Pattern: pattern}, maxInt(1, (end-start+authoredSlotsPerBar-1)/authoredSlotsPerBar), nil)
	for i := start; i < end && i < len(track.Notes); i++ {
		local := i - start
		if local >= 0 && local < len(notes) {
			track.Notes[i] = notes[local]
			if i < len(track.VelocityPattern) && track.Notes[i] >= 0 {
				track.VelocityPattern[i] += 8
			}
		}
	}
}

func applyPickup(track *gen.AuthoredRenderTrack, ctx authoredSectionContext, bars []authoredHarmonyBar, event Event, start, end int) {
	motif := strings.TrimSpace(event.Motif)
	if motif == "" {
		motif = defaultPickupMotif(baseRoleName(track.Name))
	}
	tokens := normalizeMelodyTokens(strings.Fields(strings.TrimSpace(strings.ReplaceAll(motif, "|", " "))))
	if len(tokens) == 0 {
		return
	}
	writeStart := maxInt(start, end-len(tokens))
	last := -1
	for i := writeStart; i < end && i < len(track.Notes); i++ {
		token := strings.TrimSpace(tokens[i-writeStart])
		if token == "" || token == "." || token == "r" {
			track.Notes[i] = -1
			continue
		}
		if token == "-" {
			if last >= 0 {
				track.Notes[i] = last
			}
			continue
		}
		chord := chordForSlot(bars, i)
		center := trackCenterFromRenderTrack(*track, ctx.style)
		note := melodyTokenToMidi(chord, token, center, last)
		track.Notes[i] = note
		last = note
		if i < len(track.VelocityPattern) {
			track.VelocityPattern[i] += 10
		}
		if i < len(track.TimingOffsets) {
			track.TimingOffsets[i] = -0.006
		}
	}
}

func applyStab(track *gen.AuthoredRenderTrack, ctx authoredSectionContext, event Event, start, end int) {
	pattern := strings.TrimSpace(event.Pattern)
	if pattern == "" {
		pattern = "x... ...."
	}
	grid := expandRhythmPattern(pattern, maxInt(1, (end-start+authoredSlotsPerBar-1)/authoredSlotsPerBar), "x... ....")
	last := -1
	for i := start; i < end && i < len(track.Notes); i++ {
		local := i - start
		if local >= 0 && local < len(grid) && grid[local] {
			if track.Notes[i] < 0 && last >= 0 {
				track.Notes[i] = last
			}
			if track.Notes[i] >= 0 {
				last = track.Notes[i]
				if i < len(track.VelocityPattern) {
					track.VelocityPattern[i] += 6
				}
			}
			continue
		}
		if track.Notes[i] >= 0 {
			last = track.Notes[i]
		}
		track.Notes[i] = -1
	}
}

func applyPedal(track *gen.AuthoredRenderTrack, start, end int) {
	held := -1
	for i := maxInt(0, start); i < minInt(len(track.Notes), end); i++ {
		if track.Notes[i] >= 0 {
			held = track.Notes[i]
			break
		}
	}
	if held < 0 {
		return
	}
	for i := maxInt(0, start); i < minInt(len(track.Notes), end); i++ {
		track.Notes[i] = held
		if i < len(track.VelocityPattern) {
			track.VelocityPattern[i] -= 4
		}
		if i < len(track.TimingOffsets) {
			track.TimingOffsets[i] = 0
		}
	}
}

func applySwell(track *gen.AuthoredRenderTrack, start, end int) {
	span := maxInt(1, end-start)
	for i := maxInt(0, start); i < minInt(len(track.Notes), end); i++ {
		if track.Notes[i] < 0 || i >= len(track.VelocityPattern) {
			continue
		}
		progress := float64(i-start) / float64(span)
		track.VelocityPattern[i] += int32(4 + progress*14)
	}
}

func nextAvailableChannel(used map[int32]bool) int32 {
	for channel := int32(0); channel < 16; channel++ {
		if channel == 9 || used[channel] {
			continue
		}
		return channel
	}
	return -1
}

func applyDouble(track gen.AuthoredRenderTrack, channel int32, start, end int) gen.AuthoredRenderTrack {
	clone := track
	clone.Name = track.Name + "-double"
	if channel >= 0 {
		clone.Channel = channel
	}
	clone.Pan = clampInt32(track.Pan+10, 0, 127)
	clone.Velocity = clampInt32(track.Velocity-6, 0, 127)
	clone.Notes = append([]int(nil), track.Notes...)
	clone.VelocityPattern = append([]int32(nil), track.VelocityPattern...)
	clone.TimingOffsets = append([]float64(nil), track.TimingOffsets...)
	for i := 0; i < len(clone.Notes); i++ {
		if i < start || i >= end || clone.Notes[i] < 0 {
			clone.Notes[i] = -1
			continue
		}
		switch authoredRoleKind(track.Name, Role{Family: track.Family}) {
		case "bass":
			clone.Notes[i] -= 12
		default:
			clone.Notes[i] += 12
		}
	}
	return clone
}

func applyTag(track *gen.AuthoredRenderTrack, start, end int) {
	if start <= 0 || start >= len(track.Notes) {
		return
	}
	sourceStart := maxInt(0, start-authoredSlotsPerBar)
	for i := maxInt(0, start); i < minInt(len(track.Notes), end); i++ {
		src := sourceStart + (i-start)%maxInt(1, authoredSlotsPerBar)
		if src >= 0 && src < len(track.Notes) {
			track.Notes[i] = track.Notes[src]
			if i < len(track.VelocityPattern) && src < len(track.VelocityPattern) {
				track.VelocityPattern[i] = track.VelocityPattern[src] + 6
			}
		}
	}
}

func applyEnding(track *gen.AuthoredRenderTrack, start, end int) {
	cut := start + maxInt(1, (end-start)/2)
	for i := maxInt(0, cut); i < minInt(len(track.Notes), end); i++ {
		if i%2 == 1 {
			track.Notes[i] = -1
			continue
		}
		if i < len(track.VelocityPattern) {
			track.VelocityPattern[i] -= 8
		}
	}
}

func phraseSlotRange(span gen.AuthoredPhraseSpan) (int, int) {
	start := maxInt(0, (span.StartBar-1)*authoredSlotsPerBar)
	end := maxInt(start, span.EndBar*authoredSlotsPerBar)
	return start, end
}

func applyPhraseDynamics(ctx authoredSectionContext, plan *gen.AuthoredTrackPlan) {
	if plan == nil || len(plan.PhraseSpans) == 0 {
		return
	}
	for idx := range plan.Tracks {
		track := &plan.Tracks[idx]
		role := Role{
			Family:       track.Family,
			Articulation: track.Articulation,
			Prominence:   track.Prominence,
		}
		kind := authoredRoleKind(track.Name, role)
		for _, span := range plan.PhraseSpans {
			start, end := phraseSlotRange(span)
			applyPhraseArc(ctx, track, kind, span.Label, start, end)
		}
	}
}

func applyPhraseArc(ctx authoredSectionContext, track *gen.AuthoredRenderTrack, kind, label string, start, end int) {
	if track == nil || start >= end {
		return
	}
	limit := minInt(len(track.Notes), end)
	spanLen := maxInt(1, limit-start)
	for i := maxInt(0, start); i < limit; i++ {
		if track.Notes[i] < 0 {
			continue
		}
		progress := float64(i-start) / float64(spanLen)
		switch strings.ToLower(strings.TrimSpace(label)) {
		case "statement":
			bump := int32((0.5 - math.Abs(progress-0.5)) * 10)
			track.VelocityPattern[i] += bump
		case "answer":
			track.VelocityPattern[i] -= 2
			if progress > 0.35 && progress < 0.75 {
				track.VelocityPattern[i] += 3
			}
		case "sequence":
			track.VelocityPattern[i] += int32(progress * 8)
			if i < len(track.TimingOffsets) {
				track.TimingOffsets[i] -= 0.002
			}
		case "release":
			track.VelocityPattern[i] -= int32(progress * 8)
			if i < len(track.TimingOffsets) {
				track.TimingOffsets[i] += 0.002
			}
		case "cadence":
			if progress >= 0.55 {
				track.VelocityPattern[i] += int32((1.0 - progress) * 8)
			} else {
				track.VelocityPattern[i] += int32(progress * 6)
			}
		}
	}
	if kind == "melody" && strings.ToLower(strings.TrimSpace(label)) != "cadence" {
		for i := limit - 1; i >= maxInt(0, start); i-- {
			if track.Notes[i] < 0 {
				continue
			}
			track.Notes[i] = -1
			if i < len(track.VelocityPattern) {
				track.VelocityPattern[i] = 0
			}
			break
		}
	}
	if strings.EqualFold(label, "cadence") || strings.EqualFold(label, "release") {
		for i := limit - 1; i >= maxInt(0, start); i-- {
			if track.Notes[i] < 0 {
				continue
			}
			if i < len(track.GatePattern) {
				switch kind {
				case "comp", "pad":
					track.GatePattern[i] = maxFloat(track.GatePattern[i], 1.18)
				case "melody", "bass":
					track.GatePattern[i] = maxFloat(track.GatePattern[i], 1.02)
				}
			}
			break
		}
	}
	if kind == "drum" && strings.EqualFold(label, "cadence") {
		for i := maxInt(0, limit-2); i < limit; i++ {
			if track.Notes[i] >= 0 {
				track.VelocityPattern[i] += 5
			}
		}
	}
}

func applyVelocityRamp(track *gen.AuthoredRenderTrack, start, end int, from, to int32) {
	if track == nil || start >= end {
		return
	}
	span := maxInt(1, end-start-1)
	for i := maxInt(0, start); i < minInt(len(track.Notes), end); i++ {
		if track.Notes[i] < 0 || i >= len(track.VelocityPattern) {
			continue
		}
		progress := float64(i-start) / float64(span)
		ramp := from + int32(float64(to-from)*progress)
		track.VelocityPattern[i] += ramp
	}
}

func applyBreath(track *gen.AuthoredRenderTrack, start, end int) {
	if track == nil {
		return
	}
	cutStart := maxInt(start, end-2)
	muteSlots(track, cutStart, end)
}

func applyHold(track *gen.AuthoredRenderTrack, start, end int) {
	if track == nil {
		return
	}
	applyPedal(track, start, end)
	for i := maxInt(0, start); i < minInt(len(track.GatePattern), end); i++ {
		track.GatePattern[i] = maxFloat(track.GatePattern[i], 1.18)
	}
}

func defaultFillPattern(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "kick":
		return "x.xxxxxx"
	case "snare":
		return "..x.xx.x"
	case "ride":
		return "xxxxxxxx"
	case "hat", "hihat":
		return "xxxxxxxx"
	case "rim":
		return "...x.x.x"
	default:
		return "x..xx..x"
	}
}

func defaultPickupMotif(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "trumpet", "alto", "tenor", "clarinet", "lead":
		return "5 6 7 9"
	default:
		return "3 5 6 7"
	}
}

func trackCenterFromRenderTrack(track gen.AuthoredRenderTrack, style string) int {
	return roleRegisterCenter(track.Register, style, baseRoleName(track.Name))
}

func applySectionEvents(ctx authoredSectionContext, events []Event, bars []authoredHarmonyBar, plan *gen.AuthoredTrackPlan) {
	if plan == nil || len(plan.Tracks) == 0 || len(events) == 0 {
		return
	}
	for _, event := range events {
		start, end := eventSlotRange(event, plan.BarCount)
		if start >= end {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(event.Kind)) {
		case "drop":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				muteSlots(&plan.Tracks[idx], start, end)
			}
		case "stop":
			for idx := range plan.Tracks {
				if eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				muteSlots(&plan.Tracks[idx], start, end)
			}
		case "fill":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				if authoredRoleKind(plan.Tracks[idx].Name, Role{Family: plan.Tracks[idx].Family}) == "drum" {
					applyFill(&plan.Tracks[idx], ctx, event, start, end)
				}
			}
		case "break":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				muteSlots(&plan.Tracks[idx], start, end)
			}
		case "pickup":
			for idx := range plan.Tracks {
				roleKind := authoredRoleKind(plan.Tracks[idx].Name, Role{Family: plan.Tracks[idx].Family})
				if roleKind != "melody" {
					continue
				}
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyPickup(&plan.Tracks[idx], ctx, bars, event, start, end)
			}
		case "stab":
			for idx := range plan.Tracks {
				roleKind := authoredRoleKind(plan.Tracks[idx].Name, Role{Family: plan.Tracks[idx].Family})
				if roleKind != "comp" && roleKind != "pad" {
					continue
				}
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyStab(&plan.Tracks[idx], ctx, event, start, end)
			}
		case "pedal":
			for idx := range plan.Tracks {
				roleKind := authoredRoleKind(plan.Tracks[idx].Name, Role{Family: plan.Tracks[idx].Family})
				if roleKind != "bass" && roleKind != "pad" && roleKind != "comp" {
					continue
				}
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyPedal(&plan.Tracks[idx], start, end)
			}
		case "swell":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applySwell(&plan.Tracks[idx], start, end)
			}
		case "crescendo":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyVelocityRamp(&plan.Tracks[idx], start, end, 0, 16)
			}
		case "decrescendo":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyVelocityRamp(&plan.Tracks[idx], start, end, 0, -16)
			}
		case "breath":
			for idx := range plan.Tracks {
				roleKind := authoredRoleKind(plan.Tracks[idx].Name, Role{Family: plan.Tracks[idx].Family})
				if roleKind != "melody" && roleKind != "comp" && roleKind != "pad" {
					continue
				}
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyBreath(&plan.Tracks[idx], start, end)
			}
		case "hold":
			for idx := range plan.Tracks {
				roleKind := authoredRoleKind(plan.Tracks[idx].Name, Role{Family: plan.Tracks[idx].Family})
				if roleKind != "melody" && roleKind != "comp" && roleKind != "pad" && roleKind != "bass" {
					continue
				}
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyHold(&plan.Tracks[idx], start, end)
			}
		case "silence":
			for idx := range plan.Tracks {
				if len(event.Roles) > 0 && !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				muteSlots(&plan.Tracks[idx], start, end)
			}
		case "double":
			extra := make([]gen.AuthoredRenderTrack, 0, len(plan.Tracks))
			usedChannels := map[int32]bool{}
			for _, existing := range plan.Tracks {
				usedChannels[existing.Channel] = true
			}
			for idx := range plan.Tracks {
				roleKind := authoredRoleKind(plan.Tracks[idx].Name, Role{Family: plan.Tracks[idx].Family})
				if roleKind == "drum" {
					continue
				}
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				channel := nextAvailableChannel(usedChannels)
				if channel >= 0 {
					usedChannels[channel] = true
				}
				doubled := applyDouble(plan.Tracks[idx], channel, start, end)
				extra = append(extra, doubled)
			}
			plan.Tracks = append(plan.Tracks, extra...)
		case "tag":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyTag(&plan.Tracks[idx], start, end)
			}
		case "ending":
			for idx := range plan.Tracks {
				if !eventRoleSelected(event, plan.Tracks[idx].Name) {
					continue
				}
				applyEnding(&plan.Tracks[idx], start, end)
			}
		}
	}
}

func parseHarmonyBars(src string) ([]authoredHarmonyBar, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, nil
	}
	parts := strings.Split(src, "|")
	out := make([]authoredHarmonyBar, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		bar := authoredHarmonyBar{chords: make([]authoredChord, 0, len(fields))}
		for _, token := range fields {
			chord, ok := parseAuthoredChord(token)
			if !ok {
				return nil, fmt.Errorf("invalid chord %q", token)
			}
			bar.chords = append(bar.chords, chord)
		}
		out = append(out, bar)
	}
	return out, nil
}

func parseAuthoredChord(token string) (authoredChord, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return authoredChord{}, false
	}
	main := token
	bassPart := ""
	if idx := strings.Index(token, "/"); idx >= 0 {
		main = strings.TrimSpace(token[:idx])
		bassPart = strings.TrimSpace(token[idx+1:])
	}
	root, rest, ok := parseRootToken(main)
	if !ok {
		return authoredChord{}, false
	}
	lower := strings.ToLower(rest)
	chord := authoredChord{Label: token, RootPC: root, BassPC: root}
	if bassPart != "" {
		if bassRoot, _, ok := parseRootToken(bassPart); ok {
			chord.BassPC = bassRoot
			chord.HasBass = true
		}
	}
	degrees := map[int]int{1: 0}
	switch {
	case strings.Contains(lower, "m7b5") || strings.Contains(lower, "ø"):
		chord.Kind = "half-dim"
		chord.Scale = []int{0, 1, 3, 5, 6, 8, 10}
		degrees[3], degrees[5], degrees[7] = 3, 6, 10
	case strings.Contains(lower, "dim"):
		chord.Kind = "dim"
		chord.Scale = []int{0, 2, 3, 5, 6, 8, 9}
		degrees[3], degrees[5], degrees[7] = 3, 6, 9
	case strings.Contains(lower, "sus"):
		chord.Kind = "sus"
		chord.Scale = []int{0, 2, 5, 7, 9, 10}
		degrees[3], degrees[5], degrees[7] = 5, 7, 10
	case strings.Contains(lower, "maj"):
		chord.Kind = "maj"
		chord.Scale = []int{0, 2, 4, 5, 7, 9, 11}
		degrees[3], degrees[5], degrees[7] = 4, 7, 11
	case strings.Contains(lower, "m"):
		chord.Kind = "min"
		chord.Scale = []int{0, 2, 3, 5, 7, 9, 10}
		degrees[3], degrees[5], degrees[7] = 3, 7, 10
	default:
		chord.Kind = "dom"
		chord.Scale = []int{0, 2, 4, 5, 7, 9, 10}
		degrees[3], degrees[5], degrees[7] = 4, 7, 10
	}
	if strings.Contains(lower, "#5") || strings.Contains(lower, "aug") {
		degrees[5] = 8
	}
	if strings.Contains(lower, "b5") && chord.Kind != "half-dim" {
		degrees[5] = 6
	}
	if strings.Contains(lower, "sus2") {
		degrees[3] = 2
	}
	if strings.Contains(lower, "maj6") || (strings.Contains(lower, "6") && !strings.Contains(lower, "16") && !strings.Contains(lower, "13")) {
		degrees[13] = 9
	}
	if strings.Contains(lower, "9") {
		degrees[9] = 14
	}
	if strings.Contains(lower, "b9") {
		degrees[9] = 13
	}
	if strings.Contains(lower, "#9") {
		degrees[9] = 15
	}
	if strings.Contains(lower, "11") || strings.Contains(lower, "sus") {
		degrees[11] = 17
	}
	if strings.Contains(lower, "#11") {
		degrees[11] = 18
	}
	if strings.Contains(lower, "b11") {
		degrees[11] = 16
	}
	if strings.Contains(lower, "13") {
		degrees[13] = 21
	}
	if strings.Contains(lower, "b13") {
		degrees[13] = 20
	}
	if _, ok := degrees[13]; ok {
		if _, has9 := degrees[9]; !has9 {
			degrees[9] = 14
		}
	}
	if _, ok := degrees[11]; ok {
		if _, has9 := degrees[9]; !has9 {
			degrees[9] = 14
		}
	}
	if strings.Contains(lower, "add9") {
		degrees[9] = 14
	}
	chord.Degrees = degrees
	chord.Interval = sortedChordDegrees(degrees)
	return chord, true
}

func parseRootToken(token string) (int, string, bool) {
	if token == "" {
		return 0, "", false
	}
	rootMap := map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}
	root, ok := rootMap[token[0]]
	if !ok {
		return 0, "", false
	}
	rest := token[1:]
	if len(rest) > 0 {
		switch rest[0] {
		case 'b':
			root--
			rest = rest[1:]
		case '#':
			root++
			rest = rest[1:]
		}
	}
	return wrapPitchClass(root), rest, true
}

func repeatHarmonyBars(base []authoredHarmonyBar, bars int) []authoredHarmonyBar {
	if len(base) == 0 || bars <= 0 {
		return nil
	}
	out := make([]authoredHarmonyBar, bars)
	for i := 0; i < bars; i++ {
		src := base[i%len(base)]
		dst := authoredHarmonyBar{chords: make([]authoredChord, len(src.chords))}
		copy(dst.chords, src.chords)
		out[i] = dst
	}
	return out
}

func compileChordSpans(bars []authoredHarmonyBar) []gen.AuthoredChordSpan {
	spans := make([]gen.AuthoredChordSpan, 0, len(bars)*2)
	for barIdx, bar := range bars {
		if len(bar.chords) == 0 {
			continue
		}
		perChord := authoredSlotsPerBar / len(bar.chords)
		rem := authoredSlotsPerBar % len(bar.chords)
		slot := barIdx * authoredSlotsPerBar
		for i, chord := range bar.chords {
			width := perChord
			if i < rem {
				width++
			}
			if width <= 0 {
				width = 1
			}
			spans = append(spans, gen.AuthoredChordSpan{
				StartSlot: slot,
				EndSlot:   slot + width,
				Label:     chord.Label,
			})
			slot += width
		}
	}
	return spans
}

func compileRoleTracks(ctx authoredSectionContext, name string, role Role, bars []authoredHarmonyBar, totalBars int, phraseSpans []gen.AuthoredPhraseSpan) ([]gen.AuthoredRenderTrack, error) {
	template := authoredTemplateFor(ctx.style, name, role)
	kind := authoredRoleKind(name, role)
	switch kind {
	case "drum":
		notes := compileDrumPattern(ctx, name, role, totalBars, phraseSpans)
		if len(notes) == 0 {
			return nil, nil
		}
		return []gen.AuthoredRenderTrack{{
			Name:            name,
			Family:          role.Family,
			Tone:            append([]string(nil), role.Tone...),
			Articulation:    role.Articulation,
			Register:        role.Register,
			Prominence:      role.Prominence,
			Channel:         template.Channel,
			Program:         template.Program,
			Velocity:        template.Velocity,
			Pan:             template.Pan,
			Reverb:          template.Reverb,
			Chorus:          template.Chorus,
			Brightness:      template.Brightness,
			Notes:           notes,
			VelocityPattern: compileVelocityPattern(ctx, kind, name, role, notes, phraseSpans),
			TimingOffsets:   compileTimingOffsets(ctx, kind, name, role, notes, phraseSpans),
			GatePattern:     compileGatePattern(ctx, kind, name, role, notes, phraseSpans, template.Gate),
			Gate:            template.Gate,
			SwingAmount:     template.Swing,
			Legato:          false,
			TieRepeats:      false,
			OverlapSec:      0,
			FireProbability: 1,
		}}, nil
	case "bass":
		notes := compileBassLine(ctx, name, role, bars, totalBars, phraseSpans)
		return []gen.AuthoredRenderTrack{{
			Name:            name,
			Family:          role.Family,
			Tone:            append([]string(nil), role.Tone...),
			Articulation:    role.Articulation,
			Register:        role.Register,
			Prominence:      role.Prominence,
			Channel:         template.Channel,
			Program:         template.Program,
			Velocity:        template.Velocity,
			Pan:             template.Pan,
			Reverb:          template.Reverb,
			Chorus:          template.Chorus,
			Brightness:      template.Brightness,
			Notes:           notes,
			VelocityPattern: compileVelocityPattern(ctx, kind, name, role, notes, phraseSpans),
			TimingOffsets:   compileTimingOffsets(ctx, kind, name, role, notes, phraseSpans),
			GatePattern:     compileGatePattern(ctx, kind, name, role, notes, phraseSpans, template.Gate),
			Gate:            template.Gate,
			SwingAmount:     template.Swing,
			Legato:          true,
			TieRepeats:      true,
			OverlapSec:      template.OverlapSec,
			FireProbability: 1,
		}}, nil
	case "pad":
		voices := compilePadVoices(ctx, name, role, bars, totalBars, phraseSpans)
		return authoredVoiceTracks(ctx, name, role, template, voices, phraseSpans, true), nil
	case "comp":
		voices := compileCompVoices(ctx, name, role, bars, totalBars, phraseSpans)
		return authoredVoiceTracks(ctx, name, role, template, voices, phraseSpans, false), nil
	default: // melody
		notes := compileMelody(ctx, name, role, bars, totalBars, phraseSpans)
		return []gen.AuthoredRenderTrack{{
			Name:            name,
			Family:          role.Family,
			Tone:            append([]string(nil), role.Tone...),
			Articulation:    role.Articulation,
			Register:        role.Register,
			Prominence:      role.Prominence,
			Channel:         template.Channel,
			Program:         template.Program,
			Velocity:        template.Velocity,
			Pan:             template.Pan,
			Reverb:          template.Reverb,
			Chorus:          template.Chorus,
			Brightness:      template.Brightness,
			Notes:           notes,
			VelocityPattern: compileVelocityPattern(ctx, kind, name, role, notes, phraseSpans),
			TimingOffsets:   compileTimingOffsets(ctx, kind, name, role, notes, phraseSpans),
			GatePattern:     compileGatePattern(ctx, kind, name, role, notes, phraseSpans, template.Gate),
			Gate:            template.Gate,
			SwingAmount:     template.Swing,
			Legato:          template.Legato,
			TieRepeats:      template.TieRepeats,
			OverlapSec:      template.OverlapSec,
			FireProbability: 1,
		}}, nil
	}
}

func authoredRoleKind(name string, role Role) string {
	lowerName := strings.ToLower(name)
	family := strings.ToLower(role.Family)
	switch lowerName {
	case "kick", "snare", "hat", "hihat", "ride", "crash", "openhat", "clap", "rim", "tom", "tom-low", "tom-high", "perc":
		return "drum"
	}
	if family == "drums" {
		return "drum"
	}
	if strings.Contains(lowerName, "bass") || family == "bass" || family == "synth_bass" {
		return "bass"
	}
	if family == "pad" || family == "choir" || family == "strings" {
		return "pad"
	}
	if role.Motif != "" || strings.Contains(lowerName, "lead") || strings.Contains(lowerName, "bell") {
		return "melody"
	}
	if family == "reed_lead" || family == "woodwind" || family == "brass" || family == "bells" || family == "music_box" {
		return "melody"
	}
	return "comp"
}

func authoredTemplateFor(style, name string, role Role) authoredRoleTemplate {
	family := strings.ToLower(role.Family)
	lowerName := strings.ToLower(name)
	base := authoredRoleTemplate{
		Velocity:   88,
		Pan:        64,
		Reverb:     54,
		Chorus:     18,
		Brightness: 86,
		Gate:       0.92,
	}
	assigned := false
	apply := func(channel, program, velocity, pan, reverb, chorus, brightness int32, gate float64) {
		base.Channel = channel
		base.Program = program
		base.Velocity = velocity
		base.Pan = pan
		base.Reverb = reverb
		base.Chorus = chorus
		base.Brightness = brightness
		base.Gate = gate
		assigned = true
	}
	if authoredRoleKind(name, role) == "drum" || family == "drums" {
		base.Channel = 9
		base.Program = 0
		base.Pan = 64
		base.Reverb = 24
		base.Chorus = 0
		base.Brightness = 72
		base.Gate = 0.48
		base.Velocity = 80
		switch style {
		case "lofi":
			base.Reverb, base.Brightness, base.Velocity, base.Gate, base.Swing = 12, 56, 76, 0.42, 0.08
		case "jazz":
			base.Reverb, base.Brightness, base.Velocity, base.Gate, base.Swing = 28, 78, 82, 0.38, 0.16
		default:
			base.Reverb, base.Brightness, base.Velocity, base.Gate = 18, 70, 78, 0.44
		}
		return base
	}
	switch style {
	case "lofi":
		base.Swing = 0.10
		switch lowerName {
		case "keys", "rhodes", "ep", "chords":
			apply(0, 5, 74, 64, 42, 42, 52, 0.72)
		case "bass", "sub":
			apply(1, 32, 84, 64, 18, 0, 50, 0.92)
		case "texture", "vibes", "vibraphone", "mallet":
			apply(2, 11, 62, 92, 68, 24, 78, 0.82)
		case "lead", "sax", "hook", "counter", "flute":
			apply(3, 65, 80, 40, 58, 0, 78, 0.88)
			base.Legato, base.TieRepeats, base.OverlapSec = true, true, 0.012
		case "guitar", "pluck":
			apply(4, 24, 68, 88, 38, 20, 64, 0.68)
		case "pad", "choir":
			apply(5, 89, 54, 72, 74, 12, 68, 1.10)
			base.Legato, base.TieRepeats = true, true
		}
	case "jazz":
		base.Swing = 0.16
		switch lowerName {
		case "keys", "piano", "comp":
			apply(0, 0, 78, 64, 44, 0, 88, 0.62)
		case "bass", "walk":
			apply(1, 32, 82, 64, 16, 0, 64, 0.90)
		case "lead", "sax", "horn", "alto", "tenor", "clarinet":
			apply(2, 66, 82, 44, 48, 0, 88, 0.82)
			base.Legato, base.TieRepeats, base.OverlapSec = true, true, 0.010
		case "trumpet":
			apply(2, 56, 84, 46, 52, 0, 92, 0.80)
			base.Legato, base.TieRepeats, base.OverlapSec = true, true, 0.008
		case "guitar":
			apply(3, 24, 70, 84, 34, 12, 70, 0.66)
		case "vibes", "vibraphone":
			apply(3, 11, 72, 80, 52, 16, 84, 0.76)
		case "organ":
			apply(4, 19, 68, 74, 36, 18, 82, 0.92)
		}
	case "bells":
		base.Swing = 0
		switch lowerName {
		case "bells":
			apply(0, 14, 72, 64, 98, 0, 104, 0.86)
		case "celesta":
			apply(1, 8, 66, 80, 92, 0, 96, 0.78)
		case "glock":
			apply(2, 9, 64, 46, 86, 0, 98, 0.74)
		case "box", "music_box":
			apply(3, 10, 62, 86, 92, 0, 92, 0.76)
		case "pad":
			apply(4, 89, 58, 64, 84, 0, 72, 1.20)
			base.Legato, base.TieRepeats = true, true
		case "choir":
			apply(5, 52, 58, 64, 94, 0, 76, 1.10)
			base.Legato, base.TieRepeats = true, true
		case "strings":
			apply(6, 48, 56, 74, 72, 0, 76, 1.12)
			base.Legato, base.TieRepeats = true, true
		case "bass":
			apply(7, 32, 64, 64, 28, 0, 60, 1.12)
		case "shimmer":
			apply(8, 88, 52, 94, 98, 0, 90, 0.84)
		}
	case "ambient":
		base.Swing = 0
		switch lowerName {
		case "pad":
			apply(0, 89, 56, 64, 92, 0, 76, 1.25)
			base.Legato, base.TieRepeats = true, true
		case "choir":
			apply(1, 52, 54, 76, 98, 0, 74, 1.20)
			base.Legato, base.TieRepeats = true, true
		case "texture", "bells", "sparkle":
			apply(2, 14, 50, 88, 104, 0, 88, 0.92)
		case "lead", "flute", "woodwind":
			apply(3, 73, 64, 40, 82, 0, 82, 0.96)
			base.Legato, base.TieRepeats = true, true
		case "bass":
			apply(4, 39, 66, 64, 28, 0, 60, 1.18)
		case "strings":
			apply(5, 48, 52, 52, 82, 0, 74, 1.16)
			base.Legato, base.TieRepeats = true, true
		case "shimmer":
			apply(6, 88, 48, 94, 102, 0, 90, 0.88)
		}
	case "drone":
		base.Swing = 0
		switch lowerName {
		case "bed":
			apply(0, 89, 56, 64, 98, 0, 70, 1.30)
			base.Legato, base.TieRepeats = true, true
		case "strings":
			apply(1, 48, 52, 76, 94, 0, 74, 1.24)
			base.Legato, base.TieRepeats = true, true
		case "choir":
			apply(2, 52, 50, 52, 102, 0, 72, 1.24)
			base.Legato, base.TieRepeats = true, true
		case "shimmer":
			apply(3, 88, 48, 92, 108, 0, 90, 0.94)
		case "bass":
			apply(4, 39, 62, 64, 24, 0, 58, 1.18)
		case "lead":
			apply(5, 73, 56, 40, 94, 0, 80, 1.02)
			base.Legato, base.TieRepeats = true, true
		}
	case "classical":
		base.Swing = 0
		switch lowerName {
		case "piano":
			apply(0, 0, 82, 64, 44, 0, 90, 0.86)
		case "strings":
			apply(1, 48, 66, 72, 74, 0, 82, 1.08)
			base.Legato, base.TieRepeats = true, true
		case "winds":
			apply(2, 71, 64, 48, 64, 0, 86, 0.92)
		case "brass":
			apply(3, 61, 62, 80, 68, 0, 82, 0.98)
		case "harp":
			apply(4, 46, 64, 86, 70, 0, 84, 0.76)
		case "choir":
			apply(5, 52, 54, 52, 84, 0, 72, 1.10)
			base.Legato, base.TieRepeats = true, true
		}
	case "phase":
		base.Swing = 0
		switch lowerName {
		case "mallet-a", "mallet_a":
			apply(0, 11, 72, 46, 72, 0, 96, 0.62)
		case "mallet-b", "mallet_b":
			apply(1, 11, 72, 82, 72, 0, 96, 0.62)
		case "pad":
			apply(2, 89, 54, 64, 84, 0, 74, 1.16)
			base.Legato, base.TieRepeats = true, true
		case "bass":
			apply(3, 39, 68, 64, 24, 0, 64, 1.12)
		case "shimmer":
			apply(4, 14, 50, 90, 96, 0, 88, 0.86)
		case "choir":
			apply(5, 52, 50, 56, 96, 0, 74, 1.14)
			base.Legato, base.TieRepeats = true, true
		}
	case "lullaby":
		base.Swing = 0.02
		switch lowerName {
		case "lead":
			apply(0, 10, 70, 64, 82, 0, 88, 0.84)
		case "harp":
			apply(1, 46, 64, 84, 86, 0, 82, 0.76)
		case "choir":
			apply(2, 52, 56, 52, 94, 0, 78, 1.10)
			base.Legato, base.TieRepeats = true, true
		case "box":
			apply(3, 10, 62, 76, 92, 0, 90, 0.82)
		case "pad":
			apply(4, 89, 50, 64, 88, 0, 72, 1.14)
			base.Legato, base.TieRepeats = true, true
		}
	}
	if assigned {
		switch family {
		case "acoustic_piano":
			base.Program = 0
			base.Chorus = 0
			if lowerName == "keys" || lowerName == "comp" || lowerName == "chords" {
				base.Brightness = maxInt32(base.Brightness, 84)
			}
		case "electric_piano":
			base.Program = 5
			base.Chorus = maxInt32(base.Chorus, 24)
		case "woodwind":
			base.Program = 73
			base.Legato, base.TieRepeats = true, true
			if base.OverlapSec == 0 {
				base.OverlapSec = 0.010
			}
		case "reed_lead":
			if lowerName == "alto" || lowerName == "tenor" || lowerName == "sax" || lowerName == "lead" || lowerName == "hook" || lowerName == "counter" {
				base.Program = 66
			}
			base.Legato, base.TieRepeats = true, true
			if base.OverlapSec == 0 {
				base.OverlapSec = 0.010
			}
		case "brass":
			base.Program = 56
			base.Brightness = maxInt32(base.Brightness, 88)
			base.Gate = maxFloat(base.Gate, 0.78)
		case "guitar":
			if strings.Contains(lowerName, "lead") || lowerName == "hook" || lowerName == "counter" {
				base.Program = 26
				base.Channel = 4
				base.Pan = 84
				base.Legato, base.TieRepeats = true, true
				if base.OverlapSec == 0 {
					base.OverlapSec = 0.008
				}
			} else {
				base.Program = 24
			}
		case "mallet":
			base.Program = 11
			if strings.Contains(lowerName, "lead") || lowerName == "vibes" || lowerName == "vibraphone" {
				base.Channel = maxInt32(base.Channel, 2)
				base.Legato, base.TieRepeats = false, false
				base.Gate = minFloat(base.Gate, 0.82)
			}
		case "bells":
			base.Program = 14
		case "music_box":
			base.Program = 10
		case "pad":
			base.Program = 89
			base.Legato, base.TieRepeats = true, true
			base.Gate = maxFloat(base.Gate, 1.10)
		case "choir":
			base.Program = 52
			base.Legato, base.TieRepeats = true, true
			base.Gate = maxFloat(base.Gate, 1.06)
		case "strings":
			if lowerName == "harp" {
				base.Program = 46
				base.Gate = minFloat(base.Gate, 0.80)
			} else {
				base.Program = 48
				base.Legato, base.TieRepeats = true, true
				base.Gate = maxFloat(base.Gate, 1.04)
			}
		case "synth_bass":
			base.Program = 39
			base.Gate = maxFloat(base.Gate, 1.08)
		case "bass":
			base.Program = 32
		}
	}
	if !assigned {
		switch family {
		case "acoustic_piano":
			apply(0, 0, 80, 64, 44, 0, 88, 0.80)
		case "electric_piano":
			apply(0, 5, 74, 64, 42, 28, 56, 0.76)
		case "bass":
			apply(1, 32, 82, 64, 18, 0, 60, 0.92)
		case "synth_bass":
			apply(1, 39, 72, 64, 18, 0, 62, 1.12)
		case "guitar":
			apply(4, 24, 70, 84, 38, 18, 68, 0.70)
		case "mallet":
			apply(2, 11, 66, 84, 72, 0, 86, 0.78)
		case "bells":
			apply(0, 14, 68, 64, 92, 0, 98, 0.80)
		case "music_box":
			apply(3, 10, 62, 82, 92, 0, 90, 0.80)
		case "pad":
			apply(4, 89, 54, 64, 86, 0, 74, 1.18)
			base.Legato, base.TieRepeats = true, true
		case "choir":
			apply(5, 52, 54, 56, 92, 0, 76, 1.16)
			base.Legato, base.TieRepeats = true, true
		case "strings":
			apply(6, 48, 60, 72, 76, 0, 80, 1.12)
			base.Legato, base.TieRepeats = true, true
		case "woodwind":
			apply(2, 73, 70, 44, 58, 0, 84, 0.90)
			base.Legato, base.TieRepeats = true, true
		case "reed_lead":
			apply(2, 66, 78, 44, 54, 0, 86, 0.84)
			base.Legato, base.TieRepeats = true, true
		case "brass":
			apply(3, 61, 76, 54, 58, 0, 84, 0.86)
		case "lead":
			apply(3, 88, 68, 48, 64, 0, 88, 0.88)
			base.Legato, base.TieRepeats = true, true
		default:
			apply(0, 0, 76, 64, 44, 0, 84, 0.82)
		}
	}
	base = applyRoleCharacter(base, role)
	return base
}

func applyRoleCharacter(base authoredRoleTemplate, role Role) authoredRoleTemplate {
	for _, tone := range role.Tone {
		switch strings.ToLower(strings.TrimSpace(tone)) {
		case "warm", "woody":
			base.Brightness -= 8
		case "dusty", "soft":
			base.Brightness -= 12
			base.Velocity -= 4
			base.Chorus += 6
		case "direct", "tight":
			base.Brightness += 8
			base.Gate = minFloat(base.Gate, 0.70)
		case "breathy":
			base.Reverb += 10
			base.Legato = true
			base.TieRepeats = true
		case "glass", "bright":
			base.Brightness += 10
		case "wide":
			base.Reverb += 8
			base.Pan = minInt32(base.Pan+6, 96)
		}
	}
	switch strings.ToLower(strings.TrimSpace(role.Articulation)) {
	case "stab", "pocket", "answer":
		base.Gate = minFloat(base.Gate, 0.68)
	case "lyrical":
		base.Gate = maxFloat(base.Gate, 0.92)
		base.Legato = true
		base.TieRepeats = true
		if base.OverlapSec == 0 {
			base.OverlapSec = 0.012
		}
	case "sustain", "halo":
		base.Gate = maxFloat(base.Gate, 1.12)
		base.Legato = true
		base.TieRepeats = true
	}
	switch strings.ToLower(strings.TrimSpace(role.Prominence)) {
	case "air":
		base.Velocity -= 10
		base.Reverb += 12
	case "lead":
		base.Velocity += 6
		base.Brightness += 4
	case "support":
		base.Velocity -= 4
	case "anchor":
		base.Pan = 64
	}
	base.Velocity = clampInt32(base.Velocity, 28, 118)
	base.Brightness = clampInt32(base.Brightness, 34, 118)
	base.Reverb = clampInt32(base.Reverb, 0, 118)
	base.Chorus = clampInt32(base.Chorus, 0, 118)
	return base
}

func compileDrumPattern(ctx authoredSectionContext, roleName string, role Role, totalBars int, phraseSpans []gen.AuthoredPhraseSpan) []int {
	grid := expandPhraseRhythmPattern(role, totalBars, phraseSpans, defaultRhythmPattern(roleName))
	out := make([]int, len(grid))
	for i, active := range grid {
		span, phraseIdx := phraseSpanForSlot(phraseSpans, i)
		localRole := roleForPhrase(role, span.Label)
		mode := rolePhraseMode(ctx, "drum", roleName, localRole, span, phraseIdx)
		active = phraseRhythmActive(ctx, authoredRoleKind(roleName, localRole), roleName, localRole, span, phraseIdx, i, active)
		active, note := drumSlotEvent(ctx, roleName, span, phraseIdx, totalBars, i, mode, active)
		if active {
			out[i] = note
		} else {
			out[i] = -1
		}
	}
	return out
}

func drumSlotEvent(ctx authoredSectionContext, roleName string, span gen.AuthoredPhraseSpan, phraseIdx, totalBars, slot int, mode string, active bool) (bool, int) {
	note := drumNoteFor(ctx, roleName, slot)
	lower := strings.ToLower(strings.TrimSpace(roleName))
	bar := slot / authoredSlotsPerBar
	pos := slot % authoredSlotsPerBar
	barInPhrase := bar - maxInt(0, span.StartBar-1)
	isTurnaroundBar := bar == maxInt(0, span.EndBar-1)
	isLastBar := bar == maxInt(0, totalBars-1)
	switch lower {
	case "kick":
		if isTurnaroundBar && (pos == 6 || pos == 7) {
			active = true
			note = 35
		}
		if mode == "sparse" && pos == 4 {
			active = false
		}
		if isLastBar && pos == 0 {
			active = true
			note = 36
		}
	case "snare", "clap", "rim":
		if ctx.style == "jazz" && (mode == "answer" || mode == "backbeat") && (pos == 3 || pos == 7) && barInPhrase%2 == 0 {
			active = true
			note = 37
		}
		if isTurnaroundBar && pos >= 6 {
			active = true
			note = 38
		}
		if isLastBar && pos == 7 {
			active = true
			note = 38
		}
	case "hat", "hihat":
		if ctx.style == "jazz" && (mode == "grid" || mode == "lift") {
			active = false
		}
		if mode == "answer" || mode == "air" || mode == "thin" {
			active = pos%2 == 1
		}
		if isTurnaroundBar && pos >= 6 {
			active = true
			note = 46
		}
	case "ride":
		if mode == "air" || mode == "thin" {
			active = false
		}
		if ctx.style == "jazz" && (mode == "grid" || mode == "lift" || isTurnaroundBar || isLastBar) {
			active = pos%2 == 0 || pos == 7
		}
		if isTurnaroundBar && pos == 7 {
			note = 53
		}
	case "crash":
		active = (isTurnaroundBar || isLastBar) && pos == 0
		note = 49
	default:
		if isTurnaroundBar && pos >= 6 {
			active = true
		}
	}
	if ctx.style == "lofi" && lower == "snare" && mode == "answer" && pos == 3 {
		active = true
		note = 37
	}
	return active, note
}

func compileBassLine(ctx authoredSectionContext, name string, role Role, bars []authoredHarmonyBar, totalBars int, phraseSpans []gen.AuthoredPhraseSpan) []int {
	grid := expandPhraseRhythmPattern(role, totalBars, phraseSpans, defaultRhythmPattern(name))
	out := make([]int, len(grid))
	last := -1
	for slot := range grid {
		span, phraseIdx := phraseSpanForSlot(phraseSpans, slot)
		localRole := roleForPhrase(role, span.Label)
		mode := rolePhraseMode(ctx, "bass", name, localRole, span, phraseIdx)
		if !phraseRhythmActive(ctx, "bass", name, localRole, span, phraseIdx, slot, grid[slot]) {
			out[slot] = -1
			continue
		}
		bar := slot / authoredSlotsPerBar
		pos := slot % authoredSlotsPerBar
		chord := chordForSlot(bars, slot)
		nextChord := chordForSlot(bars, minInt(totalBars*authoredSlotsPerBar-1, slot+authoredSlotsPerBar))
		rootPC := chord.RootPC
		if chord.HasBass && pos <= 1 {
			rootPC = chord.BassPC
		}
		base := rootMidiForRegister(rootPC, role.Register, ctx.style, name)
		nextBase := rootMidiForRegister(nextChord.RootPC, role.Register, ctx.style, name)
		lastBarInPhrase := bar == maxInt(0, span.EndBar-1)
		note := base
		switch {
		case strings.Contains(strings.ToLower(localRole.Family), "synth_bass"):
			switch {
			case pos == 0:
				note = base
			case pos == 4:
				note = base + 12
			case lastBarInPhrase && pos >= 6:
				note = placePitchNear(approachMidi(nextBase, base), base+2)
			case pos >= 6 && ctx.motionBias() > 0:
				note = placePitchNear(base+7, base+7)
			default:
				note = base
			}
		case mode == "pedal":
			if pos == 0 {
				note = base
			} else if pos >= 6 {
				note = placePitchNear(base-12, base-8)
			} else {
				note = placePitchNear(base+chordDegreeInterval(chord, 5), base+7)
			}
		case mode == "cadence":
			switch {
			case pos == 0:
				note = base
			case pos == 4:
				note = placePitchNear(base+chordDegreeInterval(chord, 5), base+7)
			case pos >= 6:
				note = placePitchNear(approachMidi(nextBase, base), base-1)
			default:
				note = placePitchNear(base+chordDegreeInterval(chord, 3), base+4)
			}
		case mode == "walk":
			switch pos {
			case 0:
				note = base
			case 2:
				note = placePitchNear(base+chordDegreeInterval(chord, 3), base+4)
			case 4:
				note = placePitchNear(base+chordDegreeInterval(chord, 5), base+7)
			case 6:
				note = placePitchNear(approachMidi(nextBase, base), base+2)
			default:
				note = base
			}
		case pos == 0:
			note = base
		case pos >= 6:
			note = placePitchNear(approachMidi(nextBase, base), base+2)
		case mode == "answer" && pos%4 == 2:
			note = placePitchNear(base+chordDegreeInterval(chord, 7), base+10)
		case mode == "answer" && pos == 6:
			note = placePitchNear(approachMidi(nextBase, base), base+1)
		case mode == "walk" && pos == 4:
			note = placePitchNear(base+chordDegreeInterval(chord, 9), base+12)
		case pos%4 == 2:
			note = placePitchNear(base+chordDegreeInterval(chord, 5), base+7)
		default:
			note = placePitchNear(base+chordDegreeInterval(chord, 3), base+4)
		}
		if ctx.shouldLift() && pos == 4 && !strings.Contains(strings.ToLower(localRole.Family), "synth_bass") {
			note += 12
		}
		if ctx.has("cadence", "settle", "outro") && pos >= 6 {
			note = placePitchNear(base, base-3)
		}
		if last >= 0 && note == last && pos%2 == 1 {
			note += 12
		}
		out[slot] = note
		last = note
	}
	return out
}

func approachMidi(target, around int) int {
	if target >= around {
		return target - 1
	}
	return target + 1
}

func compilePadVoices(ctx authoredSectionContext, name string, role Role, bars []authoredHarmonyBar, totalBars int, phraseSpans []gen.AuthoredPhraseSpan) [][]int {
	grid := expandPhraseRhythmPattern(role, totalBars, phraseSpans, defaultRhythmPattern(name))
	maxVoices := 4
	kind := authoredRoleKind(name, role)
	voices := make([][]int, maxVoices)
	for i := range voices {
		voices[i] = make([]int, len(grid))
		for j := range voices[i] {
			voices[i][j] = -1
		}
	}
	for slot, active := range grid {
		span, phraseIdx := phraseSpanForSlot(phraseSpans, slot)
		localRole := roleForPhrase(role, span.Label)
		mode := rolePhraseMode(ctx, kind, name, localRole, span, phraseIdx)
		active = phraseRhythmActive(ctx, kind, name, localRole, span, phraseIdx, slot, active)
		if !active {
			continue
		}
		chord := chordForSlot(bars, slot)
		voicing := chordVoicing(ctx, name, localRole, chord, mode, phraseIdx)
		center := roleRegisterCenter(localRole.Register, ctx.style, name) + ctx.registerShift()/2
		for i := range voices {
			if i >= len(voicing) {
				continue
			}
			voices[i][slot] = placePitchNear(rootMidiForRegister(chord.RootPC, role.Register, ctx.style, name)+voicing[i], center+i*3)
		}
	}
	return voices
}

func compileCompVoices(ctx authoredSectionContext, name string, role Role, bars []authoredHarmonyBar, totalBars int, phraseSpans []gen.AuthoredPhraseSpan) [][]int {
	return compilePadVoices(ctx, name, role, bars, totalBars, phraseSpans)
}

func compileMelody(ctx authoredSectionContext, name string, role Role, bars []authoredHarmonyBar, totalBars int, phraseSpans []gen.AuthoredPhraseSpan) []int {
	tokens := expandPhraseMelodyPattern(role, totalBars, phraseSpans, defaultMelodyPattern(ctx.style, name))
	out := make([]int, len(tokens))
	center := roleRegisterCenter(role.Register, ctx.style, name) + ctx.registerShift()
	last := center
	for slot, token := range tokens {
		span, phraseIdx := phraseSpanForSlot(phraseSpans, slot)
		localRole := roleForPhrase(role, span.Label)
		mode := rolePhraseMode(ctx, "melody", name, localRole, span, phraseIdx)
		token = strings.TrimSpace(token)
		if (ctx.shouldThin(slot) || mode == "tail") && token != "-" && slot%authoredSlotsPerBar > 4 {
			out[slot] = -1
			continue
		}
		if token == "" || token == "." || token == "r" {
			out[slot] = -1
			continue
		}
		if token == "-" {
			out[slot] = last
			continue
		}
		chord := chordForSlot(bars, slot)
		token = transformMelodyToken(ctx, mode, phraseIdx, token, slot)
		note := melodyTokenToMidi(chord, token, center, last)
		if (ctx.has("cadence", "outro") || mode == "cadence") && slot%authoredSlotsPerBar >= 6 {
			note = minInt(note, last)
		}
		out[slot] = note
		last = note
	}
	return out
}

func compileVelocityPattern(ctx authoredSectionContext, kind, name string, role Role, notes []int, phraseSpans []gen.AuthoredPhraseSpan) []int32 {
	if len(notes) == 0 {
		return nil
	}
	out := make([]int32, len(notes))
	lowerName := strings.ToLower(name)
	for i, note := range notes {
		if note < 0 {
			continue
		}
		pos := i % authoredSlotsPerBar
		bar := (i / authoredSlotsPerBar) % 2
		span, phraseIdx := phraseSpanForSlot(phraseSpans, i)
		mode := rolePhraseMode(ctx, kind, name, roleForPhrase(role, span.Label), span, phraseIdx)
		delta := int32(ctx.rng.Intn(5) - 2)
		switch kind {
		case "drum":
			switch lowerName {
			case "kick":
				if pos == 0 || pos == 4 {
					delta += 10
				} else {
					delta += 2
				}
			case "snare", "clap":
				if pos >= 4 && pos <= 5 {
					delta += 12
				} else {
					delta -= 6
				}
			case "hat", "hihat", "ride":
				if pos%2 == 0 {
					delta -= 5
				}
				if note == 46 || note == 51 {
					delta += 6
				}
				if ctx.motionBias() > 0 {
					delta += 2
				}
			default:
				delta += 3
			}
			switch mode {
			case "cadence", "fill":
				delta += 5
			case "thin", "air":
				delta -= 4
			case "push":
				delta += 3
			}
		case "bass":
			if pos == 0 {
				delta += 8
			} else if pos >= 6 {
				delta += 4
			}
			if strings.Contains(lowerName, "sub") {
				delta += 2
			}
			switch mode {
			case "pedal":
				delta -= 3
			case "walk", "answer":
				delta += 2
			case "cadence":
				if pos >= 6 {
					delta += 6
				}
			}
		case "melody":
			if pos == 0 {
				delta -= 2
			}
			if pos == 4 || pos == 6 {
				delta += 8
			}
			if ctx.shouldLift() && bar == 1 {
				delta += 4
			}
			if ctx.has("cadence", "outro") && pos >= 6 {
				delta -= 6
			}
			switch mode {
			case "foreground", "climb":
				if pos == 4 || pos == 6 {
					delta += 4
				}
			case "answer", "echo":
				delta -= 3
			case "cadence":
				if pos >= 6 {
					delta += 5
				}
			case "tail":
				delta -= 5
			}
		default:
			if pos == 0 || pos == 4 {
				delta += 5
			}
			if ctx.has("thin", "hush", "breakdown") {
				delta -= 4
			}
			switch mode {
			case "stab":
				if pos == 0 || pos == 4 {
					delta += 4
				}
			case "hold", "bed":
				delta -= 2
			case "push", "lift":
				if pos >= 4 {
					delta += 3
				}
			case "answer":
				delta -= 1
			}
		}
		out[i] = delta
	}
	return out
}

func compileTimingOffsets(ctx authoredSectionContext, kind, name string, role Role, notes []int, phraseSpans []gen.AuthoredPhraseSpan) []float64 {
	if len(notes) == 0 {
		return nil
	}
	out := make([]float64, len(notes))
	lowerName := strings.ToLower(name)
	baseLate := 0.0
	if ctx.style == "lofi" {
		baseLate = 0.010
	}
	for i, note := range notes {
		if note < 0 {
			continue
		}
		pos := i % authoredSlotsPerBar
		span, phraseIdx := phraseSpanForSlot(phraseSpans, i)
		mode := rolePhraseMode(ctx, kind, name, roleForPhrase(role, span.Label), span, phraseIdx)
		switch kind {
		case "drum":
			switch lowerName {
			case "kick":
				out[i] = 0
			case "snare", "clap":
				out[i] = baseLate + 0.006
			case "hat", "hihat":
				if pos%2 == 1 {
					out[i] = -0.003
				} else {
					out[i] = 0.002
				}
			default:
				out[i] = 0.001
			}
			switch mode {
			case "fill", "push":
				out[i] -= 0.003
			case "thin", "air":
				out[i] += 0.002
			}
		case "bass":
			if ctx.motionBias() > 0 {
				out[i] = -0.002
			} else {
				out[i] = baseLate * 0.5
			}
			switch mode {
			case "pedal":
				out[i] += 0.004
			case "walk", "answer":
				out[i] -= 0.002
			case "cadence":
				if pos >= 6 {
					out[i] -= 0.003
				}
			}
		case "melody":
			if pos == 0 {
				out[i] = -0.004
			} else {
				out[i] = 0.001
			}
			switch mode {
			case "answer", "echo":
				out[i] += 0.003
			case "climb":
				out[i] -= 0.002
			case "cadence":
				if pos >= 6 {
					out[i] += 0.002
				}
			}
		default:
			if strings.Contains(lowerName, "guitar") || strings.Contains(lowerName, "piano") || strings.Contains(lowerName, "keys") {
				out[i] = 0.004
			}
			switch mode {
			case "stab", "push":
				out[i] -= 0.003
			case "hold", "bed":
				out[i] += 0.003
			case "answer":
				out[i] += 0.001
			}
		}
	}
	return out
}

func compileGatePattern(ctx authoredSectionContext, kind, name string, role Role, notes []int, phraseSpans []gen.AuthoredPhraseSpan, base float64) []float64 {
	if len(notes) == 0 {
		return nil
	}
	out := make([]float64, len(notes))
	for i, note := range notes {
		out[i] = base
		if note < 0 {
			continue
		}
		span, phraseIdx := phraseSpanForSlot(phraseSpans, i)
		mode := rolePhraseMode(ctx, kind, name, roleForPhrase(role, span.Label), span, phraseIdx)
		switch kind {
		case "drum":
			switch mode {
			case "fill":
				out[i] = minFloat(base, 0.46)
			case "lift", "grid":
				out[i] = minFloat(base, 0.52)
			case "thin", "air":
				out[i] = minFloat(base, 0.42)
			default:
				out[i] = minFloat(base, 0.58)
			}
		case "bass":
			switch mode {
			case "pedal":
				out[i] = maxFloat(base, 1.10)
			case "cadence":
				out[i] = maxFloat(minFloat(base, 0.92), 0.78)
			case "walk", "answer":
				out[i] = maxFloat(minFloat(base, 0.88), 0.74)
			default:
				out[i] = maxFloat(minFloat(base, 0.96), 0.80)
			}
		case "melody":
			switch mode {
			case "foreground", "climb":
				out[i] = maxFloat(minFloat(base, 0.88), 0.74)
			case "answer", "echo":
				out[i] = maxFloat(minFloat(base, 0.78), 0.62)
			case "cadence":
				out[i] = maxFloat(base, 0.98)
			case "tail":
				out[i] = maxFloat(minFloat(base, 0.66), 0.54)
			default:
				out[i] = base
			}
		default:
			switch mode {
			case "stab":
				out[i] = maxFloat(minFloat(base, 0.52), 0.36)
			case "push", "answer":
				out[i] = maxFloat(minFloat(base, 0.70), 0.48)
			case "hold", "bed", "close":
				out[i] = maxFloat(base, 1.08)
			case "lift":
				out[i] = maxFloat(base, 0.92)
			default:
				out[i] = base
			}
		}
	}
	return out
}

func rolePhraseMode(ctx authoredSectionContext, kind, name string, role Role, span gen.AuthoredPhraseSpan, phraseIdx int) string {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	prominence := strings.ToLower(strings.TrimSpace(role.Prominence))
	switch kind {
	case "melody":
		switch span.Label {
		case "cadence":
			return "cadence"
		case "release":
			return "tail"
		case "answer":
			if strings.Contains(prominence, "air") || strings.Contains(prominence, "support") {
				return "echo"
			}
			return "answer"
		case "sequence":
			return "climb"
		default:
			return "foreground"
		}
	case "bass":
		switch span.Label {
		case "cadence":
			return "cadence"
		case "release":
			return "pedal"
		case "answer":
			return "answer"
		case "sequence":
			return "walk"
		default:
			return "anchor"
		}
	case "pad":
		switch span.Label {
		case "cadence":
			return "close"
		case "release":
			return "hold"
		case "answer":
			if strings.Contains(prominence, "air") || lowerName == "texture" || lowerName == "shimmer" {
				return "echo"
			}
			return "answer"
		case "sequence":
			return "lift"
		default:
			return "bed"
		}
	case "comp":
		switch span.Label {
		case "cadence":
			return "stab"
		case "release":
			return "hold"
		case "answer":
			return "answer"
		case "sequence":
			return "push"
		default:
			return "support"
		}
	case "drum":
		switch lowerName {
		case "kick":
			switch span.Label {
			case "cadence":
				return "cadence"
			case "release":
				return "sparse"
			case "sequence":
				return "push"
			default:
				return "anchor"
			}
		case "snare", "clap", "rim":
			switch span.Label {
			case "cadence":
				return "fill"
			case "release":
				return "thin"
			case "answer", "sequence":
				return "answer"
			default:
				return "backbeat"
			}
		default:
			switch span.Label {
			case "cadence":
				return "lift"
			case "release":
				return "thin"
			case "answer":
				return "air"
			case "sequence":
				return "push"
			default:
				return "grid"
			}
		}
	default:
		return span.Label
	}
}

func phraseRhythmActive(ctx authoredSectionContext, kind, name string, role Role, span gen.AuthoredPhraseSpan, phraseIdx, slot int, active bool) bool {
	pos := slot % authoredSlotsPerBar
	lower := strings.ToLower(strings.TrimSpace(name))
	mode := rolePhraseMode(ctx, kind, name, role, span, phraseIdx)
	if !active {
		switch mode {
		case "fill":
			return pos >= 6
		case "walk":
			if kind == "bass" {
				return pos%2 == 0
			}
		case "answer":
			if kind == "bass" {
				return pos == 6
			}
		case "pedal":
			if kind == "bass" {
				return pos == 4 || pos == 6 || pos == 7
			}
		case "cadence":
			if kind == "bass" {
				return pos == 6 || pos == 7
			}
		case "lift":
			return (lower == "ride" || lower == "hat" || lower == "hihat") && pos%2 == 1
		case "stab":
			return pos == 0 || pos == 4
		default:
			return false
		}
	}
	switch mode {
	case "air":
		if kind == "drum" && (lower == "hat" || lower == "hihat" || lower == "ride") && pos%2 == 0 {
			return false
		}
		if kind == "pad" && pos >= 6 {
			return false
		}
	case "walk":
		if kind == "bass" {
			return pos%2 == 0
		}
	case "answer":
		if kind == "bass" {
			return pos == 0 || pos == 4 || pos == 6
		}
	case "pedal":
		if kind == "bass" {
			return pos == 0 || pos == 4 || pos == 6 || pos == 7
		}
	case "hold":
		if (kind == "comp" || kind == "pad") && pos != 0 && pos != 4 {
			return false
		}
	case "tail":
		if kind == "melody" && pos >= 6 {
			return false
		}
	case "close":
		if kind == "pad" && pos == 0 && phraseIdx > 0 {
			return false
		}
	case "cadence":
		if kind == "bass" {
			return pos == 0 || pos == 4 || pos == 6 || pos == 7
		}
		if kind == "drum" && (lower == "snare" || lower == "kick") && pos >= 6 {
			return true
		}
		if (kind == "comp" || kind == "pad") && pos >= 4 {
			return false
		}
	}
	return true
}

func transformMelodyToken(ctx authoredSectionContext, phraseMode string, phraseIdx int, token string, slot int) string {
	if token == "" || token == "." || token == "-" || token == "r" {
		return token
	}
	bar := slot / authoredSlotsPerBar
	pos := slot % authoredSlotsPerBar
	switch {
	case phraseMode == "answer" && pos <= 2:
		return shiftMelodyToken(token, 2, false)
	case phraseMode == "climb" && pos == 4:
		return shiftMelodyToken(token, 1, false)
	case phraseMode == "cadence" && pos >= 6:
		return shiftMelodyToken(token, -2, false)
	case phraseMode == "tail" && phraseIdx > 0 && pos >= 4:
		return shiftMelodyToken(token, -1, false)
	case phraseMode == "echo" && pos >= 4:
		return shiftMelodyToken(token, -3, false)
	case ctx.has("sequence-up", "answer-lift") && bar%2 == 1 && pos <= 2:
		return shiftMelodyToken(token, 2, false)
	case ctx.has("open-register", "lift-register") && pos == 4:
		return shiftMelodyToken(token, 0, true)
	case ctx.has("cadence", "outro", "settle") && pos >= 6:
		return shiftMelodyToken(token, -2, false)
	default:
		return token
	}
}

func shiftMelodyToken(token string, degreeDelta int, octaveUp bool) string {
	prefix := ""
	for strings.HasPrefix(token, ">") || strings.HasPrefix(token, "^") || strings.HasPrefix(token, "<") {
		prefix += token[:1]
		token = token[1:]
	}
	acc := ""
	for strings.HasPrefix(token, "b") || strings.HasPrefix(token, "#") {
		acc += token[:1]
		token = token[1:]
	}
	degree, err := strconv.Atoi(token)
	if err != nil {
		return prefix + acc + token
	}
	degree += degreeDelta
	if degree < 1 {
		degree = 1
	}
	if degree > 13 {
		degree = 13
	}
	if octaveUp {
		prefix = ">" + prefix
	}
	return prefix + acc + strconv.Itoa(degree)
}

func authoredVoiceTracks(ctx authoredSectionContext, name string, role Role, template authoredRoleTemplate, voices [][]int, phraseSpans []gen.AuthoredPhraseSpan, sustained bool) []gen.AuthoredRenderTrack {
	out := make([]gen.AuthoredRenderTrack, 0, len(voices))
	for idx, notes := range voices {
		if isAllRest(notes) {
			continue
		}
		out = append(out, gen.AuthoredRenderTrack{
			Name:            fmt.Sprintf("%s-%d", name, idx+1),
			Family:          role.Family,
			Tone:            append([]string(nil), role.Tone...),
			Articulation:    role.Articulation,
			Register:        role.Register,
			Prominence:      role.Prominence,
			Channel:         template.Channel,
			Program:         template.Program,
			Velocity:        template.Velocity - int32(idx*4),
			Pan:             template.Pan,
			Reverb:          template.Reverb,
			Chorus:          template.Chorus,
			Brightness:      template.Brightness,
			Notes:           notes,
			VelocityPattern: compileVelocityPattern(ctx, authoredRoleKind(name, role), name, role, notes, phraseSpans),
			TimingOffsets:   compileTimingOffsets(ctx, authoredRoleKind(name, role), name, role, notes, phraseSpans),
			GatePattern:     compileGatePattern(ctx, authoredRoleKind(name, role), name, role, notes, phraseSpans, template.Gate),
			Gate:            template.Gate,
			SwingAmount:     template.Swing,
			Legato:          sustained || template.Legato,
			TieRepeats:      sustained || template.TieRepeats,
			OverlapSec:      template.OverlapSec,
			FireProbability: 1,
		})
	}
	return out
}

func defaultRhythmPattern(name string) string {
	switch strings.ToLower(name) {
	case "kick":
		return "x... x..."
	case "snare":
		return ".... x..."
	case "hat", "hihat":
		return "x.x.x.x."
	case "ride":
		return "x.x. x.x."
	case "crash":
		return "x......."
	case "bass":
		return "x... x..."
	case "keys", "piano", "guitar":
		return "x..x .x.."
	default:
		return "x... ...."
	}
}

func defaultMelodyPattern(style, name string) string {
	switch {
	case strings.Contains(strings.ToLower(name), "bell"):
		return "5 . . 7 | 9 . 7 5"
	case style == "jazz":
		return "5 . 6 7 | 9 . 7 3"
	case style == "lofi":
		return "5 . . 7 | 9 . 7 5"
	default:
		return "5 . 3 . | 1 . . ."
	}
}

func roleForPhrase(role Role, label string) Role {
	if len(role.Phrases) == 0 {
		return role
	}
	block, ok := role.Phrases[strings.ToLower(strings.TrimSpace(label))]
	if !ok {
		return role
	}
	out := role
	if strings.TrimSpace(block.Pattern) != "" {
		out.Pattern = block.Pattern
	}
	if strings.TrimSpace(block.Motif) != "" {
		out.Motif = block.Motif
	}
	if strings.TrimSpace(block.Harmony) != "" {
		out.Harmony = block.Harmony
	}
	if block.Active != nil {
		out.Active = block.Active
	}
	return out
}

func expandPhraseRhythmPattern(role Role, totalBars int, phraseSpans []gen.AuthoredPhraseSpan, fallback string) []bool {
	if len(phraseSpans) == 0 {
		return expandRhythmPattern(role.Pattern, totalBars, fallback)
	}
	out := make([]bool, totalBars*authoredSlotsPerBar)
	for _, span := range phraseSpans {
		localRole := roleForPhrase(role, span.Label)
		if localRole.Active != nil && !*localRole.Active {
			continue
		}
		bars := maxInt(1, span.EndBar-span.StartBar+1)
		local := expandRhythmPattern(localRole.Pattern, bars, fallback)
		start := maxInt(0, (span.StartBar-1)*authoredSlotsPerBar)
		copy(out[start:], local)
	}
	return out
}

func expandPhraseMelodyPattern(role Role, totalBars int, phraseSpans []gen.AuthoredPhraseSpan, fallback string) []string {
	if len(phraseSpans) == 0 {
		return expandMelodyPattern(roleValue(role.Motif, role.Pattern), totalBars, fallback)
	}
	out := repeatString(".", totalBars*authoredSlotsPerBar)
	for _, span := range phraseSpans {
		localRole := roleForPhrase(role, span.Label)
		if localRole.Active != nil && !*localRole.Active {
			continue
		}
		bars := maxInt(1, span.EndBar-span.StartBar+1)
		local := expandMelodyPattern(roleValue(localRole.Motif, localRole.Pattern), bars, fallback)
		start := maxInt(0, (span.StartBar-1)*authoredSlotsPerBar)
		copy(out[start:], local)
	}
	return out
}

func expandRhythmPattern(pattern string, totalBars int, fallback string) []bool {
	if strings.TrimSpace(pattern) == "" {
		pattern = fallback
	}
	bars := strings.Split(pattern, "|")
	out := make([]bool, totalBars*authoredSlotsPerBar)
	for bar := 0; bar < totalBars; bar++ {
		src := ""
		if len(bars) > 0 {
			src = normalizeRhythmCells(bars[bar%len(bars)])
		}
		if src == "" {
			src = normalizeRhythmCells(fallback)
		}
		for i := 0; i < authoredSlotsPerBar && i < len(src); i++ {
			out[bar*authoredSlotsPerBar+i] = src[i] == 'x'
		}
	}
	return out
}

func normalizeRhythmCells(src string) string {
	fields := strings.Fields(strings.TrimSpace(src))
	joined := strings.Join(fields, "")
	switch len(joined) {
	case 0:
		return ""
	case 4:
		var b strings.Builder
		b.Grow(8)
		for _, ch := range joined {
			b.WriteRune(ch)
			b.WriteRune(ch)
		}
		return b.String()
	default:
		if len(joined) < authoredSlotsPerBar {
			var b strings.Builder
			for len(joined) < authoredSlotsPerBar {
				joined += "."
			}
			b.WriteString(joined[:authoredSlotsPerBar])
			return b.String()
		}
		return joined[:authoredSlotsPerBar]
	}
}

func expandMelodyPattern(pattern string, totalBars int, fallback string) []string {
	if strings.TrimSpace(pattern) == "" {
		pattern = fallback
	}
	bars := strings.Split(pattern, "|")
	out := make([]string, totalBars*authoredSlotsPerBar)
	for bar := 0; bar < totalBars; bar++ {
		src := ""
		if len(bars) > 0 {
			src = bars[bar%len(bars)]
		}
		tokens := strings.Fields(strings.TrimSpace(src))
		if len(tokens) == 0 {
			tokens = strings.Fields(strings.TrimSpace(fallback))
		}
		expanded := normalizeMelodyTokens(tokens)
		copy(out[bar*authoredSlotsPerBar:(bar+1)*authoredSlotsPerBar], expanded)
	}
	return out
}

func normalizeMelodyTokens(tokens []string) []string {
	if len(tokens) == 0 {
		return repeatString(".", authoredSlotsPerBar)
	}
	switch len(tokens) {
	case authoredSlotsPerBar:
		return append([]string(nil), tokens...)
	case 4:
		out := make([]string, 0, authoredSlotsPerBar)
		for _, token := range tokens {
			out = append(out, token, token)
		}
		return out
	default:
		out := make([]string, authoredSlotsPerBar)
		for i := 0; i < authoredSlotsPerBar; i++ {
			idx := i * len(tokens) / authoredSlotsPerBar
			if idx >= len(tokens) {
				idx = len(tokens) - 1
			}
			out[i] = tokens[idx]
		}
		return out
	}
}

func repeatString(value string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = value
	}
	return out
}

func roleValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func chordForSlot(bars []authoredHarmonyBar, slot int) authoredChord {
	bar := minInt(len(bars)-1, maxInt(0, slot/authoredSlotsPerBar))
	if len(bars[bar].chords) == 0 {
		return authoredChord{Label: "I", RootPC: 0, Kind: "maj", Scale: []int{0, 2, 4, 5, 7, 9, 11}}
	}
	pos := slot % authoredSlotsPerBar
	perChord := authoredSlotsPerBar / len(bars[bar].chords)
	rem := authoredSlotsPerBar % len(bars[bar].chords)
	start := 0
	for i, chord := range bars[bar].chords {
		width := perChord
		if i < rem {
			width++
		}
		if pos >= start && pos < start+width {
			return chord
		}
		start += width
	}
	return bars[bar].chords[len(bars[bar].chords)-1]
}

func chordVoicing(ctx authoredSectionContext, name string, role Role, chord authoredChord, phraseMode string, phraseIdx int) []int {
	lowerName := strings.ToLower(name)
	family := strings.ToLower(role.Family)
	switch authoredRoleKind(name, role) {
	case "pad":
		switch family {
		case "pad", "choir":
			if ctx.has("thin", "subtract", "breakdown") || phraseMode == "hold" {
				return []int{0, chordDegreeInterval(chord, 5), chordDegreeInterval(chord, 9)}
			}
			if phraseMode == "answer" || phraseMode == "echo" {
				return []int{0, chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 11)}
			}
			return []int{0, chordDegreeInterval(chord, 5), chordDegreeInterval(chord, 9), chordDegreeInterval(chord, 11)}
		case "strings":
			if phraseMode == "close" {
				return []int{0, chordDegreeInterval(chord, 5), chordDegreeInterval(chord, 9)}
			}
			return []int{0, chordDegreeInterval(chord, 3), chordDegreeInterval(chord, 5), chordDegreeInterval(chord, 9)}
		default:
			return []int{0, chordDegreeInterval(chord, 5), chordDegreeInterval(chord, 9)}
		}
	default:
		switch lowerName {
		case "guitar", "pluck":
			return []int{chordDegreeInterval(chord, 9), chordDegreeInterval(chord, 3), chordDegreeInterval(chord, 13)}
		case "piano", "keys", "rhodes", "ep", "comp", "organ":
			if family == "mallet" || strings.Contains(strings.ToLower(role.Prominence), "air") {
				return []int{chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 9)}
			}
			if ctx.has("thin", "hush", "breakdown") || phraseMode == "hold" {
				return []int{chordDegreeInterval(chord, 3), chordDegreeInterval(chord, 7)}
			}
			if phraseMode == "answer" {
				return []int{chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 9), chordDegreeInterval(chord, 13)}
			}
			if phraseMode == "stab" || phraseIdx%2 == 1 {
				return []int{chordDegreeInterval(chord, 3), chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 13)}
			}
			return []int{chordDegreeInterval(chord, 3), chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 9), chordDegreeInterval(chord, 13)}
		case "vibes", "vibraphone", "mallet":
			if ctx.has("thin", "air", "breakdown") || phraseMode == "hold" {
				return []int{chordDegreeInterval(chord, 9)}
			}
			if phraseMode == "answer" || phraseMode == "echo" {
				return []int{chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 9)}
			}
			return []int{chordDegreeInterval(chord, 3), chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 9)}
		default:
			return []int{chordDegreeInterval(chord, 3), chordDegreeInterval(chord, 7), chordDegreeInterval(chord, 9)}
		}
	}
}

func chordDegreeInterval(ch authoredChord, degree int) int {
	if interval, ok := ch.Degrees[degree]; ok {
		return interval
	}
	scale := ch.Scale
	switch degree {
	case 1:
		return 0
	case 3:
		return scaleInterval(scale, 2)
	case 5:
		return scaleInterval(scale, 4)
	case 7:
		return scaleInterval(scale, 6)
	case 9:
		return scaleInterval(scale, 1) + 12
	case 11:
		return scaleInterval(scale, 3) + 12
	case 13:
		return scaleInterval(scale, 5) + 12
	default:
		return 0
	}
}

func sortedChordDegrees(degrees map[int]int) []int {
	if len(degrees) == 0 {
		return nil
	}
	keys := make([]int, 0, len(degrees))
	for degree := range degrees {
		keys = append(keys, degree)
	}
	sort.Ints(keys)
	out := make([]int, 0, len(keys))
	for _, degree := range keys {
		out = append(out, degrees[degree])
	}
	return out
}

func scaleInterval(scale []int, idx int) int {
	if len(scale) == 0 {
		return 0
	}
	idx = idx % len(scale)
	return scale[idx]
}

func melodyTokenToMidi(ch authoredChord, token string, center, prev int) int {
	octaveShift := 0
	for strings.HasPrefix(token, ">") || strings.HasPrefix(token, "^") || strings.HasPrefix(token, "<") {
		switch token[0] {
		case '>', '^':
			octaveShift += 12
		case '<':
			octaveShift -= 12
		}
		token = token[1:]
	}
	accidental := 0
	for strings.HasPrefix(token, "b") || strings.HasPrefix(token, "#") {
		if token[0] == 'b' {
			accidental--
		} else {
			accidental++
		}
		token = token[1:]
	}
	degree, err := strconv.Atoi(token)
	if err != nil {
		return prev
	}
	base := rootMidiForCenter(ch.RootPC, center)
	interval := chordDegreeInterval(ch, degree)
	note := placePitchNear(base+interval+accidental+octaveShift, center)
	if prev > 0 && absInt(note-prev) > 9 {
		note = placePitchNear(base+interval+accidental+octaveShift, prev)
	}
	return note
}

func rootMidiForRegister(rootPC int, register, style, name string) int {
	center := roleRegisterCenter(register, style, name)
	return rootMidiForCenter(rootPC, center)
}

func rootMidiForCenter(rootPC, center int) int {
	note := 48 + rootPC
	return placePitchNear(note, center)
}

func roleRegisterCenter(register, style, name string) int {
	switch strings.ToLower(register) {
	case "sub":
		return 36
	case "low":
		return 48
	case "mid":
		return 60
	case "mid-high":
		return 72
	case "high":
		return 79
	case "air":
		return 84
	}
	switch strings.ToLower(name) {
	case "bass":
		return 44
	case "sub":
		return 36
	case "keys", "piano", "guitar", "organ", "rhodes", "ep":
		return 62
	case "lead", "sax", "horn", "bells", "flute", "clarinet", "trumpet", "vibes", "celesta", "glock", "box", "music_box":
		return 76
	case "texture", "choir", "pad", "strings", "shimmer":
		return 72
	default:
		switch style {
		case "lofi":
			return 68
		case "jazz":
			return 72
		case "bells":
			return 78
		default:
			return 64
		}
	}
}

func placePitchNear(note, center int) int {
	for note < center-6 {
		note += 12
	}
	for note > center+6 {
		note -= 12
	}
	for note < 24 {
		note += 12
	}
	for note > 108 {
		note -= 12
	}
	return note
}

func approachTo(rootPC, base int) int {
	target := placePitchNear(48+rootPC, base+7)
	if target >= base {
		return target - 1
	}
	return target + 1
}

func drumNoteFor(ctx authoredSectionContext, name string, slot int) int {
	lower := strings.ToLower(name)
	switch lower {
	case "kick":
		if ctx.has("drive", "pulse") && slot%authoredSlotsPerBar == 3 {
			return 35
		}
		return 36
	case "snare":
		if ctx.densityBias() > 0 && slot%authoredSlotsPerBar == 3 {
			return 37
		}
		return 38
	case "clap":
		return 39
	case "hat", "hihat":
		if ctx.shouldLift() && slot%authoredSlotsPerBar >= 6 {
			return 46
		}
		if slot%2 == 1 {
			return 44
		}
		return 42
	case "openhat":
		return 46
	case "ride":
		if slot%authoredSlotsPerBar == 7 {
			return 53
		}
		return 51
	case "crash":
		return 49
	case "rim":
		return 37
	default:
		return 42
	}
}

func keyRootPitchClass(key string) int {
	root, _, ok := parseRootToken(strings.Title(strings.TrimSpace(key)))
	if !ok {
		return 0
	}
	return root
}

func isAllRest(notes []int) bool {
	for _, note := range notes {
		if note >= 0 {
			return false
		}
	}
	return true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func clampInt32(v, low, high int32) int32 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func wrapPitchClass(pc int) int {
	for pc < 0 {
		pc += 12
	}
	for pc >= 12 {
		pc -= 12
	}
	return pc
}
