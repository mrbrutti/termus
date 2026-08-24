# Prompt for Claude Code

Paste this into Claude Code from the root of the `termus` repository, with this handoff folder available (e.g. dropped in at `./design_handoff_tui_redesign/`):

---

Implement the TUI redesign described in `design_handoff_tui_redesign/README.md`.

Context: this repo is `termus`, a Go terminal music instrument built on Bubble Tea + Lip Gloss (`internal/tui/`). The handoff folder contains two HTML files — `Termus Current UI.dc.html` is a faithful recreation of the CURRENT TUI (your baseline, already matching this codebase), and `Termus Redesign.dc.html` is the TARGET. Both are design references only; the implementation is pure Go changes to `internal/tui/`.

The redesign's principle: **the interface should narrate the music**. Same keys, same two-layer model (play view + control center), same indigo theme — but every surface now says what the instrument is doing.

Work through the README screen by screen:
1. Play view — station identity header, live musical narration line, form rail (replaces the current anonymous top/playback bars).
2. Control center — full-screen layout with ●●●○○ value scales and visible word ladders for the nine music macros.
3. Track library — form map with per-section bars, durations, harmony, textures, and tags.
4. Splash/loading — block TERMUS wordmark + browsable station dial.
5. Help overlay — full grouped key reference (re-advertise the hidden power-user keys).

Constraints:
- Keep every existing keybinding and behavior; this is a presentation-layer redesign.
- Reuse existing state — the narration line is built from `gen.DebugStatus` (bar/section/chord/preset) and `TrackNavEntry.Structure`; do not invent new engine plumbing beyond wiring what exists.
- All colors come from the active `ColorTheme` (BarFg/BarHi/gradient) so all six themes keep working; the README's hex values are the indigo theme only.
- Respect `useCompactLayout` and the ≥40×10 minimum; degrade gracefully (drop the narration line first, then the form rail).
- Update golden tests / add new ones as the views change.

Start by reading `design_handoff_tui_redesign/README.md` fully, then `internal/tui/model.go` (View, topBar, playbackBar, bottomBar), `control_center.go` (controlsPanel), `tracks.go` (trackPanel), and `themes.go`. Make a plan, then implement screen by screen with a test pass after each.

---
