# Handoff: Termus TUI Redesign — "the interface narrates the music"

## Overview
Redesign of the `termus` terminal music instrument's TUI (Go, Bubble Tea + Lip Gloss, `internal/tui/`). The current UI is deliberately minimal but "says nothing"; the redesign keeps the same keys, the same two-layer model (play view + control center), and the same theme system, but makes every surface communicate what the instrument is doing: station identity, live musical narration, position in the piece's form, and visible value scales.

## About the Design Files
The two `.dc.html` files in this bundle are **design references created in HTML** — they show intended look and content, they are not production code. The task is to **recreate these designs inside the existing Go TUI** using the codebase's established patterns (Lip Gloss styles, the `ColorTheme` struct, braille-grid rendering in `braille.go`/`visuals.go`, panel composition in `model.go`).

- `Termus Current UI.dc.html` — faithful recreation of the CURRENT TUI, built from `internal/tui/` source. Use it as the baseline/diff reference.
- `Termus Redesign.dc.html` — the TARGET. Five screens, rendered at 118×32 terminal cells, indigo theme.

## Fidelity
**High-fidelity** in terminal terms: exact strings, glyphs, cell layouts, color roles, and alignment are specified and should be matched. Hex values below are the **indigo theme's** concrete values — implement everything through the active `ColorTheme` (BarFg / BarHi / gradient / faint) so all six themes keep working.

## Color roles (indigo theme values)
| Role | Indigo value | Source |
| - | - | - |
| BarHi (highlight, bold headers, active row/section, current form section) | `#5bfaff` | `themes.go` |
| BarFg (primary chrome text, values, completed form sections) | `#a0a0ff` | `themes.go` |
| Default text | terminal default (`#d8d8d8` in mocks) | — |
| Faint default (hints, meta, footers) | Faint(true) (`#6a6a72` in mocks) | — |
| Faint BarFg (dimmed structure: dividers, future sections, subtitles) | (`#5a5a8c` in mocks) | Faint on BarFg |
| Inactive pips / idle bar dots | (`#4a4a6e` in mocks) | Faint |
| Scope gradient | center `#5bfaff` → edge `#5b4bff`, per braille row | `verticalGradient` in `themes.go` |

Glyphs reused from `tracks.go`: ◌ ambient, ▤ drone, ✶ bells, ☾ lullaby, ◇ classical, ∿ phase, ◒ lofi, ♬ jazz; `[SF2]` / `[AI]` engine badges.

---

## Screen 1 — Play view
**Purpose:** default live screen. Adds identity + narration + form position around the existing scope visual.

Layout (top to bottom, full width):
1. **Station header row** (replaces `topBar`):
   - Left: `◌ NIGHT DRIFT` — glyph + station label, UPPERCASE, **bold BarHi**; then 3 spaces and `ambient · C major · slow drift` in BarFg (algo · key · tempo/character).
   - Right: `seed 71001 · A 70992 · B 71001 · keep 2` faint. `● REC` in `#ff5b5b` appends when recording (unchanged behavior).
2. **Narration row** (replaces `playbackBar`'s left side):
   - Left, faint: `movement II · episode 3 · section drift · bar 129 · Dm9 → Gm7` — built from existing `gen.DebugStatus` (bar, section, chord) + long-form movement/episode state. This is the debug bar's data promoted to always-on chrome, in listening order: movement · episode · section · bar · chord-motion.
   - Right: existing level meter unchanged: `lvl ` BarFg + filled `─` BarHi + idle `─` faint + ` ok`/`clip`.
3. **Scope** — unchanged braille visual system (all 5 visuals, all themes).
4. **Form rail row** (new, one line above footer):
   - Left: section chain, e.g. `intro ─── ● drift ─── bloom ─── recede ─── coda`. Completed sections BarFg; current section prefixed `● `, **bold BarHi**; future sections faint BarFg; ` ─── ` connectors faint. Source: `TrackNavEntry.Structure` for authored tracks, or the long-form planner's episode/section list for procedural stations.
   - Right, faint: `2:13 to next section · endless` (time-to-next + listening mode).
5. **Footer row**: left faint `[space] play   [m] control   [t] tracks   [?] help`; right faint `[z] zen`.

**Zen mode** (`z`, existing reducedChrome): scope + a single bottom row `ambient` (BarFg) / `?` (faint). Narration and form rail hidden.

**Degradation:** on `useCompactLayout`, drop the narration row first, then the form rail; never drop the station header.

## Screen 2 — Control center (opens with `m`)
**Purpose:** same 8 sections (now/look/music/seeds/library/export/audio/debug), redesigned as a **full-screen pane** (no floating bordered box) with value scales.

Layout (padding 1 row / 2 cols):
1. Header row: `CONTROL CENTER` bold BarHi left; summary right faint: `Night Drift · ambient · seed 71001 · playing`.
2. Blank row.
3. Two columns, 4-col gap:
   - **Left rail, 14 cols:** section names lowercase; active section `▌ music` bold BarHi (▌ marker instead of ›); inactive `  now` etc. faint.
   - **Right pane (rest of width):** section title row `MUSIC` bold BarHi + faint annotation `nine macros · every change rebuilds the world in place`; blank; then one row per item:
     - `› ` cursor (default color) + title in BarHi, left.
     - Right-aligned cluster: **pip scale** `●●●○○` (filled pips BarHi, empty `○` faint) + 2 spaces + current value word in BarFg + the full word ladder faint, e.g. `air · lean · steady · lush · full`.
     - The nine music macros and ladders are exactly those in `control_center.go` `musicControlItems()` (density, brightness, motion, reverb, swing, drone depth, tempo, phrase length, seed morph).
     - Non-macro sections keep title/value/hint rows in the same geometry (pips only where a 5-step scale exists).
   - Below items, faint explainer line: `changes take effect at the next phrase boundary — nothing hard-cuts`.
4. Footer: left faint `[tab] section   [↑↓] row   [←→] adjust   [enter] apply   [m] close`; right faint `3 of 8` (active section index).

## Screen 3 — Track library (opens with `t`)
**Purpose:** browse authored `.tm` tracks; redesigned detail pane shows the piece's shape.

1. Header: `TRACK LIBRARY` bold BarHi left; right faint `53 authored tracks · one performer`.
2. Style filter row: `all 53   ◌ ambient 8   • blues 9   • chill 10   ♬ jazz 9   ▌◒ lofi 9   • rock 8` — active filter gets `▌` prefix + bold BarHi; others faint. (Counts come from the live catalog.)
3. Two panes (list ~46 cols, detail ~65 cols, 3-col gap; both clip at pane width like `trimToWidth`):
   - **List:** one entry per track: `  ◒ Title` padded, engine badge `[SF2]`/`[AI]` right-ish faint; meta line below faint (`substyle · NN sec · NN bpm`). Selected row: `▌ ` prefix, bold BarHi title, meta in dimmed BarHi (`#4bb8c4` in mocks).
   - **Detail:** title + badge bold BarHi; meta line BarFg `lofi · rainy-cafe · D minor · 84 bpm · hour-stream · 6m`; one-line description faint; blank; `FORM` faint; then per section: `label` (current/loaded section bold BarHi, others BarFg) + duration-proportional bar of `▰` (colored as label) padded with faint `▰`-width spaces + ` M:SS` + harmony/events faint. Then `ensemble …` and `textures rain −36 dB · vinyl −44 dB` faint, tag row `#lofi  #rhodes  #rain  #brushes` in faint BarFg, and `● currently loaded` BarHi when applicable. Data: `TrackNavEntry` (Structure, Ensemble, tags) — textures line needs `.tm` `textures:` exposed on the entry.
4. Footer faint: `[t] close   [←→] style   [↑↓] browse   [enter] play`.

## Screen 4 — Splash + loading (startup)
**Purpose:** merge `splashPanel` and `startupLoadingView` into one screen with identity and a station choice.

Centered, no chrome bars:
1. **TERMUS block wordmark**, 5 rows of `█` block letters (see the mock / generate per letterform), colored with the theme's vertical gradient (center BarHi-ish, edges toward gradient edge).
2. Tagline faint: `a terminal music instrument`.
3. Two blank rows.
4. **Station dial** (two lines): current station `◌ NIGHT DRIFT` bold BarHi; remaining stations lowercase faint with glyphs (`▤ deep field   ✶ glass chapel   ☾ sleep walk   ◇ chamber loop` / `∿ slow signal   ◒ soft tape   ♬ dusty swing`). `←→` moves the selection before playback starts.
5. Existing animated braille loading bar (`renderStartupBrailleBar`, 64×3 cells) — unchanged rendering, active portion gradient, remainder faint.
6. Progress row: `42%` BarHi + ` · loading SoundFont preset · general (32 MB)` faint (percent + current detail).
7. Blank; footer faint: `[←→] choose a station · [enter] begin · [t] authored tracks`.

## Screen 5 — Help overlay (`?`)
**Purpose:** replace the 6-line minimal help with a full grouped key reference — re-advertise the power-user keys that still work (`c C n p d z [ ] a b tab k x e`).

Centered bordered panel (rounded border BarFg, ~96 cols, padding 1 row / 3 cols):
- Title `TERMUS HELP` bold BarHi; blank.
- **Two columns** (~42 cols each, 6-col gap) of groups. Group headers BarHi; key column BarFg (~11 cols), action column default text:
  - Col 1: PLAYBACK (`space` play/pause, `↑ ↓ + −` volume, `n / p` next/previous algorithm, `r` record to ./exports) · VIEW (`c / C` theme/visual, `z` zen — scope only, `d` debug narration bar) · OPEN (`m` control center, `t` track library, `e` export drawer).
  - Col 2: SEEDS (`[ ]` browse seeds, `a / b` store slot A/B, `tab` compare A/B, `k / x` keep/reject take) · INSIDE PANELS (`↑ ↓` browse rows, `← →` adjust value, `enter` apply/open, `tab` next section) · GLOBAL (`?` this help, `q` quit).
- Footer line faint: `every key still works everywhere — the footer just stopped shouting about it`.

---

## Interactions & Behavior
- **No keybinding changes.** All existing keys, sections, and flows are preserved; this is a presentation redesign.
- Narration line and form rail update on the existing tick; form rail current-section marker moves at section boundaries (data already flows for the debug bar and SP17 section labels).
- Volume overlay, debug bar, REC indicator, A/B/keep chrome, playlist position labels: keep existing behavior, restyled into the new rows where they lived before.
- Visual/theme transitions (`startVisualTransition`, 340ms blend) unchanged.

## State Management
No new persistent state. New derived values only:
- station display name/glyph from `gen.AlgoSpec` (labels already exist),
- narration string from `gen.DebugStatus` + long-form movement/episode state,
- form rail segments from `TrackNavEntry.Structure` or the procedural planner,
- splash station-dial selection index (pre-play only; feeds the existing algo switch).

## Design Tokens (terminal)
- Grid: character cells; mock rendered at 118×32. Minimum stays ≥40×10 with `useCompactLayout` degradation.
- Weights: bold only for BarHi identity/active elements.
- Scale glyphs: pips `●`/`○`; form-map bars `▰`; markers `▌` (active rail/list row), `›` (row cursor), `●` (current form section / loaded track).
- All colors via `ColorTheme`; never hardcode indigo hexes.

## Assets
None. All glyphs are Unicode already used by the codebase (braille block U+2800–U+28FF, box drawing, `▰ ● ○ ▌`); wordmark is generated block letters.

## Files
- `Termus Redesign.dc.html` — target design (5 screens; includes station/seed/zen tweak variants).
- `Termus Current UI.dc.html` — baseline recreation of today's TUI.
- `PROMPT.md` — paste-ready prompt for Claude Code.

Key implementation sites in the repo: `internal/tui/model.go` (`View`, `topBar`, `playbackBar`, `bottomBar`, `splashPanel`, `startupLoadingView`, `helpPanel`), `internal/tui/control_center.go` (`controlsPanel`, `renderControlItem`), `internal/tui/tracks.go` (`trackPanel`, panes), `internal/tui/themes.go` (no changes needed).
