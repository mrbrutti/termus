package track

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func TestCompileBuildsTrackPlaylist(t *testing.T) {
	const src = `
title: Soft Tape / Rain Bus
style: lofi
listen_mode: album-side
seed: 42
roles:
  keys:
    family: electric_piano
    pattern: "x..x .x.."
  lead:
    family: reed_lead
    motif: "5 . 6 5 | 3 . 2 1"
sections:
  - id: intro
    title: curbside intro
    duration: 90s
    harmony: "Dm9 G13 | Cmaj9 A7"
    scene: "intro sparse"
    profile:
      density: sparse
      motion: gentle
  - id: return
    title: late platform
    duration: 120s
    harmony: "Fm9 Bb13 | Ebmaj9 C7"
    scene: "return lift"
    profile:
      density: busy
      swing: groove
    roles:
      lead:
        active: true
        motif: "9 . 7 5 | 3 . 2 1"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 99, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if got, want := compiled.Playlist.Mode, gen.PlaylistScore; got != want {
		t.Fatalf("playlist mode = %v, want %v", got, want)
	}
	if got, want := compiled.Playlist.ListenMode, gen.ListeningModeAlbumSide; got != want {
		t.Fatalf("listen mode = %q, want %q", got, want)
	}
	if got, want := len(compiled.Playlist.Tracks), 2; got != want {
		t.Fatalf("track count = %d, want %d", got, want)
	}
	if compiled.Playlist.Tracks[0].Title != "curbside intro" {
		t.Fatalf("track title = %q", compiled.Playlist.Tracks[0].Title)
	}
	if len(compiled.Plans) != 2 {
		t.Fatalf("plan count = %d, want 2", len(compiled.Plans))
	}
	for _, plan := range compiled.Plans {
		if len(plan.PhraseSpans) == 0 {
			t.Fatal("expected phrase spans in authored plan")
		}
	}
}

func TestCompileRejectsBadPattern(t *testing.T) {
	const src = `
title: Broken
style: lofi
roles:
  lead:
    family: reed_lead
    motif: "5 % 3"
sections:
  - title: bad
    duration: 30s
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := Compile(file, 1, gen.ListeningModeEndless); err == nil {
		t.Fatal("expected compile error for bad melody token")
	}
}

func TestCompileAppliesSectionEvents(t *testing.T) {
	const src = `
title: Eventful
style: jazz
seed: 17
roles:
  piano:
    family: acoustic_piano
    pattern: "x..x.x.. | .x..x..x"
  kick:
    family: drums
    pattern: "x...x... | x...x..."
  snare:
    family: drums
    pattern: "....x... | ....x..."
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
sections:
  - id: head
    duration: 16s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
    roles:
      lead:
        active: true
    events:
      - kind: fill
        bar: 2
        roles: [snare]
      - kind: drop
        bar: 3
        roles: [kick]
      - kind: pickup
        bar: 4
        roles: [lead]
        motif: "3 5 6 9"
      - kind: stab
        bar: 1
        roles: [piano]
        pattern: "x... ...."
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 17, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if len(compiled.Plans) != 1 {
		t.Fatalf("plan count = %d, want 1", len(compiled.Plans))
	}
	var plan gen.AuthoredTrackPlan
	for _, got := range compiled.Plans {
		plan = got
	}
	findTrack := func(name string) *gen.AuthoredRenderTrack {
		for i := range plan.Tracks {
			if plan.Tracks[i].Name == name {
				return &plan.Tracks[i]
			}
		}
		return nil
	}
	findPrefix := func(prefix string) *gen.AuthoredRenderTrack {
		for i := range plan.Tracks {
			if strings.HasPrefix(plan.Tracks[i].Name, prefix) {
				return &plan.Tracks[i]
			}
		}
		return nil
	}
	snare := findTrack("snare")
	if snare == nil {
		t.Fatal("expected snare track")
	}
	fillHasHit := false
	for i := 8; i < 16; i++ {
		if snare.Notes[i] >= 0 {
			fillHasHit = true
			break
		}
	}
	if !fillHasHit {
		t.Fatal("expected fill event to add a snare hit in bar 2")
	}
	kick := findTrack("kick")
	if kick == nil {
		t.Fatal("expected kick track")
	}
	for i := 16; i < 24; i++ {
		if kick.Notes[i] != -1 {
			t.Fatalf("expected dropped kick at slot %d, got %d", i, kick.Notes[i])
		}
	}
	lead := findTrack("lead")
	if lead == nil {
		t.Fatal("expected lead track")
	}
	pickupHasNote := false
	for i := 28; i < 32; i++ {
		if lead.Notes[i] >= 0 {
			pickupHasNote = true
			break
		}
	}
	if !pickupHasNote {
		t.Fatal("expected pickup event to add lead notes near the section close")
	}
	piano := findPrefix("piano-")
	if piano == nil {
		t.Fatal("expected piano voice track")
	}
	if got := piano.Notes[1]; got != -1 {
		t.Fatalf("expected stabbed piano slot 1 to be muted, got %d", got)
	}
}

func TestCompileBuildsPhraseBlocks(t *testing.T) {
	const src = `
title: Phrase Blocks
style: lofi
seed: 21
roles:
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
  keys:
    family: electric_piano
    pattern: "x..x .x.."
sections:
  - id: long
    duration: 32s
    harmony: "Dm9 G13 | Cmaj9 A7 | Bbmaj9 A7 | Dm9 G13"
    scene: "head glide"
    variation: "introduce-hook"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 21, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, got := range compiled.Plans {
		plan = got
	}
	if len(plan.PhraseSpans) < 2 {
		t.Fatalf("expected multiple phrase spans, got %d", len(plan.PhraseSpans))
	}
	if got, want := plan.PhraseSpans[0].Label, "statement"; got != want {
		t.Fatalf("first phrase label = %q, want %q", got, want)
	}
	if got, want := plan.PhraseSpans[len(plan.PhraseSpans)-1].Label, "release"; got != want {
		t.Fatalf("last phrase label = %q, want %q", got, want)
	}
}

func TestParseAuthoredChordPreservesColor(t *testing.T) {
	tests := []struct {
		token    string
		kind     string
		wantBass int
		hasBass  bool
		degrees  map[int]int
	}{
		{
			token:    "Dmaj9",
			kind:     "maj",
			wantBass: 2,
			degrees: map[int]int{
				3: 4,
				5: 7,
				7: 11,
				9: 14,
			},
		},
		{
			token:    "G13",
			kind:     "dom",
			wantBass: 7,
			degrees: map[int]int{
				3:  4,
				7:  10,
				9:  14,
				13: 21,
			},
		},
		{
			token:    "Em7b5",
			kind:     "half-dim",
			wantBass: 4,
			degrees: map[int]int{
				3: 3,
				5: 6,
				7: 10,
			},
		},
		{
			token:    "Asus4",
			kind:     "sus",
			wantBass: 9,
			degrees: map[int]int{
				3:  5,
				5:  7,
				7:  10,
				11: 17,
			},
		},
		{
			token:    "C7b9",
			kind:     "dom",
			wantBass: 0,
			degrees: map[int]int{
				3: 4,
				7: 10,
				9: 13,
			},
		},
		{
			token:    "A/C#",
			kind:     "dom",
			wantBass: 1,
			hasBass:  true,
			degrees: map[int]int{
				3: 4,
				5: 7,
				7: 10,
			},
		},
	}
	for _, tt := range tests {
		chord, ok := parseAuthoredChord(tt.token)
		if !ok {
			t.Fatalf("parseAuthoredChord(%q) failed", tt.token)
		}
		if chord.Kind != tt.kind {
			t.Fatalf("%q kind = %q, want %q", tt.token, chord.Kind, tt.kind)
		}
		if chord.BassPC != tt.wantBass {
			t.Fatalf("%q bass = %d, want %d", tt.token, chord.BassPC, tt.wantBass)
		}
		if chord.HasBass != tt.hasBass {
			t.Fatalf("%q hasBass = %v, want %v", tt.token, chord.HasBass, tt.hasBass)
		}
		for degree, interval := range tt.degrees {
			if got := chord.Degrees[degree]; got != interval {
				t.Fatalf("%q degree %d = %d, want %d", tt.token, degree, got, interval)
			}
		}
	}
}

func TestCompileBassLineHonorsSlashBass(t *testing.T) {
	chord, ok := parseAuthoredChord("Dmaj9/F#")
	if !ok {
		t.Fatal("failed to parse slash chord")
	}
	ctx := authoredSectionContext{
		style:   "lofi",
		profile: gen.DefaultControlProfile(),
	}
	role := Role{
		Family:   "bass",
		Register: "low",
		Pattern:  "x... ....",
	}
	notes := compileBassLine(ctx, "bass", role, []authoredHarmonyBar{{chords: []authoredChord{chord}}}, 1, []gen.AuthoredPhraseSpan{{
		StartBar: 1,
		EndBar:   1,
		Label:    "statement",
	}})
	if len(notes) == 0 || notes[0] < 0 {
		t.Fatalf("expected bass note, got %v", notes)
	}
	if got := ((notes[0] % 12) + 12) % 12; got != chord.BassPC {
		t.Fatalf("bass pitch class = %d, want %d", got, chord.BassPC)
	}
}

func TestRolePhraseModeOwnership(t *testing.T) {
	ctx := authoredSectionContext{
		style:   "lofi",
		profile: gen.DefaultControlProfile(),
	}
	statement := gen.AuthoredPhraseSpan{StartBar: 1, EndBar: 4, Label: "statement"}
	answer := gen.AuthoredPhraseSpan{StartBar: 5, EndBar: 8, Label: "answer"}
	release := gen.AuthoredPhraseSpan{StartBar: 9, EndBar: 12, Label: "release"}
	cadence := gen.AuthoredPhraseSpan{StartBar: 13, EndBar: 16, Label: "cadence"}

	if got := rolePhraseMode(ctx, "melody", "lead", Role{Family: "reed_lead", Prominence: "lead"}, statement, 0); got != "foreground" {
		t.Fatalf("melody statement mode = %q", got)
	}
	if got := rolePhraseMode(ctx, "melody", "lead", Role{Family: "reed_lead", Prominence: "lead"}, release, 1); got != "tail" {
		t.Fatalf("melody release mode = %q", got)
	}
	if got := rolePhraseMode(ctx, "bass", "bass", Role{Family: "bass", Prominence: "anchor"}, cadence, 2); got != "cadence" {
		t.Fatalf("bass cadence mode = %q", got)
	}
	if got := rolePhraseMode(ctx, "comp", "keys", Role{Family: "electric_piano", Prominence: "support"}, answer, 1); got != "answer" {
		t.Fatalf("comp answer mode = %q", got)
	}
	if got := rolePhraseMode(ctx, "pad", "texture", Role{Family: "bells", Prominence: "air"}, answer, 1); got != "echo" {
		t.Fatalf("texture answer mode = %q", got)
	}
	if got := rolePhraseMode(ctx, "drum", "snare", Role{Family: "drums", Prominence: "support"}, cadence, 2); got != "fill" {
		t.Fatalf("snare cadence mode = %q", got)
	}
}

func TestCompileAppliesPhraseOwnership(t *testing.T) {
	const src = `
title: Phrase Ownership
style: lofi
seed: 55
roles:
  keys:
    family: electric_piano
    pattern: "x..x .x.."
  bass:
    family: bass
    pattern: "x.x.x.x. | x.x.x.x."
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
sections:
  - id: loop
    duration: 24s
    harmony: "Dm9 G13 | Cmaj9 A7 | Bbmaj9 A7 | Dm9 G13"
    scene: "room steady"
    variation: "statement"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 55, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, got := range compiled.Plans {
		plan = got
	}
	if len(plan.PhraseSpans) != 2 {
		t.Fatalf("expected 2 phrase spans, got %d", len(plan.PhraseSpans))
	}
	countNotes := func(trackName string, start, end int) int {
		for _, track := range plan.Tracks {
			if track.Name != trackName && !strings.HasPrefix(track.Name, trackName+"-") {
				continue
			}
			count := 0
			for i := start; i < end && i < len(track.Notes); i++ {
				if track.Notes[i] >= 0 {
					count++
				}
			}
			return count
		}
		return 0
	}
	statementEnd := plan.PhraseSpans[0].EndBar * authoredSlotsPerBar
	releaseStart := (plan.PhraseSpans[1].StartBar - 1) * authoredSlotsPerBar
	releaseEnd := plan.PhraseSpans[1].EndBar * authoredSlotsPerBar
	leadStatement := countNotes("lead", 0, statementEnd)
	leadRelease := countNotes("lead", releaseStart, releaseEnd)
	if leadRelease >= leadStatement {
		t.Fatalf("expected lead release to thin out, got statement=%d release=%d", leadStatement, leadRelease)
	}
	keysStatement := countNotes("keys", 0, statementEnd)
	keysRelease := countNotes("keys", releaseStart, releaseEnd)
	if keysRelease >= keysStatement {
		t.Fatalf("expected comp release to thin out, got statement=%d release=%d", keysStatement, keysRelease)
	}
}

func TestResolveSectionsSupportsDeriveAndTransforms(t *testing.T) {
	const src = `
title: Derived Head
style: jazz
key: Dmajor
roles:
  lead:
    family: reed_lead
    register: mid
    motif: "5 . 6 7 | 3 . 2 1"
  keys:
    family: acoustic_piano
    register: mid
    pattern: "x..x .x.."
sections:
  - id: a
    title: head
    duration: 24s
    harmony: "Dmaj9 A/C# | Bm9 Gmaj9"
    scene: "head"
    variation: "statement"
  - id: a-prime
    derive: a
    title: head answer
    duration: 24s
    transforms: [sequence, lift-register, cadence-rewrite]
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sections, err := resolveSections(file)
	if err != nil {
		t.Fatalf("resolveSections: %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("resolved sections = %d, want 2", len(sections))
	}
	derived := sections[1]
	if !strings.Contains(derived.Variation, "sequence-up") || !strings.Contains(derived.Variation, "cadence") {
		t.Fatalf("derived variation = %q", derived.Variation)
	}
	if !strings.Contains(derived.Harmony, "Dmaj9") {
		t.Fatalf("derived harmony = %q", derived.Harmony)
	}

	compiled, err := Compile(file, 88, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if len(compiled.Playlist.Tracks) != 2 {
		t.Fatalf("compiled tracks = %d, want 2", len(compiled.Playlist.Tracks))
	}
	var secondPlan gen.AuthoredTrackPlan
	found := false
	for key, plan := range compiled.Plans {
		if strings.Contains(key, ":1097") {
			secondPlan = plan
			found = true
		}
	}
	if !found {
		for _, plan := range compiled.Plans {
			if plan.Section == "head answer" {
				secondPlan = plan
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected derived section plan")
	}
	var leadTrack *gen.AuthoredRenderTrack
	for i := range secondPlan.Tracks {
		if secondPlan.Tracks[i].Name == "lead" {
			leadTrack = &secondPlan.Tracks[i]
			break
		}
	}
	if leadTrack == nil {
		t.Fatal("expected derived lead track")
	}
	if leadTrack.Register != "mid-high" {
		t.Fatalf("derived lead register = %q", leadTrack.Register)
	}
	if got := secondPlan.PhraseSpans[len(secondPlan.PhraseSpans)-1].Label; got != "cadence" {
		t.Fatalf("derived last phrase = %q, want cadence", got)
	}
}

func TestCompileSupportsArrangementBlock(t *testing.T) {
	const src = `
title: Arrangement Block
style: jazz
roles:
  keys:
    family: acoustic_piano
    pattern: "x..x.x.. | .x..x..x"
  bass:
    family: bass
    pattern: "x.x.x.x. | x.x.x.x."
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
sections:
  - id: head
    duration: 16s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
    arrangement:
      events:
        - kind: pedal
          bar: 1
          roles: [bass]
        - kind: double
          bar: 2
          roles: [lead]
        - kind: swell
          bar: 3
          roles: [keys]
        - kind: ending
          bar: 4
          roles: [lead, keys]
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 91, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, got := range compiled.Plans {
		plan = got
	}
	var bass *gen.AuthoredRenderTrack
	var lead *gen.AuthoredRenderTrack
	var leadDouble *gen.AuthoredRenderTrack
	doubleFound := false
	for i := range plan.Tracks {
		track := &plan.Tracks[i]
		if track.Name == "bass" {
			bass = track
		}
		if track.Name == "lead" {
			lead = track
		}
		if strings.HasPrefix(track.Name, "lead-double") {
			doubleFound = true
			leadDouble = track
		}
	}
	if bass == nil {
		t.Fatal("expected bass track")
	}
	if !doubleFound {
		t.Fatal("expected arrangement double track")
	}
	if lead == nil || leadDouble == nil {
		t.Fatalf("expected lead and doubled lead tracks, got lead=%v leadDouble=%v", lead != nil, leadDouble != nil)
	}
	if lead.Channel == leadDouble.Channel {
		t.Fatalf("expected doubled lead on a distinct channel, both were %d", lead.Channel)
	}
	held := bass.Notes[0]
	if held < 0 {
		t.Fatalf("expected pedal note, got %d", held)
	}
	for i := 0; i < authoredSlotsPerBar; i++ {
		if bass.Notes[i] != held {
			t.Fatalf("expected pedal hold across bar 1, slot %d = %d want %d", i, bass.Notes[i], held)
		}
	}
}

func TestCompileSupportsRolePhraseBlocks(t *testing.T) {
	const src = `
title: Phrase Blocks In Roles
style: lofi
roles:
  keys:
    family: electric_piano
    pattern: "x..x .x.."
    phrases:
      release:
        pattern: "x....... | ....x..."
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
    phrases:
      release:
        motif: "3 . 2 . 1 . . ."
sections:
  - id: score
    duration: 24s
    harmony: "Dm9 G13 | Cmaj9 A7 | Bbmaj9 A7 | Dm9 G13"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 73, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var plan gen.AuthoredTrackPlan
	for _, got := range compiled.Plans {
		plan = got
	}
	if len(plan.PhraseSpans) != 2 {
		t.Fatalf("expected 2 phrase spans, got %d", len(plan.PhraseSpans))
	}
	releaseStart := (plan.PhraseSpans[1].StartBar - 1) * authoredSlotsPerBar
	releaseEnd := plan.PhraseSpans[1].EndBar * authoredSlotsPerBar
	var lead *gen.AuthoredRenderTrack
	var keys *gen.AuthoredRenderTrack
	for i := range plan.Tracks {
		if plan.Tracks[i].Name == "lead" {
			lead = &plan.Tracks[i]
		}
		if strings.HasPrefix(plan.Tracks[i].Name, "keys-") {
			keys = &plan.Tracks[i]
		}
	}
	if lead == nil || keys == nil {
		t.Fatalf("expected lead and keys tracks, got lead=%v keys=%v", lead != nil, keys != nil)
	}
	if got := ((lead.Notes[releaseStart] % 12) + 12) % 12; got != 5 {
		t.Fatalf("expected release phrase to start on scale degree 3 over Dm9, got pitch class %d", got)
	}
	active := 0
	for i := releaseStart; i < releaseEnd; i++ {
		if keys.Notes[i] >= 0 {
			active++
		}
	}
	if active > 2*authoredSlotsPerBar {
		t.Fatalf("expected release comp phrase to stay sparse, got %d active notes", active)
	}
}

func TestCompileSupportsOrchestrationDirectives(t *testing.T) {
	const src = `
title: Orchestration Directives
style: jazz
roles:
  lead:
    family: reed_lead
    register: mid-high
    prominence: lead
    motif: "5 . 6 7 | 3 . 2 1"
  comp:
    family: acoustic_piano
    register: mid
    pattern: "x..x.x.. | .x..x..x"
sections:
  - id: a
    duration: 24s
    harmony: "Dm7 G7 | Cmaj7 A7 | Dm7 G7 | Cmaj7 Cmaj7"
  - id: b
    duration: 24s
    harmony: "Fmaj7 E7 | Dm7 G7 | Em7 A7 | Dm7 G7"
    orchestration:
      roles:
        lead:
          family: brass
          register: high
          articulation: bright
        comp:
          family: organ
          prominence: support
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 99, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var derived gen.AuthoredTrackPlan
	for _, plan := range compiled.Plans {
		if plan.Section == "b" {
			derived = plan
			break
		}
	}
	var lead *gen.AuthoredRenderTrack
	var comp *gen.AuthoredRenderTrack
	for i := range derived.Tracks {
		if derived.Tracks[i].Name == "lead" {
			lead = &derived.Tracks[i]
		}
		if strings.HasPrefix(derived.Tracks[i].Name, "comp-") {
			comp = &derived.Tracks[i]
		}
	}
	if lead == nil || comp == nil {
		t.Fatalf("expected lead and comp tracks, got lead=%v comp=%v", lead != nil, comp != nil)
	}
	if lead.Family != "brass" || lead.Register != "high" {
		t.Fatalf("lead orchestration = family %q register %q", lead.Family, lead.Register)
	}
	if comp.Family != "organ" {
		t.Fatalf("comp family = %q", comp.Family)
	}
	if lead.Channel == comp.Channel {
		t.Fatalf("expected substituted brass lead and organ comp to stay on separate channels, both were %d", lead.Channel)
	}
}

func TestCompileVariationBudgetWarnings(t *testing.T) {
	const src = `
title: Budget Warnings
style: lofi
variation_budget:
  max_harmony_repeat: 1
  max_scene_repeat: 1
  max_motif_repeat: 1
  require_return_transform: true
roles:
  lead:
    family: reed_lead
    motif: "5 . 6 7 | 3 . 2 1"
sections:
  - id: a
    duration: 16s
    harmony: "Dm9 G13 | Cmaj9 A7"
    scene: "same-room"
  - id: b
    duration: 16s
    harmony: "Dm9 G13 | Cmaj9 A7"
    scene: "same-room"
  - id: c
    derive: a
    duration: 16s
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 11, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if len(compiled.Warnings) == 0 {
		t.Fatal("expected variation budget warnings")
	}
	text := make([]string, 0, len(compiled.Warnings))
	for _, warning := range compiled.Warnings {
		text = append(text, warning.Path+" "+warning.Message)
	}
	joined := strings.Join(text, "\n")
	for _, want := range []string{
		"variation_budget.max_harmony_repeat",
		"variation_budget.max_scene_repeat",
		"variation_budget.max_motif_repeat",
		"sections[2].transforms",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected warning containing %q, got:\n%s", want, joined)
		}
	}
}

func TestCompileLinterFlagsWeakContrastAndBrightOverload(t *testing.T) {
	const src = `
title: Linter Pressure
style: bells
roles:
  bells:
    family: bells
    tone: [glass, bright]
    articulation: bloom
    motif: "5 . 6 7 | 3 . 2 1"
  celesta:
    family: mallet
    tone: [sparkle, bright]
    articulation: echo
    pattern: "x....... | ....x..."
  glock:
    family: bells
    tone: [glass, luminous]
    articulation: echo
    pattern: "x....... | ....x..."
  box:
    family: music_box
    tone: [sparkle, bright]
    articulation: echo
    pattern: "x....... | ....x..."
sections:
  - id: a
    duration: 16s
    harmony: "Am7 Gmaj7 | Dm7 E7"
    scene: "same-room"
    variation: "steady"
  - id: b
    duration: 16s
    harmony: "Am7 Gmaj7 | Dm7 E7"
    scene: "same-room"
    variation: "steady"
`
	file, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	compiled, err := Compile(file, 18, gen.ListeningModeEndless)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	joined := make([]string, 0, len(compiled.Warnings))
	for _, warning := range compiled.Warnings {
		joined = append(joined, warning.Path+" "+warning.Message)
	}
	text := strings.Join(joined, "\n")
	for _, want := range []string{
		"sections[0].roles",
		"track has no clear cadence or ending shape",
		"section is too similar to its neighbor",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected warning containing %q, got:\n%s", want, text)
		}
	}
}

func TestBundledTracksParseAndCompile(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "tracks", "*", "*.tm"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(paths) < 10 {
		t.Fatalf("expected at least 10 bundled tracks, got %d", len(paths))
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile %s: %v", path, err)
		}
		file, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse %s: %v", path, err)
		}
		if _, err := Compile(file, 7, gen.ListeningModeEndless); err != nil {
			t.Fatalf("Compile %s: %v", path, err)
		}
		eventCount := 0
		for _, section := range file.Sections {
			eventCount += len(section.Events)
		}
		if eventCount == 0 {
			t.Fatalf("expected curated arrangement events in %s", path)
		}
	}
}

func TestResolveAcceptsDirectPath(t *testing.T) {
	entries, err := Discover(filepath.Join("..", "..", "tracks"))
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected bundled track entries")
	}
	entry, ok := Resolve(entries, entries[0].Path)
	if !ok {
		t.Fatal("Resolve should accept direct path")
	}
	if entry.Path != entries[0].Path {
		t.Fatalf("resolved path = %q, want %q", entry.Path, entries[0].Path)
	}
}
