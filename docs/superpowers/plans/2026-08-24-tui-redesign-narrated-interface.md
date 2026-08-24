# TUI Redesign — Narrated Interface — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the five-screen TUI redesign from `docs/design_handoff_tui_redesign/README.md` — station identity, live musical narration, form rail, pip value scales, form-map track library, block-wordmark splash with station dial, and a full grouped help reference.

**Architecture:** Pure presentation redesign of `internal/tui/` plus two small data-plumbing changes: (1) `gen.DebugStatus` gains optional narration fields populated from the algorithms' existing `EpisodePlan`/progression state; (2) `track.EntrySection`/`track.Entry` gain per-section durations and texture labels, mirrored onto `tui.TrackNavEntry`. All colors flow through the active `ColorTheme` (never hardcode indigo hexes). No keybinding changes; the only behavior addition is the non-blocking splash station dial (`←`/`→` no longer dismiss the splash — they switch stations via the existing `switchAlgo` path).

**Tech Stack:** Go, Bubble Tea, Lip Gloss. Tests are Go `testing` substring assertions (existing style in `internal/tui/model_test.go`).

**Spec:** `docs/superpowers/specs/2026-08-24-tui-redesign-narrated-interface-design.md` + `docs/design_handoff_tui_redesign/README.md` (authoritative strings/layout).

**Verify commands:** `go build ./...` and `go test ./internal/... ./cmd/...` (run from repo root).

---

### Task 1: Narration fields on `gen.DebugStatus`

**Files:**
- Modify: `internal/gen/debug_status.go` (DebugStatus struct)
- Modify: `internal/gen/form.go` (add `EpisodePlan.FormStatus`)
- Modify: `internal/gen/chill.go:1434`, `internal/gen/jazz.go:1294`, `internal/gen/sf2_markov.go:576`, `internal/gen/ambient.go:466` (DebugStatus methods)
- Test: `internal/gen/debug_status_narration_test.go` (new)

- [ ] **Step 1: Write the failing tests**

Create `internal/gen/debug_status_narration_test.go`:

```go
package gen

import (
	"math/rand"
	"testing"
)

func TestEpisodePlanFormStatus(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	plan := NewEpisodePlan(rng, 1000, "jazz")
	movement, episode, chain, idx := plan.FormStatus(0)
	if movement != string(MovementEstablish) {
		t.Fatalf("movement = %q, want %q", movement, MovementEstablish)
	}
	if episode != 1 {
		t.Fatalf("episode = %d, want 1", episode)
	}
	if len(chain) == 0 || chain[0] != string(FormIntro) {
		t.Fatalf("chain = %v, want first section %q", chain, FormIntro)
	}
	if idx != 0 {
		t.Fatalf("idx = %d, want 0", idx)
	}
	// Move into the second section of episode 1: intro is 4 bars in the
	// jazz profile, so bar 5 (samples = 4*barSamples) is inside section A.
	_, _, _, idx2 := plan.FormStatus(4 * 1000)
	if idx2 != 1 {
		t.Fatalf("idx after intro = %d, want 1", idx2)
	}
}

func TestEpisodePlanFormStatusNilSafe(t *testing.T) {
	var plan *EpisodePlan
	movement, episode, chain, idx := plan.FormStatus(0)
	if movement != "" || episode != 0 || chain != nil || idx != 0 {
		t.Fatalf("nil plan should return zero values, got %q %d %v %d", movement, episode, chain, idx)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/gen/ -run TestEpisodePlanFormStatus -v`
Expected: FAIL — `plan.FormStatus undefined`

- [ ] **Step 3: Add the fields and the FormStatus method**

In `internal/gen/debug_status.go`, replace the `DebugStatus` struct with:

```go
// DebugStatus is a lightweight, UI-safe snapshot of the algorithm's current
// musical state.
type DebugStatus struct {
	Chord   string
	Section string
	Bar     int
	Preset  string

	// Narration fields (TUI redesign). Optional: zero values mean "not
	// exposed by this algorithm" and the TUI omits those narration parts.
	Movement  string   // long-form movement name, e.g. "develop"
	Episode   int      // 1-based episode number
	NextChord string   // label of the chord after the current one
	FormChain []string // section-kind chain of the current episode
	FormIndex int      // index of the current section within FormChain
}
```

In `internal/gen/form.go`, after `EpisodeIndexAt` (line ~310), add:

```go
// FormStatus reports the narration view of the plan at a sample position:
// the movement name, 1-based episode number, the episode's section-kind
// chain, and the index of the current section within that chain.
func (p *EpisodePlan) FormStatus(samples int64) (movement string, episode int, chain []string, idx int) {
	if p == nil || p.barSamples <= 0 {
		return "", 0, nil, 0
	}
	bar := sampleBarIndex(samples, p.barSamples)
	ep, epIdx := p.locateEpisode(bar)
	chain = make([]string, 0, len(ep.Sections))
	relative := bar - ep.StartBar
	idx = maxInt(0, len(ep.Sections)-1)
	acc := 0
	for i, section := range ep.Sections {
		chain = append(chain, string(section.Kind))
		if relative >= acc && relative < acc+section.Bars {
			idx = i
		}
		acc += section.Bars
	}
	return string(ep.Movement), epIdx + 1, chain, idx
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/gen/ -run TestEpisodePlanFormStatus -v`
Expected: PASS (both tests)

- [ ] **Step 5: Populate the fields in the four algorithms**

`internal/gen/chill.go` — replace `func (a *Chill) DebugStatus()`:

```go
func (a *Chill) DebugStatus() DebugStatus {
	bar := 0
	chord := ""
	next := ""
	if len(a.progression) > 0 {
		bar = a.currentBar()
		chord = a.progression[bar].label
		next = a.progression[(bar+1)%len(a.progression)].label
	}
	movement, episode, chain, idx := a.form.FormStatus(a.samplesElapsed)
	return DebugStatus{
		Chord:     chord,
		Section:   string(a.section.Kind),
		Bar:       a.form.BarAt(a.samplesElapsed),
		Movement:  movement,
		Episode:   episode,
		NextChord: next,
		FormChain: chain,
		FormIndex: idx,
	}
}
```

`internal/gen/jazz.go` — replace `func (a *Jazz) DebugStatus()` with the identical body (it has the same `progression`/`form`/`section`/`samplesElapsed` fields; keep the existing `bar = a.currentBar()` call).

`internal/gen/sf2_markov.go` — replace `func (a *SF2Markov) DebugStatus()`:

```go
func (a *SF2Markov) DebugStatus() DebugStatus {
	chord := ""
	next := ""
	if len(a.progression) > 0 {
		bar := sampleBarIndex(a.samplesElapsed, a.barSamples) % len(a.progression)
		chord = a.progression[bar].label
		next = a.progression[(bar+1)%len(a.progression)].label
	}
	movement, episode, chain, idx := a.form.FormStatus(a.samplesElapsed)
	return DebugStatus{
		Chord:     chord,
		Section:   string(a.section.Kind),
		Bar:       a.form.BarAt(a.samplesElapsed),
		Movement:  movement,
		Episode:   episode,
		NextChord: next,
		FormChain: chain,
		FormIndex: idx,
	}
}
```

`internal/gen/ambient.go` — replace `func (a *Ambient) DebugStatus()` (no EpisodePlan here; only NextChord is added):

```go
func (a *Ambient) DebugStatus() DebugStatus {
	chord := ""
	next := ""
	if len(a.chords) > 0 {
		chord = a.chords[a.currentChordIdx].label
		next = a.chords[(a.currentChordIdx+1)%len(a.chords)].label
	}
	return DebugStatus{
		Chord:     chord,
		Section:   string(a.section.Kind),
		Bar:       a.currentChordIdx + 1,
		NextChord: next,
	}
}
```

- [ ] **Step 6: Build and run the full gen test suite**

Run: `go build ./... && go test ./internal/gen/`
Expected: PASS. (`FormatDebugStatus` is untouched, so debug-bar output is unchanged.)

- [ ] **Step 7: Commit**

```bash
git add internal/gen/debug_status.go internal/gen/form.go internal/gen/chill.go internal/gen/jazz.go internal/gen/sf2_markov.go internal/gen/ambient.go internal/gen/debug_status_narration_test.go
git commit -m "feat(gen): expose movement/episode/next-chord narration state in DebugStatus"
```

---

### Task 2: Section durations + textures on `track.Entry`

**Files:**
- Modify: `internal/track/model.go:639-645` (EntrySection), `internal/track/model.go:614-637` (Entry)
- Modify: `internal/track/discover.go` (`buildEntrySummary`, both `Entry{...}` literals, new helpers)
- Test: `internal/track/discover_entry_test.go` (new)

- [ ] **Step 1: Write the failing test**

Create `internal/track/discover_entry_test.go`. It parses a real authored track from the repo's `tracks/` tree via `Discover` on a temp copy — instead, keep it hermetic with an inline `.tm` file:

```go
package track

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const entryTestTM = `title: Rainy Test
style: lofi
substyle: rainy-cafe
tempo: 84 bpm
key: D minor
textures:
  - name: rain
    level_db: -36
  - name: vinyl
    level_db: -44
sections:
  - id: intro
    duration: 30s
    harmony: "Dm9 | Gm7"
  - id: body
    duration: 90s
    harmony: "Dm9 | Bbmaj7"
`

func TestDiscoverEntryDurationsAndTextures(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "lofi")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "rainy-test.tm"), []byte(entryTestTM), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	entry := entries[0]
	if len(entry.Structure) != 2 {
		t.Fatalf("structure = %d sections, want 2", len(entry.Structure))
	}
	if entry.Structure[0].Duration != 30*time.Second {
		t.Fatalf("intro duration = %v, want 30s", entry.Structure[0].Duration)
	}
	if entry.Structure[1].Duration != 90*time.Second {
		t.Fatalf("body duration = %v, want 90s", entry.Structure[1].Duration)
	}
	if len(entry.Textures) != 2 || entry.Textures[0] != "rain -36 dB" || entry.Textures[1] != "vinyl -44 dB" {
		t.Fatalf("textures = %v, want [rain -36 dB, vinyl -44 dB]", entry.Textures)
	}
}
```

Note: if the minimal `.tm` above fails `ParseFile` validation (e.g. sections require roles), copy the smallest passing shape from an existing test in `internal/track/track_test.go` and keep the `textures:` block and two `duration:` values — the assertions are what matter.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/track/ -run TestDiscoverEntryDurationsAndTextures -v`
Expected: FAIL — `entry.Structure[0].Duration undefined` (compile error)

- [ ] **Step 3: Add the fields and fill them**

In `internal/track/model.go`, add to `EntrySection`:

```go
type EntrySection struct {
	ID        string
	Label     string
	Harmony   string
	RoleNames []string
	Events    []string
	// Duration is the section's authored length, parsed from the .tm
	// duration (or bars) field. Zero when unknown. Drives the
	// duration-proportional form-map bars in the TUI track library.
	Duration time.Duration
}
```

Add `"time"` to the imports of `model.go` if not present.

Add to `Entry` (after `Structure []EntrySection`):

```go
	// Textures are pre-formatted texture labels ("rain -36 dB") from the
	// file's textures: block, for the TUI track-library detail pane.
	Textures []string
```

In `internal/track/discover.go`, inside `buildEntrySummary`'s section loop, extend the `EntrySection{...}` literal (note: `resolveSections` has already converted `bars:` into a `Duration` string by this point — see `sections.go:28`):

```go
		dur, _ := time.ParseDuration(strings.TrimSpace(section.Duration))
		if dur < 0 {
			dur = 0
		}
		structure = append(structure, EntrySection{
			ID:        section.ID,
			Label:     label,
			Harmony:   section.Harmony,
			RoleNames: roleNames,
			Events:    append([]string(nil), events...),
			Duration:  dur,
		})
```

Add `"time"` to `discover.go` imports. Add the texture helper at the bottom of `discover.go`:

```go
// textureLabels renders the file's textures: block as display strings for
// the track-library detail pane, e.g. "rain -36 dB".
func textureLabels(textures []TextureSpec) []string {
	out := make([]string, 0, len(textures))
	for _, tx := range textures {
		name := strings.TrimSpace(tx.Name)
		if name == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s %.0f dB", name, tx.LevelDB))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
```

Add `Textures: textureLabels(file.Textures),` to **both** `Entry{...}` literals (in `Resolve` at discover.go:84 and in `loadEntry` at discover.go:131).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/track/ -run TestDiscoverEntryDurationsAndTextures -v`
Expected: PASS

- [ ] **Step 5: Run the full track suite and commit**

Run: `go test ./internal/track/`
Expected: PASS

```bash
git add internal/track/model.go internal/track/discover.go internal/track/discover_entry_test.go
git commit -m "feat(track): expose per-section durations and texture labels on Entry"
```

---

### Task 3: Mirror the new fields onto `tui.TrackNavEntry`

**Files:**
- Modify: `internal/tui/model.go:21-50` (TrackNavEntry / TrackNavSection)
- Modify: `cmd/termus/main.go:576-606` (entry copy loop)

- [ ] **Step 1: Add the fields**

In `internal/tui/model.go`, add to `TrackNavSection`:

```go
type TrackNavSection struct {
	ID        string
	Label     string
	Harmony   string
	RoleNames []string
	Events    []string
	Duration  time.Duration
}
```

Add to `TrackNavEntry` (after `Structure []TrackNavSection`):

```go
	// Textures are pre-formatted texture labels ("rain -36 dB") shown in
	// the track-library detail pane.
	Textures []string
```

- [ ] **Step 2: Copy them in cmd/termus**

In `cmd/termus/main.go`'s `trackNav` loop (line ~577), add `Duration: section.Duration,` to the `tui.TrackNavSection{...}` literal and `Textures: append([]string(nil), entry.Textures...),` to the `tui.TrackNavEntry{...}` literal.

- [ ] **Step 3: Build, test, commit**

Run: `go build ./... && go test ./internal/tui/ ./cmd/...`
Expected: PASS (fields are additive).

```bash
git add internal/tui/model.go cmd/termus/main.go
git commit -m "feat(tui): carry section durations and textures on TrackNavEntry"
```

---

### Task 4: Play view — station header, narration row, form rail, footer

**Files:**
- Create: `internal/tui/playview.go` (new chrome renderers)
- Modify: `internal/tui/model.go` (`View`, delete old `topBar`/`playbackBar` bodies, `bottomBar`)
- Test: `internal/tui/playview_test.go` (new) + update `internal/tui/model_test.go`

The old `topBar`, `playbackBar` are **replaced**; `bottomBar` is restyled. `debugBar`, `renderCompactMeter`, `renderVolumeLine` are unchanged.

- [ ] **Step 1: Write the failing tests**

Create `internal/tui/playview_test.go`:

```go
package tui

import (
	"strings"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func stationTestModel() Model {
	spec, _ := gen.Resolve("ambient")
	m := Model{width: 118, height: 32, keyName: "Cmin", seed: 71001, algo: spec.Label()}
	m.genres = []gen.AlgoSpec{spec}
	m.genreIdx = 0
	return m
}

func TestStationHeaderShowsIdentity(t *testing.T) {
	m := stationTestModel()
	out := topBar(m, 118, DefaultTheme(), false)
	if !strings.Contains(out, "NIGHT DRIFT") {
		t.Fatalf("station header missing uppercase station name: %q", out)
	}
	if !strings.Contains(out, "ambient") || !strings.Contains(out, "Cmin") {
		t.Fatalf("station header missing algo/key line: %q", out)
	}
	if !strings.Contains(out, "seed 71001") {
		t.Fatalf("station header missing seed: %q", out)
	}
	if !strings.Contains(out, "◌") {
		t.Fatalf("station header missing style glyph: %q", out)
	}
}

func TestNarrationRowPromotesDebugStatus(t *testing.T) {
	m := stationTestModel()
	m.debug = gen.DebugStatus{
		Movement: "develop", Episode: 3, Section: "drift",
		Bar: 129, Chord: "Dm9", NextChord: "Gm7",
	}
	out := playbackBar(m, 118, DefaultTheme(), make([]float64, 64), false)
	for _, want := range []string{"movement develop", "episode 3", "section drift", "bar 129", "Dm9 → Gm7", "lvl"} {
		if !strings.Contains(out, want) {
			t.Fatalf("narration row missing %q: %q", want, out)
		}
	}
}

func TestNarrationRowOmitsMissingParts(t *testing.T) {
	m := stationTestModel()
	m.debug = gen.DebugStatus{Section: "A", Bar: 4, Chord: "Cmaj7"}
	out := playbackBar(m, 118, DefaultTheme(), make([]float64, 64), false)
	if strings.Contains(out, "movement") || strings.Contains(out, "episode") {
		t.Fatalf("narration row should omit movement/episode when absent: %q", out)
	}
	if !strings.Contains(out, "Cmaj7") || strings.Contains(out, "→") {
		t.Fatalf("chord without next-chord should render bare: %q", out)
	}
}

func TestFormRailMarksCurrentSection(t *testing.T) {
	m := stationTestModel()
	m.debug = gen.DebugStatus{FormChain: []string{"intro", "A", "B", "outro"}, FormIndex: 1}
	out := formRailBar(m, 118, DefaultTheme())
	if !strings.Contains(out, "● A") {
		t.Fatalf("form rail missing current-section marker: %q", out)
	}
	if !strings.Contains(out, "intro") || !strings.Contains(out, "outro") {
		t.Fatalf("form rail missing chain sections: %q", out)
	}
	if !strings.Contains(out, "endless") {
		t.Fatalf("form rail missing listening mode: %q", out)
	}
}

func TestFormRailEmptyWithoutSource(t *testing.T) {
	m := stationTestModel()
	if out := formRailBar(m, 118, DefaultTheme()); out != "" {
		t.Fatalf("form rail should collapse without a source, got %q", out)
	}
}

func TestFooterAdvertisesCoreKeys(t *testing.T) {
	m := stationTestModel()
	out := bottomBar(m, 118, DefaultTheme(), false)
	for _, want := range []string{"[space] play", "[m] control", "[t] tracks", "[?] help", "[z] zen"} {
		if !strings.Contains(out, want) {
			t.Fatalf("footer missing %q: %q", want, out)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/ -run 'TestStationHeader|TestNarrationRow|TestFormRail|TestFooterAdvertises' -v`
Expected: FAIL (`formRailBar` undefined; old topBar output)

- [ ] **Step 3: Implement the new chrome in `internal/tui/playview.go`**

Create `internal/tui/playview.go`:

```go
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// stationGlyph maps an algorithm's canonical name to its style glyph.
// -synth variants share the base algorithm's glyph.
func stationGlyph(algoName string) string {
	return trackStyleGlyph(strings.TrimSuffix(strings.ToLower(strings.TrimSpace(algoName)), "-synth"))
}

// stationCharacter is the short tempo/character descriptor shown in the
// station header (presentation-only vocabulary, one entry per algorithm).
func stationCharacter(algoName string) string {
	switch strings.TrimSuffix(strings.ToLower(strings.TrimSpace(algoName)), "-synth") {
	case "ambient":
		return "slow drift"
	case "drone":
		return "held tones"
	case "bells":
		return "bright chimes"
	case "lullaby":
		return "gentle walk"
	case "classical":
		return "chamber lines"
	case "phase":
		return "shifting patterns"
	case "lofi":
		return "dusty groove"
	case "jazz":
		return "late swing"
	default:
		return ""
	}
}

// narrationParts assembles the always-on musical narration from DebugStatus,
// in listening order: movement · episode · section · bar · chord-motion.
// Missing fields are omitted so every algorithm degrades gracefully.
func narrationParts(m Model) []string {
	parts := make([]string, 0, 5)
	d := m.debug
	if d.Movement != "" {
		parts = append(parts, "movement "+d.Movement)
	}
	if d.Episode > 0 {
		parts = append(parts, fmt.Sprintf("episode %d", d.Episode))
	}
	section := d.Section
	if label := m.currentSectionLabel(); label != "" {
		section = label
	}
	if section != "" {
		parts = append(parts, "section "+section)
	}
	if d.Bar > 0 {
		parts = append(parts, fmt.Sprintf("bar %d", d.Bar))
	}
	switch {
	case d.Chord != "" && d.NextChord != "" && d.NextChord != d.Chord:
		parts = append(parts, d.Chord+" → "+d.NextChord)
	case d.Chord != "":
		parts = append(parts, d.Chord)
	}
	return parts
}

// formRailSegments picks the form-rail source: the authored track's section
// schedule when one is playing, else the procedural episode chain from
// DebugStatus. Empty when neither exists (the rail row collapses).
func formRailSegments(m Model) ([]string, int) {
	if track, ok := m.activeTrack(); ok && len(track.Sections) > 1 {
		labels := make([]string, 0, len(track.Sections))
		for i, s := range track.Sections {
			title := strings.TrimSpace(s.Title)
			if title == "" {
				title = fmt.Sprintf("section %d", i+1)
			}
			labels = append(labels, title)
		}
		return labels, clampInt(m.sectionIdx, 0, len(labels)-1)
	}
	if len(m.debug.FormChain) > 0 {
		return m.debug.FormChain, clampInt(m.debug.FormIndex, 0, len(m.debug.FormChain)-1)
	}
	return nil, 0
}

// formRailBar renders the one-line section chain with the current-section
// marker, plus time-to-next-section and the listening mode on the right.
// Returns "" when there is no form source.
func formRailBar(m Model, w int, theme ColorTheme) string {
	segments, current := formRailSegments(m)
	if len(segments) == 0 {
		return ""
	}
	mode := m.listeningMode
	if mode == "" {
		mode = "endless"
	}
	rightParts := make([]string, 0, 2)
	if !m.nextSectionAt.IsZero() {
		rightParts = append(rightParts, shortDuration(time.Until(m.nextSectionAt))+" to next section")
	}
	rightParts = append(rightParts, mode)
	right := lipgloss.NewStyle().Faint(true).Render(strings.Join(rightParts, " · "))

	connector := lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(" ─── ")
	plainWidth := 0
	rendered := make([]string, 0, len(segments))
	for i, label := range segments {
		display := label
		if i == current {
			display = "● " + label
		}
		plainWidth += lipgloss.Width(display)
		if i > 0 {
			plainWidth += 5 // " ─── "
		}
		switch {
		case i == current:
			rendered = append(rendered, lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(display))
		case i < current:
			rendered = append(rendered, lipgloss.NewStyle().Foreground(theme.BarFg).Render(display))
		default:
			rendered = append(rendered, lipgloss.NewStyle().Foreground(theme.BarFg).Faint(true).Render(display))
		}
	}
	left := strings.Join(rendered, connector)
	if plainWidth > w-lipgloss.Width(right)-1 {
		// Compact fallback: current section + position, never a mid-ANSI trim.
		left = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).
			Render(fmt.Sprintf("● %s · %d/%d", segments[current], current+1, len(segments)))
	}
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}
```

- [ ] **Step 4: Replace `topBar` in `internal/tui/model.go`**

Replace the whole `func topBar(...)` body with:

```go
func topBar(m Model, w int, theme ColorTheme, compact bool) string {
	spec, hasSpec := m.activeSpec()
	station := m.algo
	glyph := "•"
	metaParts := make([]string, 0, 3)
	if hasSpec {
		station = spec.Label()
		glyph = stationGlyph(spec.Name)
		metaParts = append(metaParts, spec.Name)
	}
	if m.keyName != "" {
		metaParts = append(metaParts, m.keyName)
	}
	if hasSpec {
		if character := stationCharacter(spec.Name); character != "" {
			metaParts = append(metaParts, character)
		}
	}
	identity := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).
		Render(glyph + " " + strings.ToUpper(station))
	left := identity
	if len(metaParts) > 0 && !compact {
		left += "   " + lipgloss.NewStyle().Foreground(theme.BarFg).Render(strings.Join(metaParts, " · "))
	}

	rightParts := []string{fmt.Sprintf("seed %d", m.seed)}
	if !compact {
		if m.seedA != nil {
			rightParts = append(rightParts, fmt.Sprintf("A %d", m.seedA.Seed))
		}
		if m.seedB != nil {
			rightParts = append(rightParts, fmt.Sprintf("B %d", m.seedB.Seed))
		}
	}
	if len(m.kept) > 0 {
		rightParts = append(rightParts, fmt.Sprintf("keep %d", len(m.kept)))
	}
	right := lipgloss.NewStyle().Faint(true).Render(strings.Join(rightParts, " · "))
	if m.recording {
		right += "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5b5b")).Render("● REC")
	}
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}
```

Delete the now-unused `seedSlotsLabel` method (its display moved inline) — run `go vet ./internal/tui/` to confirm nothing else references it; if `model_test.go` does, update that test in Step 7.

- [ ] **Step 5: Replace `playbackBar` in `internal/tui/model.go`**

```go
func playbackBar(m Model, w int, theme ColorTheme, samples []float64, compact bool) string {
	leftParts := narrationParts(m)
	if m.recording && !m.recordStartedAt.IsZero() {
		leftParts = append(leftParts, formatElapsed("rec", time.Since(m.recordStartedAt)))
	}
	if m.aceRenderActive {
		label := m.aceRenderDetail
		if label == "" {
			label = "generating next track"
		}
		leftParts = append(leftParts, "AI: "+label)
	}
	leftText := trimToWidth(strings.Join(leftParts, " · "), maxInt(0, w-22))
	meter, clipped := meterSummary(samples)
	meterWidth := 14
	if compact {
		meterWidth = 8
	}
	right := renderCompactMeter(theme, meter, clipped, meterWidth)
	left := lipgloss.NewStyle().Faint(true).Render(leftText)
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}
```

(The old live/track/next/fade timing drops from default chrome per the mock; time-to-next-section lives in the form rail, and the playlist-position label still appears in flash statuses and the control center.)

- [ ] **Step 6: Restyle `bottomBar` and wire the rail into `View`**

Replace `bottomBar`'s non-zen branch (keep the `reducedChrome` branch exactly as is):

```go
	leftText := lipgloss.NewStyle().Faint(true).
		Render("[space] play   [m] control   [t] tracks   [?] help")
	rightText := lipgloss.NewStyle().Faint(true).Render("[z] zen")
	status := m.currentStatus(time.Now())
	statusStyle := lipgloss.NewStyle().Foreground(theme.BarHi)
	if status == "" {
		statusStyle = lipgloss.NewStyle().Faint(true)
		status = " "
	}
	centerText := statusStyle.Render(trimToWidth(status, maxInt(0, w/3)))
	available := w - lipgloss.Width(leftText) - lipgloss.Width(rightText) - 2
	if available < 1 {
		available = 1
	}
	centerWidth := lipgloss.Width(centerText)
	if centerWidth > available {
		centerText = statusStyle.Render(trimToWidth(status, available))
		centerWidth = lipgloss.Width(centerText)
	}
	leftPad := (available - centerWidth) / 2
	rightPad := available - centerWidth - leftPad
	return leftText + spaces(leftPad+1) + centerText + spaces(rightPad+1) + rightText
```

In `View` (model.go:1395), replace the chrome-height computation and final joins:

```go
	compact := useCompactLayout(m.width, m.height)
	showNarration := !compact
	rail := ""
	if !m.reducedChrome && m.height >= 14 {
		rail = formRailBar(m, m.width, theme)
	}
	chromeH := 2 // station header + footer
	if m.reducedChrome {
		chromeH = 1
	} else {
		if showNarration {
			chromeH++
		}
		if rail != "" {
			chromeH++
		}
		if m.debugVisible {
			chromeH++
		}
	}
	if m.volumeOverlayVisible(now) {
		chromeH++
	}
	innerH := m.height - chromeH
	innerW := m.width
```

And the assembly at the bottom of `View` (replacing the current joins; overlays/panels selection stays as-is above this):

```go
	if m.reducedChrome {
		if volumeLine != "" {
			return lipgloss.JoinVertical(lipgloss.Left, body, volumeLine, bottom)
		}
		return lipgloss.JoinVertical(lipgloss.Left, body, bottom)
	}
	rows := []string{top}
	if showNarration {
		rows = append(rows, playback)
	}
	if volumeLine != "" {
		rows = append(rows, volumeLine)
	}
	if m.debugVisible {
		rows = append(rows, debugBar(m, innerW, theme))
	}
	rows = append(rows, body)
	if rail != "" {
		rows = append(rows, rail)
	}
	rows = append(rows, bottom)
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
```

(`playback := playbackBar(...)` can stay computed unconditionally; the rail sits one line above the footer per the mock.)

- [ ] **Step 7: Run the tui suite; update stale assertions**

Run: `go test ./internal/tui/ -v 2>&1 | grep -E '^(=== RUN|--- FAIL|FAIL|ok)' | head -40`

Update the tests that asserted the old chrome strings to the new spec strings — expected failures and their fixes:
- `TestTopBarShowsTitle` / `TestTopBarShowsStationAndAlgoNameWhenSpecAvailable`: assert `NIGHT DRIFT` / station+algo per new format instead of `termus ·`.
- `TestPlaybackBarShowsTimingAndMeter`: rename to narration semantics — assert `lvl` still present and (with a `debug` fixture) a narration part; drop `live`/`fade` assertions.
- `TestBottomBarLeavesRoomForStatus`: keep the status-centering assertion, update the left/right expectations to `[space] play` / `[z] zen`.
- Any test asserting `?  m` in the footer: update to `[?] help`.

Expected after updates: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/tui/playview.go internal/tui/playview_test.go internal/tui/model.go internal/tui/model_test.go
git commit -m "feat(tui): play view narrates the music — station header, narration row, form rail"
```

---

### Task 5: Control center — full-screen pane with pip scales and word ladders

**Files:**
- Modify: `internal/tui/control_center.go` (`controlItem`, `musicControlItems`, `controlsPanel`, `renderControlSection`, `renderControlItem`, `controlCenterSummary`)
- Modify: `internal/tui/model.go` `View` (controls render full-screen)
- Test: `internal/tui/control_center_test.go` (new) + update `internal/tui/model_test.go` control tests

- [ ] **Step 1: Write the failing tests**

Create `internal/tui/control_center_test.go`:

```go
package tui

import (
	"strings"
	"testing"

	"github.com/mrbrutti/termus/internal/gen"
)

func controlsTestModel() Model {
	spec, _ := gen.Resolve("ambient")
	profile := gen.DefaultControlProfile()
	m := Model{width: 118, height: 32, keyName: "Cmin", seed: 71001, algo: spec.Label()}
	m.genres = []gen.AlgoSpec{spec}
	m.musicProfile = &profile
	m.controlsVisible = true
	m.controlTab = controlTabMusic
	return m
}

func TestControlsPanelShowsPipScalesAndLadders(t *testing.T) {
	m := controlsTestModel()
	out := controlsPanel(m, 118, 32, DefaultTheme())
	if !strings.Contains(out, "CONTROL CENTER") || !strings.Contains(out, "MUSIC") {
		t.Fatalf("missing headers: %q", out)
	}
	if !strings.Contains(out, "●") || !strings.Contains(out, "○") {
		t.Fatalf("missing pip scale: %q", out)
	}
	if !strings.Contains(out, "air · lean · steady · lush · full") {
		t.Fatalf("missing density word ladder: %q", out)
	}
	if !strings.Contains(out, "nine macros") {
		t.Fatalf("missing section annotation: %q", out)
	}
	if !strings.Contains(out, "next phrase boundary") {
		t.Fatalf("missing explainer line: %q", out)
	}
	if !strings.Contains(out, "3 of 8") {
		t.Fatalf("missing section index (music is tab 3): %q", out)
	}
	if !strings.Contains(out, "▌ music") {
		t.Fatalf("missing active-section marker in rail: %q", out)
	}
}

func TestControlsPanelPipCountMatchesValue(t *testing.T) {
	theme := DefaultTheme()
	pips := renderPips(theme, 2, 5) // value "steady" = index 2 -> 3 filled
	if got := strings.Count(pips, "●"); got != 3 {
		t.Fatalf("filled pips = %d, want 3", got)
	}
	if got := strings.Count(pips, "○"); got != 2 {
		t.Fatalf("empty pips = %d, want 2", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/ -run TestControlsPanel -v`
Expected: FAIL (`renderPips` undefined; old panel content)

- [ ] **Step 3: Add Scale fields + renderPips + annotations**

In `control_center.go`, extend `controlItem`:

```go
type controlItem struct {
	Title    string
	Value    string
	Hint     string
	Disabled bool
	// Scale/ScaleValue drive the ●●●○○ pip cluster and the faint word
	// ladder. Scale is nil for rows without a 5-step macro.
	Scale      []string
	ScaleValue int
	Adjust     func(*Model, int)
	Activate   func(*Model) tea.Cmd
}
```

In `musicControlItems`, for each of the nine macros add `Scale` and `ScaleValue` using the same ladder already passed to `macroLabel`. Pattern (repeat for all nine; the ladders are already in the file):

```go
		{
			Title:      "density",
			Value:      macroLabel(profile.Density, []string{"air", "lean", "steady", "lush", "full"}),
			Scale:      []string{"air", "lean", "steady", "lush", "full"},
			ScaleValue: profile.Density,
			Hint:       "left/right rebuild",
			Adjust: func(m *Model, delta int) {
				m.updateMusicProfile("density", func(profile *gen.ControlProfile) {
					profile.Density += delta
				})
			},
		},
```

To stay DRY, hoist each ladder into a package-level `var` (e.g. `var densityLadder = []string{"air", "lean", "steady", "lush", "full"}`) and use it in both `Value` and `Scale`. Do the same for brightness, motion, reverb, swing, drone depth (`droneLadder`), tempo, phrase length (`phraseLadder`), and seed morph (`morphLadder` — `ScaleValue: m.morphMode`).

Add helpers at the bottom of `control_center.go`:

```go
// renderPips draws the 5-step value scale: value is the 0-based macro index,
// so index 2 fills three pips (●●●○○).
func renderPips(theme ColorTheme, value, steps int) string {
	filled := clampInt(value+1, 0, steps)
	return lipgloss.NewStyle().Foreground(theme.BarHi).Render(strings.Repeat("●", filled)) +
		lipgloss.NewStyle().Faint(true).Render(strings.Repeat("○", steps-filled))
}

// controlSectionAnnotation is the faint one-liner next to the section title.
func controlSectionAnnotation(tab controlTab) string {
	switch tab {
	case controlTabNow:
		return "playback and session at a glance"
	case controlTabLook:
		return "themes · visuals · chrome"
	case controlTabMusic:
		return "nine macros · every change rebuilds the world in place"
	case controlTabSeeds:
		return "browse, bookmark, and keep generative seeds"
	case controlTabLibrary:
		return "curation · history · sessions"
	case controlTabExport:
		return "render artifacts to ./exports"
	case controlTabAudio:
		return "backend health and recovery"
	default:
		return "live engine state"
	}
}

// controlSectionExplainer is the faint line below the item list.
func controlSectionExplainer(tab controlTab) string {
	if tab == controlTabMusic {
		return "changes take effect at the next phrase boundary — nothing hard-cuts"
	}
	return ""
}
```

- [ ] **Step 4: Rewrite `controlsPanel` as a full-screen pane**

Replace `controlsPanel`, `renderControlSection`, `renderControlItem`, and `controlCenterSummary`:

```go
func controlsPanel(m Model, w, h int, theme ColorTheme) string {
	innerW := maxInt(40, w-4)
	sidebarW := 14
	rightW := maxInt(24, innerW-sidebarW-4)
	sections := []controlTab{
		controlTabNow, controlTabLook, controlTabMusic, controlTabSeeds,
		controlTabLibrary, controlTabExport, controlTabAudio, controlTabDebug,
	}
	sidebarLines := make([]string, 0, len(sections))
	for _, section := range sections {
		sidebarLines = append(sidebarLines, renderControlSection(theme, m.controlTab == section, section.label()))
	}

	title := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(strings.ToUpper(m.controlTab.label()))
	annotation := lipgloss.NewStyle().Faint(true).Render(controlSectionAnnotation(m.controlTab))
	lines := []string{title + "  " + annotation, ""}
	items := m.controlItems()
	for i, item := range items {
		lines = append(lines, renderControlItem(theme, i == m.controlRow, item, rightW))
	}
	if explainer := controlSectionExplainer(m.controlTab); explainer != "" {
		lines = append(lines, "", lipgloss.NewStyle().Faint(true).Render(explainer))
	}
	if m.controlTab == controlTabDebug {
		if preview := renderControlTrackStructure(m, theme, rightW, h-len(lines)-8); preview != "" {
			lines = append(lines, "", preview)
		}
	}

	sidebar := lipgloss.NewStyle().Width(sidebarW).Render(strings.Join(sidebarLines, "\n"))
	content := lipgloss.NewStyle().Width(rightW).Render(strings.Join(lines, "\n"))
	main := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, spaces(4), content)

	header := controlHeaderRow(m, theme, innerW)
	footerLeft := lipgloss.NewStyle().Faint(true).
		Render("[tab] section   [↑↓] row   [←→] adjust   [enter] apply   [m] close")
	footerRight := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("%d of 8", int(m.controlTab)+1))
	footerPad := innerW - lipgloss.Width(footerLeft) - lipgloss.Width(footerRight)
	if footerPad < 1 {
		footerPad = 1
	}
	footer := footerLeft + spaces(footerPad) + footerRight

	bodyH := maxInt(1, h-2-3) // padding rows + header + blank + footer
	body := lipgloss.NewStyle().Height(bodyH).Render(main)
	return lipgloss.NewStyle().Width(w).Height(h).Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, header, "", body, footer))
}

func controlHeaderRow(m Model, theme ColorTheme, w int) string {
	left := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("CONTROL CENTER")
	right := lipgloss.NewStyle().Faint(true).Render(controlCenterSummary(m))
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}

func renderControlSection(theme ColorTheme, active bool, label string) string {
	if active {
		return lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("▌ " + label)
	}
	return lipgloss.NewStyle().Faint(true).Render("  " + label)
}

func renderControlItem(theme ColorTheme, active bool, item controlItem, w int) string {
	cursor := "  "
	if active {
		cursor = "› "
	}
	titleStyle := lipgloss.NewStyle().Foreground(theme.BarHi)
	valueStyle := lipgloss.NewStyle().Foreground(theme.BarFg)
	hintStyle := lipgloss.NewStyle().Faint(true)
	if item.Disabled {
		titleStyle = titleStyle.Faint(true)
		valueStyle = valueStyle.Faint(true)
	}
	left := cursor + titleStyle.Render(item.Title)
	var right string
	if len(item.Scale) > 0 {
		right = renderPips(theme, item.ScaleValue, len(item.Scale)) + "  " +
			valueStyle.Render(item.Value) + "  " +
			hintStyle.Render(strings.Join(item.Scale, " · "))
	} else {
		right = valueStyle.Render(item.Value)
		if item.Hint != "" {
			right += "  " + hintStyle.Render(item.Hint)
		}
	}
	if lipgloss.Width(left)+lipgloss.Width(right)+1 > w && len(item.Scale) > 0 {
		// Not enough room for the ladder — drop it, keep pips + value.
		right = renderPips(theme, item.ScaleValue, len(item.Scale)) + "  " + valueStyle.Render(item.Value)
	}
	pad := w - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 1 {
		pad = 1
	}
	return left + spaces(pad) + right
}

func controlCenterSummary(m Model) string {
	parts := []string{m.algo}
	if spec, ok := m.activeSpec(); ok {
		parts = []string{spec.Label(), spec.Name}
	}
	parts = append(parts, fmt.Sprintf("seed %d", m.seed))
	if m.paused {
		parts = append(parts, "paused")
	} else {
		parts = append(parts, "playing")
	}
	return strings.Join(parts, " · ")
}
```

Delete the now-unused `currentTabItems` wrapper if nothing else references it.

In `model.go` `View`, render the control center full-screen — move its check above the chrome assembly, next to the `startupLoading` early return:

```go
	if m.controlsVisible {
		return controlsPanel(m, m.width, m.height, theme)
	}
```

…and remove `controlsVisible` from the body-selection chain.

- [ ] **Step 5: Run tests, update stale ones**

Run: `go test ./internal/tui/ -run 'TestControls' -v`
Expected new tests PASS. Update the older control-center tests that asserted the bordered-overlay layout:
- `TestControlsPanelShowsTabbedOverlay`: assert `▌` marker and `CONTROL CENTER` instead of `›`-marked rail.
- `TestControlsPanelShowsAudioRecoveryActions` / `TestControlsPanelShowsTrackStructureInspector`: content assertions should still pass; fix only if the geometry trimmed their strings (widen the test width to 118 if needed).

Run: `go test ./internal/tui/`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/tui/control_center.go internal/tui/control_center_test.go internal/tui/model.go internal/tui/model_test.go
git commit -m "feat(tui): full-screen control center with pip scales and word ladders"
```

---

### Task 6: Track library — filter row, form map, textures, tags

**Files:**
- Modify: `internal/tui/tracks.go` (`trackPanel`, `renderTrackStyleBar`, `renderTrackListPane`, `renderTrackDetailPane`, new `renderTrackFormMap`)
- Modify: `internal/tui/model.go` `View` (track panel renders full-screen)
- Test: `internal/tui/tracks_test.go` (extend)

- [ ] **Step 1: Write the failing tests**

Append to `internal/tui/tracks_test.go`:

```go
func formMapTestModel() Model {
	m := Model{width: 118, height: 32}
	m.tracks = []TrackNavEntry{{
		ID: "lofi/rainy", Style: "lofi", Substyle: "rainy-cafe", Title: "Rainy Cafe",
		Key: "D minor", Tempo: "84", ListenMode: "hour-stream",
		Description: "warm rhodes over rain",
		Ensemble:    []string{"rhodes", "bass"},
		Tags:        []string{"lofi", "rhodes"},
		Textures:    []string{"rain -36 dB", "vinyl -44 dB"},
		Structure: []TrackNavSection{
			{ID: "intro", Label: "intro", Harmony: "Dm9", Duration: 30 * time.Second},
			{ID: "body", Label: "body", Harmony: "Gm7", Duration: 90 * time.Second},
		},
	}}
	m.trackVisible = true
	return m
}

func TestTrackDetailShowsFormMap(t *testing.T) {
	m := formMapTestModel()
	out := trackPanel(m, 118, 32, DefaultTheme())
	if !strings.Contains(out, "FORM") || !strings.Contains(out, "▰") {
		t.Fatalf("detail missing form map: %q", out)
	}
	if !strings.Contains(out, "0:30") || !strings.Contains(out, "1:30") {
		t.Fatalf("detail missing section durations: %q", out)
	}
	if !strings.Contains(out, "textures  rain -36 dB · vinyl -44 dB") {
		t.Fatalf("detail missing textures line: %q", out)
	}
	if !strings.Contains(out, "#lofi") || !strings.Contains(out, "#rhodes") {
		t.Fatalf("detail missing tag row: %q", out)
	}
	if !strings.Contains(out, "warm rhodes over rain") {
		t.Fatalf("detail missing description: %q", out)
	}
}

func TestTrackHeaderShowsCountAndFilters(t *testing.T) {
	m := formMapTestModel()
	out := trackPanel(m, 118, 32, DefaultTheme())
	if !strings.Contains(out, "1 authored track") {
		t.Fatalf("header missing count: %q", out)
	}
	if !strings.Contains(out, "▌") {
		t.Fatalf("style bar missing active marker: %q", out)
	}
	if !strings.Contains(out, "[t] close   [←→] style   [↑↓] browse   [enter] play") {
		t.Fatalf("footer changed: %q", out)
	}
}
```

Add `"time"` to the test file's imports if missing. Note `shortDuration` renders `00:30`; the assertion strings above must match the helper you use — this plan adds `formMapDuration` below which renders `M:SS` (`0:30`, `1:30`). Keep them consistent.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/ -run 'TestTrackDetail|TestTrackHeader' -v`
Expected: FAIL

- [ ] **Step 3: Rewrite the header/filter/detail rendering in `tracks.go`**

In `trackPanel`, replace the header block (title/subtitle/braille accent):

```go
	title := lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TRACK LIBRARY")
	countNoun := "authored tracks"
	if len(m.tracks) == 1 {
		countNoun = "authored track"
	}
	right := lipgloss.NewStyle().Faint(true).
		Render(fmt.Sprintf("%d %s · one performer", len(m.tracks), countNoun))
	headPad := maxInt(1, (w-4)-lipgloss.Width(title)-lipgloss.Width(right))
	header := title + spaces(headPad) + right
```

(and use `header` where the old three-line `header` was joined; the braille `accent` line is removed).

Replace `renderTrackStyleBar`'s active-style branch:

```go
		if strings.EqualFold(style, active) {
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("▌"+text))
			continue
		}
```

In `renderTrackListPane`, change the selected-row prefix from `"▸ "` to `"▌ "` and render the selected meta line in dimmed BarHi:

```go
		metaStyle := lipgloss.NewStyle().Faint(true)
		if idx == m.trackIdx {
			metaStyle = lipgloss.NewStyle().Foreground(blendColor(theme.BarHi, lipgloss.Color("#000000"), 0.27))
		}
		block := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(idx == m.trackIdx).Render(titleLine),
			metaStyle.Render(trimToWidth("  "+meta, maxInt(8, w-2))),
		)
```

Rewrite `renderTrackDetailPane` to the mock's structure (title+badge; meta line with computed total duration; description; blank; FORM map; ensemble; textures; tags; loaded marker):

```go
func renderTrackDetailPane(m Model, w, h int, theme ColorTheme) string {
	style := lipgloss.NewStyle().Width(w).Height(h)
	if len(m.tracks) == 0 || m.trackIdx < 0 || m.trackIdx >= len(m.tracks) {
		return style.Render("")
	}
	entry := m.tracks[m.trackIdx]
	title := firstNonEmpty(entry.Title, entry.ID)
	titleLine := trackStyleGlyph(entry.Style) + trackSubstyleGlyph(entry.Substyle) + " " + title
	if badge := renderEngineBadge(entry.Engine, theme, true); badge != "" {
		titleLine += " " + badge
	}
	lines := []string{lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render(titleLine)}

	meta := make([]string, 0, 6)
	for _, part := range []string{entry.Style, entry.Substyle, entry.Key} {
		if part != "" {
			meta = append(meta, part)
		}
	}
	if entry.Tempo != "" {
		meta = append(meta, entry.Tempo+" bpm")
	}
	if entry.ListenMode != "" {
		meta = append(meta, entry.ListenMode)
	}
	var total time.Duration
	for _, s := range entry.Structure {
		total += s.Duration
	}
	if total > 0 {
		meta = append(meta, formMapDuration(total))
	}
	if len(meta) > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.BarFg).Render(trimToWidth(strings.Join(meta, " · "), w)))
	}
	if desc := strings.TrimSpace(entry.Description); desc != "" {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(trimToWidth(desc, w)))
	}
	if len(entry.Structure) > 0 {
		lines = append(lines, "", lipgloss.NewStyle().Faint(true).Render("FORM"))
		budget := maxInt(3, h-len(lines)-6)
		lines = append(lines, renderTrackFormMap(m, entry, w, budget, theme)...)
	}
	if len(entry.Ensemble) > 0 {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(trimToWidth("ensemble  "+strings.Join(entry.Ensemble, " · "), w)))
	}
	if len(entry.Textures) > 0 {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(trimToWidth("textures  "+strings.Join(entry.Textures, " · "), w)))
	}
	if len(entry.Tags) > 0 {
		lines = append(lines, renderTrackTags(entry.Tags, theme, w))
	}
	if entry.ID == m.activeTrackID {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(theme.BarHi).Render("● currently loaded"))
	}
	return style.Render(strings.Join(lines, "\n"))
}

// renderTrackFormMap draws one row per section: label, a duration-
// proportional ▰ bar, the section length, and harmony/events meta.
func renderTrackFormMap(m Model, entry TrackNavEntry, w, budget int, theme ColorTheme) []string {
	var longest time.Duration
	labelW := 0
	for _, s := range entry.Structure {
		if s.Duration > longest {
			longest = s.Duration
		}
		if n := lipgloss.Width(firstNonEmpty(s.Label, s.ID)); n > labelW {
			labelW = n
		}
	}
	labelW = clampInt(labelW, 4, 14)
	barMax := clampInt(w-labelW-24, 6, 24)
	isActive := entry.ID == m.activeTrackID
	current, hasCurrent := currentTrackStructureSection(entry, m.debug.Section)
	lines := make([]string, 0, len(entry.Structure))
	for i, s := range entry.Structure {
		if i >= budget {
			lines = append(lines, lipgloss.NewStyle().Faint(true).Render("…"))
			break
		}
		label := firstNonEmpty(s.Label, s.ID)
		cells := 1
		if longest > 0 && s.Duration > 0 {
			cells = maxInt(1, int(float64(barMax)*float64(s.Duration)/float64(longest)))
		}
		labelStyle := lipgloss.NewStyle().Foreground(theme.BarFg)
		if isActive && hasCurrent && sectionsMatch(s, current) {
			labelStyle = lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true)
		}
		dur := ""
		if s.Duration > 0 {
			dur = " " + formMapDuration(s.Duration)
		}
		meta := ""
		if harmony := compactHarmony(s.Harmony); harmony != "" {
			meta = harmony
		} else if len(s.Events) > 0 {
			meta = strings.Join(s.Events, " · ")
		}
		line := labelStyle.Render(padRight(trimToWidth(label, labelW), labelW)) + "  " +
			labelStyle.Render(strings.Repeat("▰", cells)) + spaces(barMax-cells) +
			lipgloss.NewStyle().Faint(true).Render(dur)
		if meta != "" {
			line += "  " + lipgloss.NewStyle().Faint(true).Render(trimToWidth(meta, maxInt(0, w-lipgloss.Width(line)-2)))
		}
		lines = append(lines, line)
	}
	return lines
}

// formMapDuration renders M:SS without the zero-padded minute of
// shortDuration, matching the mock ("0:30", "6:00").
func formMapDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Round(time.Second).Seconds())
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

func padRight(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return s + spaces(w-n)
	}
	return s
}
```

Add `"time"` to `tracks.go` imports. Adjust the pane split in `trackPanel` toward the mock's list/detail ratio:

```go
	leftW := clampInt(int(float64(w)*0.40), 24, 46)
```

In `model.go` `View`, render the track panel full-screen like the control center:

```go
	if m.trackVisible {
		return trackPanel(m, m.width, m.height, theme)
	}
```

…and remove `trackVisible` from the body-selection chain.

- [ ] **Step 4: Run tests, update stale ones**

Run: `go test ./internal/tui/ -run 'TestTrack' -v`
Expected: new tests PASS. Update older `tracks_test.go` assertions that referenced `▸` or the removed `authored songs · one performer · [enter] play` subtitle.

Run: `go test ./internal/tui/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/tracks.go internal/tui/tracks_test.go internal/tui/model.go
git commit -m "feat(tui): track library form map with durations, textures, and tags"
```

---

### Task 7: Splash + loading — wordmark, station dial, merged screen

**Files:**
- Create: `internal/tui/splash.go`
- Modify: `internal/tui/model.go` (`View` splash/loading routing, `Update` splash key handling; delete old `splashPanel`, fold `startupLoadingView`)
- Test: `internal/tui/splash_test.go` (new) + update splash/loading tests in `model_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/tui/splash_test.go`:

```go
package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrbrutti/termus/internal/gen"
)

func splashTestModel() Model {
	ambient, _ := gen.Resolve("ambient")
	drone, _ := gen.Resolve("drone")
	m := Model{width: 118, height: 32, algo: ambient.Label(), cmd: &tuiCommanderStub{}}
	m.genres = []gen.AlgoSpec{ambient, drone}
	m.genreIdx = 0
	m.buildFn = func(spec gen.AlgoSpec, seed int64) gen.Algorithm { return &tuiAlgoStub{name: spec.Name} }
	m.splashVisible = true
	m.splashUntil = time.Now().Add(5 * time.Second)
	return m
}

func TestSplashScreenShowsWordmarkAndDial(t *testing.T) {
	m := splashTestModel()
	out := splashScreen(m, 118, 32, DefaultTheme(), time.Now())
	if !strings.Contains(out, "█") {
		t.Fatalf("splash missing block wordmark: %q", out)
	}
	if !strings.Contains(out, "a terminal music instrument") {
		t.Fatalf("splash missing tagline: %q", out)
	}
	if !strings.Contains(out, "NIGHT DRIFT") {
		t.Fatalf("splash missing selected station: %q", out)
	}
	if !strings.Contains(out, "deep field") {
		t.Fatalf("splash missing other stations: %q", out)
	}
	if !strings.Contains(out, "[←→] choose a station · [enter] begin · [t] authored tracks") {
		t.Fatalf("splash missing footer: %q", out)
	}
}

func TestSplashArrowKeysSwitchStationWithoutDismissing(t *testing.T) {
	m := splashTestModel()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated := next.(Model)
	if !updated.splashVisible {
		t.Fatal("right arrow should not dismiss the splash")
	}
	if updated.genreIdx != 1 {
		t.Fatalf("genreIdx = %d, want 1 after right arrow", updated.genreIdx)
	}
	next2, _ := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next2.(Model).splashVisible {
		t.Fatal("enter should dismiss the splash")
	}
}

func TestStartupLoadingShowsProgressOnSplash(t *testing.T) {
	m := splashTestModel()
	m.startupLoading = true
	m.startupTitle = "Loading termus"
	m.startupDetail = "loading SoundFont preset"
	m.startupPercent = 0.42
	out := splashScreen(m, 118, 32, DefaultTheme(), time.Now())
	if !strings.Contains(out, "42%") || !strings.Contains(out, "loading SoundFont preset") {
		t.Fatalf("loading splash missing progress: %q", out)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/ -run TestSplash -v`
Expected: FAIL (`splashScreen` undefined; arrows dismiss splash)

- [ ] **Step 3: Implement `internal/tui/splash.go`**

```go
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// wordmarkLetters are the 5-row block letterforms from the design mock.
var wordmarkLetters = map[rune][5]string{
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'E': {"████ ", "█    ", "███  ", "█    ", "████ "},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
	'M': {"█   █", "██ ██", "█ █ █", "█   █", "█   █"},
	'U': {"█   █", "█   █", "█   █", "█   █", " ███ "},
	'S': {" ████", "█    ", " ███ ", "    █", "████ "},
}

// renderWordmark draws text in block letters, colored per row with the
// theme's gradient (center rows brightest, edges toward the gradient edge).
func renderWordmark(text string, theme ColorTheme) string {
	rows := make([]string, 5)
	for r := 0; r < 5; r++ {
		var b strings.Builder
		first := true
		for _, ch := range text {
			letter, ok := wordmarkLetters[ch]
			if !ok {
				continue
			}
			if !first {
				b.WriteString("  ")
			}
			first = false
			b.WriteString(letter[r])
		}
		color := theme.BarHi
		if theme.ColorAt != nil {
			color = theme.ColorAt(0, r, 1, 5)
		}
		rows[r] = lipgloss.NewStyle().Foreground(color).Render(b.String())
	}
	return strings.Join(rows, "\n")
}

// stationDialLines renders the splash station chooser: the selected station
// bold BarHi, the remaining stations faint, wrapped to maxW.
func stationDialLines(m Model, theme ColorTheme, maxW int) []string {
	if len(m.genres) == 0 {
		return nil
	}
	cur := m.genres[clampInt(m.genreIdx, 0, len(m.genres)-1)]
	lines := []string{lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).
		Render(stationGlyph(cur.Name) + " " + strings.ToUpper(cur.Label()))}
	faint := lipgloss.NewStyle().Faint(true)
	row := ""
	for i, g := range m.genres {
		if i == clampInt(m.genreIdx, 0, len(m.genres)-1) {
			continue
		}
		s := stationGlyph(g.Name) + " " + strings.ToLower(g.Label())
		if row != "" && lipgloss.Width(row)+3+lipgloss.Width(s) > maxW {
			lines = append(lines, faint.Render(row))
			row = s
			continue
		}
		if row == "" {
			row = s
		} else {
			row += "   " + s
		}
	}
	if row != "" {
		lines = append(lines, faint.Render(row))
	}
	return lines
}

// splashScreen is the merged splash + startup-loading screen: wordmark,
// tagline, station dial, braille loading bar, and (while loading) the
// progress row and any composing-context block.
func splashScreen(m Model, w, h int, theme ColorTheme, now time.Time) string {
	barW := maxInt(26, minInt(w-10, 64))
	phase := float64(now.UnixNano()) / float64(time.Second)
	progress := 1.0
	if m.startupLoading {
		progress = clamp01(m.startupPercent)
	}
	parts := []string{
		renderWordmark("TERMUS", theme),
		lipgloss.NewStyle().Faint(true).Render("a terminal music instrument"),
		"",
		"",
	}
	parts = append(parts, stationDialLines(m, theme, barW)...)
	parts = append(parts, "", renderStartupBrailleBar(barW, 3, progress, phase, theme))
	if m.startupLoading {
		pct := lipgloss.NewStyle().Foreground(theme.BarHi).Render(fmt.Sprintf("%d%%", int(progress*100)))
		detailParts := make([]string, 0, 2)
		if m.startupTitle != "" {
			detailParts = append(detailParts, m.startupTitle)
		}
		if m.startupDetail != "" {
			detailParts = append(detailParts, m.startupDetail)
		}
		row := pct
		if len(detailParts) > 0 {
			row += lipgloss.NewStyle().Faint(true).Render(" · " + strings.Join(detailParts, " · "))
		}
		parts = append(parts, row)
		if ctx := composingContextBlock(m, barW, theme); ctx != "" {
			parts = append(parts, "", ctx)
		}
	}
	parts = append(parts, "",
		lipgloss.NewStyle().Faint(true).Render("[←→] choose a station · [enter] begin · [t] authored tracks"))
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}
```

- [ ] **Step 4: Route View/Update through the new screen**

In `model.go` `View`, replace the `startupLoading` early return and drop `splashVisible` from the body chain:

```go
	if m.startupLoading || m.splashVisible {
		return splashScreen(m, m.width, m.height, theme, now)
	}
```

Delete `splashPanel` and `startupLoadingView` (keep `renderStartupBrailleBar` and `composingContextBlock` — the new screen uses them; move them into `splash.go` if you prefer, updating nothing else).

In `Update`'s `tea.KeyMsg` case, replace the bare `if m.splashVisible { m.splashVisible = false }` with:

```go
		if m.splashVisible {
			// Station dial: ←/→ browse stations without dismissing the
			// splash (playback is already running — this rides the normal
			// switchAlgo path). Any other key dismisses as before.
			switch msg.String() {
			case "left":
				m.switchAlgo(-1)
				m.splashUntil = time.Now().Add(5 * time.Second)
				return m, nil
			case "right":
				m.switchAlgo(1)
				m.splashUntil = time.Now().Add(5 * time.Second)
				return m, nil
			}
			m.splashVisible = false
		}
```

(Re-arming `splashUntil` keeps the dial from vanishing mid-browse.)

- [ ] **Step 5: Run tests, update stale ones**

Run: `go test ./internal/tui/ -run 'TestSplash|TestStartupLoading' -v`
Expected: new tests PASS. Update in `model_test.go`:
- `TestSplashPanelShowsOnboarding`: now asserts wordmark/tagline/footer via `splashScreen` (delete or rewrite against the new content).
- `TestSplashPanelShowsStartupLoading` / `TestStartupLoadingViewShowsBrailleStyleProgress` / `TestStartupLoadingViewBypassesChrome`: point them at `splashScreen` / `View` output; the braille bar and percent assertions still apply.
- `TestStartupLoadingBlocksDismissal`: unchanged behavior; should still pass.

Run: `go test ./internal/tui/`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/tui/splash.go internal/tui/splash_test.go internal/tui/model.go internal/tui/model_test.go
git commit -m "feat(tui): merged splash/loading screen with TERMUS wordmark and station dial"
```

---

### Task 8: Help overlay — full grouped key reference

**Files:**
- Modify: `internal/tui/model.go` (`helpPanel`; delete `styleHelpLine` uses there if unused elsewhere — note `inspectorPanel`/`exportPanel`/`splash` may still use it, keep it then)
- Test: update `TestHelpPanelShowsCoreControls` in `internal/tui/model_test.go`

- [ ] **Step 1: Update the test first**

Replace `TestHelpPanelShowsCoreControls`'s assertions:

```go
func TestHelpPanelShowsCoreControls(t *testing.T) {
	m := Model{width: 118, height: 32}
	out := helpPanel(m, 118, 32, DefaultTheme())
	for _, want := range []string{
		"TERMUS HELP", "PLAYBACK", "VIEW", "OPEN", "SEEDS", "INSIDE PANELS", "GLOBAL",
		"n / p", "[ ]", "a / b", "k / x", "zen — scope only",
		"every key still works everywhere",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("help panel missing %q", want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestHelpPanelShowsCoreControls -v`
Expected: FAIL

- [ ] **Step 3: Rewrite `helpPanel`**

```go
type helpGroup struct {
	Title string
	Rows  [][2]string
}

func helpPanel(m Model, w, h int, theme ColorTheme) string {
	col1 := []helpGroup{
		{"PLAYBACK", [][2]string{
			{"space", "play / pause"},
			{"↑ ↓ + −", "volume"},
			{"n / p", "next / previous algorithm"},
			{"r", "record to ./exports"},
		}},
		{"VIEW", [][2]string{
			{"c / C", "theme / visual"},
			{"z", "zen — scope only"},
			{"d", "debug narration bar"},
		}},
		{"OPEN", [][2]string{
			{"m", "control center"},
			{"t", "track library"},
			{"e", "export drawer"},
		}},
	}
	col2 := []helpGroup{
		{"SEEDS", [][2]string{
			{"[ ]", "browse seeds"},
			{"a / b", "store slot A / B"},
			{"tab", "compare A/B"},
			{"k / x", "keep / reject take"},
		}},
		{"INSIDE PANELS", [][2]string{
			{"↑ ↓", "browse rows"},
			{"← →", "adjust value"},
			{"enter", "apply / open"},
			{"tab", "next section"},
		}},
		{"GLOBAL", [][2]string{
			{"?", "this help"},
			{"q", "quit"},
		}},
	}
	bodyW := maxInt(48, minInt(w-4, 96))
	colW := (bodyW - 6 - 6) / 2 // padding cols + gap
	left := renderHelpColumn(col1, colW, theme)
	right := renderHelpColumn(col2, colW, theme)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, left, spaces(6), right)
	footer := lipgloss.NewStyle().Faint(true).
		Render("every key still works everywhere — the footer just stopped shouting about it")
	panel := lipgloss.NewStyle().
		Width(bodyW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BarFg).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(theme.BarHi).Bold(true).Render("TERMUS HELP"),
			"",
			columns,
			"",
			footer,
		))
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, panel)
}

func renderHelpColumn(groups []helpGroup, w int, theme ColorTheme) string {
	const keyW = 11
	lines := make([]string, 0, 16)
	for gi, group := range groups {
		if gi > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.BarHi).Render(group.Title))
		for _, row := range group.Rows {
			key := lipgloss.NewStyle().Foreground(theme.BarFg).Render(padRight(row[0], keyW))
			lines = append(lines, key+trimToWidth(row[1], maxInt(0, w-keyW)))
		}
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(lines, "\n"))
}
```

Delete `filterHelpLines` (unused). Keep `styleHelpLine` — `inspectorPanel`/`exportPanel` still use it.

- [ ] **Step 4: Run test to verify it passes, then the package**

Run: `go test ./internal/tui/ -run TestHelpPanel -v && go test ./internal/tui/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/model_test.go
git commit -m "feat(tui): full grouped help reference re-advertising power-user keys"
```

---

### Task 9: Full verification + degradation checks

**Files:**
- Test: `internal/tui/playview_test.go` (add degradation tests)

- [ ] **Step 1: Add compact-degradation tests**

Append to `internal/tui/playview_test.go`:

```go
func TestCompactLayoutDropsNarrationKeepsHeader(t *testing.T) {
	m := stationTestModel()
	m.width, m.height = 60, 16 // compact per useCompactLayout
	m.debug = gen.DebugStatus{Section: "A", Bar: 3, Chord: "Cm"}
	out := m.View()
	if !strings.Contains(out, "NIGHT DRIFT") {
		t.Fatalf("compact view must keep the station header: %q", out)
	}
	if strings.Contains(out, "section A · bar 3") {
		t.Fatalf("compact view should drop the narration row: %q", out)
	}
}

func TestTinyLayoutDropsFormRail(t *testing.T) {
	m := stationTestModel()
	m.width, m.height = 50, 12
	m.debug = gen.DebugStatus{FormChain: []string{"intro", "A"}, FormIndex: 0}
	out := m.View()
	if strings.Contains(out, "─── ") {
		t.Fatalf("tiny view should drop the form rail: %q", out)
	}
}
```

Note: `m.View()` needs `m.ring` — `stationTestModel` builds a bare `Model{}`, so add `ring: scope.NewRing(1024)` to `stationTestModel` (import `github.com/mrbrutti/termus/internal/scope`) and `cmd: &tuiCommanderStub{}`; `View` calls `m.ring.Snapshot` and the tick path is not exercised.

- [ ] **Step 2: Run the full suite**

Run: `go build ./... && go test ./...`
Expected: PASS across `internal/...` and `cmd/...`. Fix any straggler assertions against removed strings (`termus ·`, `▸ `, old help lines) by aligning them with the new spec strings.

- [ ] **Step 3: Visual smoke test across themes**

Run: `go run ./cmd/termus --help` to confirm the binary builds, then a short interactive smoke if a terminal is available: check play view (header/narration/rail), `m`, `t`, `?`, `z`, `c` twice (theme cycling — confirm no hardcoded indigo), resize small.
For a non-interactive check: `go test ./internal/tui/ -run 'TestStation|TestNarration|TestFormRail|TestControls|TestTrack|TestSplash|TestHelp' -v`
Expected: PASS

- [ ] **Step 4: Final commit**

```bash
git add -A internal/ cmd/ docs/superpowers/plans/2026-08-24-tui-redesign-narrated-interface.md
git commit -m "feat(tui): complete narrated-interface redesign (SP31)"
```

---

## Notes for the implementer

- **Never hardcode theme colors.** Only `#ff5b5b` (REC) is a sanctioned literal (pre-existing). Everything else: `theme.BarFg`, `theme.BarHi`, `.Faint(true)`, `theme.ColorAt`, `blendColor`.
- **Never `trimToWidth` an already-styled string** — it slices runes and will cut ANSI sequences. Trim plain text, then style (existing code follows this; the form rail uses a compact fallback instead).
- **`useCompactLayout` (w<72 or h<18)** drops the narration row; the form rail additionally requires `h >= 14`. The station header never drops (only zen hides it).
- The README's exact footer/annotation strings are load-bearing for tests — copy them from the spec, don't paraphrase.
- `gen.FormatDebugStatus` must keep its exact current output (debug bar + inspector depend on it).
- If a step's stated line numbers have drifted, locate by function name — they are all unique in their files.
