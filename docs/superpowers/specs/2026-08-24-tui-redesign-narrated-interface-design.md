# TUI Redesign — "the interface narrates the music"

**Date:** 2026-08-24
**Source:** `docs/design_handoff_tui_redesign/` (README.md is the authoritative visual spec; this doc records the implementation decisions and data plumbing on top of it).

## Goal

Implement the five-screen TUI redesign from the design handoff inside `internal/tui/`. Same keys, same two-layer model (play view + control center), same theme system — but every surface communicates what the instrument is doing: station identity, live musical narration, position in the piece's form, visible value scales.

This is a presentation-layer redesign. **No keybinding changes. No behavior changes**, with one sanctioned addition: the splash station dial (non-blocking, see below).

## Decisions made (with user approval)

1. **Narration data — wire the existing state.** `gen.DebugStatus` gains optional fields `Movement string`, `Episode int`, `NextChord string`. Algorithms that already hold a `FormPlan` / `LongHorizonState` populate them in their existing `DebugStatus()` methods; others leave them zero-valued and the narration line omits those segments. `gen.FormatDebugStatus` output is unchanged (debug bar stays stable).
2. **Splash dial is non-blocking.** Auto-start and the 5-second auto-dismiss stay exactly as today. While the splash is visible, `←`/`→` switch the station via the existing `switchAlgo` path *without dismissing the splash*; `enter` or any other key dismisses it as before.

## Visual spec

`docs/design_handoff_tui_redesign/README.md` is the fidelity reference: exact strings, glyphs, cell layouts, color roles, and alignment as specified there, per screen:

1. **Play view** — station header row (glyph + UPPERCASE station, bold BarHi; algo · key · tempo/character in BarFg; seed/A/B/keep + `● REC` right), narration row (`movement II · episode 3 · section drift · bar 129 · Dm9 → Gm7` faint left; existing level meter right), unchanged scope, form rail row (section chain with `● ` bold-BarHi current marker, `2:13 to next section · endless` right), footer `[space] play   [m] control   [t] tracks   [?] help` / `[z] zen`.
2. **Control center** — full-screen pane (no floating bordered box), `CONTROL CENTER` header + faint summary, 14-col left rail with `▌ ` active-section marker, right pane with `› ` row cursor, right-aligned `●●●○○` pip scales + current value word + full faint word ladder for the nine music macros, phrase-boundary explainer line, footer with `3 of 8` section index.
3. **Track library** — `TRACK LIBRARY` header + `NN authored tracks · one performer`, style filter row with `▌`-prefixed active filter and live counts, list pane (~46 cols) with engine badges and meta lines, detail pane (~65 cols) with meta line, description, `FORM` map (per-section duration-proportional `▰` bars + `M:SS` + harmony/events), ensemble, textures, `#tag` row, `● currently loaded`.
4. **Splash + loading** — merged `splashPanel`/`startupLoadingView`: 5-row block `█` TERMUS wordmark colored by the theme's vertical gradient (letterforms from the mock), faint tagline, station dial (current station bold BarHi, others faint with glyphs), existing animated braille loading bar, `42% · loading SoundFont preset · …` progress row, footer `[←→] choose a station · [enter] begin · [t] authored tracks`.
5. **Help overlay** — centered rounded-border panel ~96 cols, `TERMUS HELP`, two columns of grouped key references (PLAYBACK / VIEW / OPEN / SEEDS / INSIDE PANELS / GLOBAL), faint footer line.

**Color roles:** everything through the active `ColorTheme` (`BarFg` / `BarHi` / `Faint` / gradient). Never hardcode the indigo hexes; all six themes must keep working. Bold only for BarHi identity/active elements.

**Degradation:** minimum stays ≥40×10. On `useCompactLayout`, drop the narration row first, then the form rail; never drop the station header. Zen mode (`z`): scope + single bottom row (algo BarFg / `?` faint), narration and form rail hidden.

## Data plumbing (the only non-TUI changes)

- `gen.DebugStatus`: add `Movement string`, `Episode int`, `NextChord string` (optional; zero values omitted from narration). Populate in algorithms that already track this state (chill, jazz, sf2_markov, and any other `FormPlan` holders). No new engine logic.
- Procedural form rail: an optional provider interface (e.g. `FormChain() (sections []string, currentIdx int)`) surfaced through the same snapshot path, for algorithms whose `FormPlan` episode section chain exists. Authored tracks use `TrackNavEntry.Structure` / the active playlist track's sections instead. If no source is available, the form rail row collapses (row not rendered).
- `track.EntrySection`: add `Duration time.Duration`, filled in `discover.go` from the parsed file.
- `track.Entry`: add `Textures []string`, pre-formatted (e.g. `rain −36 dB`), from `File.Textures` (`TextureSpec.Name` + `LevelDB`).
- `tui.TrackNavSection` / `tui.TrackNavEntry`: mirror the two new fields; copy in `cmd/termus/main.go`.
- Splash dial selection feeds the existing algo-switch path; no new persistent state.

## Untouched

All keybindings; seeds/library/export/audio/debug control-center section behaviors; the inspector, saved-seeds library, and export panels (not part of the five redesigned screens — current look stays); visual/theme transitions (340 ms blend); volume overlay; REC indicator; A/B/keep chrome; playlist position labels (restyled into the new rows where they lived, behavior unchanged).

## Testing

- Extend the substring-assertion tests in `internal/tui/model_test.go` (and `tracks_test.go`, control-center coverage) per screen: station header content, narration segments and their graceful omission, form rail markers, pip scales + word ladders, form-map bars/durations, wordmark + station dial, help groups.
- New tests for plumbing: `DebugStatus` field population, `EntrySection.Duration` / `Entry.Textures` discovery, splash `←`/`→` switching station without dismissing, compact-layout degradation order.
- A test pass after each screen (per the handoff PROMPT).

## Implementation sites

`internal/tui/model.go` (`View`, `topBar`, `playbackBar`, `bottomBar`, `splashPanel`, `startupLoadingView`, `helpPanel`), `internal/tui/control_center.go` (`controlsPanel`, `renderControlItem`), `internal/tui/tracks.go` (`trackPanel`, panes), `internal/gen/debug_status.go` + the `FormPlan`-holding algorithms, `internal/track/model.go` + `discover.go`, `cmd/termus/main.go` (entry copy). `internal/tui/themes.go` unchanged.
