# TODO

## UI Improvements
- [x] Auto-detect terminal color and adapt colors and controls.
- [x] Improve alternate visualizers so they match the default scope's line-based design language.
- [x] Add a `?` help overlay and reduce the always-visible footer chrome.
- [x] Persist kept seeds to disk and add a browsable saved-seed library overlay.
- [x] Add a proper now-playing strip with elapsed time, playlist progress, crossfade countdown, and recording duration.
- [x] Add a compact audio meter and clip indicator.
- [x] Add a seed and track inspector overlay with algo, seed, A/B slots, kept count, chord, section, bar, and export actions.
- [x] Add an in-TUI export drawer for record, WAV, stems, and MIDI actions.
- [x] Improve narrow-terminal behavior with a simplified compact layout.
- [x] Add a reduced-chrome mode.
- [x] Add a startup splash / onboarding screen for the core controls.
- [x] Add transient volume feedback as a thin symmetric line in the existing color scheme that appears only while volume changes.

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
- [x] Surface chord, section, bar, and SF2 preset state in the TUI for debugging and tuning.

## Music Quality Backlog

- [x] Fix true legato, note ties, and controlled overlap in `internal/gen/sf2_engine.go` so repeated-pitch phrases can connect instead of always re-articulating.
- [x] Add deterministic groove templates for `jazz` and `chill` so timing feel comes from role-based pocket, not only jitter and swing offsets.
- [x] Add motif memory above the current form engine so A / A' / B / cadence sections can recall, sequence, and answer shared musical cells.
- [x] Expand harmonic language in `jazz` and `chill` with reharmonization, borrowed changes, secondary dominants, and deceptive turns.
- [x] Replace coarse section energy shifts with per-instrument section scenes for lead, comp, bass, drums, and texture layers.
- [x] Add explicit call-response / dialogue behavior between lead and accompaniment parts so phrases stop competing for the same space.
- [x] Improve the tuning workflow with stem and MIDI export support alongside the existing listencheck renders.

## Next Music Pass

- [x] Extend the motif-memory, section-scene, and dialogue treatment into the ambient-family generators (`ambient`, `glass`, `drone`, `phase`, `lullaby`) so the quieter modes evolve like composed textures instead of independent loops.
- [x] Add phrase scoring and seed ranking to `cmd/termus-listencheck` so seed triage can use cadence, repetition, occupancy, and harmonic-motion metrics instead of manual listening alone.
- [x] Bring stem and MIDI export into the main `./termus` binary so direct renders can emit full tuning artifacts without going through `listencheck`.
- [x] Add richer SF2 performance modulation such as vibrato curves, note-level drift, and brightness shaping so sustained lines sound less static.
- [x] Make drum writing more phrase-based with 2-bar / 4-bar memory, stronger fill targeting, and bar-to-bar anti-repetition in `jazz` and `chill`.
- [x] Add a TUI seed browser / A-B workflow so generated takes can be compared, replayed, and kept or rejected quickly during tuning.

## Control Center

- [x] Add a toggleable control-center panel for advanced music, session, curation, and audio actions, and document it in the `?` help overlay without expanding the default chrome.
- [x] Add saveable sessions that persist the current algorithm, seed, visual, theme, volume, and playback context for later recall.
- [x] Add favorites, ratings, tags, recent history, and best-takes browsing for seeds so curation can happen inside the app instead of only through raw seed numbers.
- [x] Add live musical controls behind the panel, including density, brightness, swing, reverb, motion, and drone-depth style macros where the active algorithm supports them.
- [ ] Add live seed morphing so the app can glide between takes instead of only hard-switching seeds.
- [ ] Add tempo-aware rhythmic controls for groove-based genres and phrase-length controls for ambient-style genres behind the same panel.
- [ ] Add explicit audio recovery controls such as retry live audio and render-only fallback actions behind the panel.
- [ ] Improve endings with graceful cadences and export outros so rendered tracks stop feeling abruptly truncated.
