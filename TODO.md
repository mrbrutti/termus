# TODO

## UI Improvements
- [x] Auto-detect terminal color and adapt colors and controls.

## Music Realism

- [x] Add an articulation layer in `internal/gen/sf2_engine.go` with per-slot gate, legato/overlap, dynamic velocity resolution, and optional expression curves so phrases stop sounding uniformly tongued.
- [x] Quantize macro events to musical boundaries so section toggles, key drift, and arrangement changes land on the next bar or phrase instead of mid-phrase.
- [x] Teach `jazz` and `lofi` phrase generators to target upcoming harmony with pickups, approach tones, suspensions, and delayed resolutions instead of resolving only against the current chord.
- [x] Add a top-level form engine that can schedule intros, A/A'/B sections, breakdowns, cadences, and outros above the repeating bar loops.
- [x] Expand drum writing with cadence-aware fills, ghost-note cells, pickup bars, crash/drop moments, and section-specific variation for `jazz` and `lofi`.
- [x] Replace random program swaps with arrangement-first changes such as muting, doubling, register shifts, wet/dry changes, and density changes that preserve ensemble identity.
- [x] Upgrade chord voicing and inner parts to choose nearest inversions and true voice-leading paths instead of alternating between a small fixed set of valid shapes.
- [x] Build a listening harness that renders a seed corpus to short WAVs and checks bar alignment, cadence landings, section boundaries, and regression snapshots.

## Next Up

- [x] Add a proper audio backend state layer so startup can report `starting`, `ready`, `no default device`, `backend hung`, and `render-only` instead of failing opaquely on bad CoreAudio state.
- [x] Add single-track WAV export to `./termus` via `--out` and `--seconds` so music generation works without live playback.
- [x] Extend export workflow with `--playlist-out` and batch rendering from the main binary.
- [x] Tighten cross-algorithm intros, outros, cadence handling, and loudness normalization so switching genres feels cohesive.
- [ ] Surface chord, section, bar, and SF2 preset state in the TUI for debugging and tuning.
