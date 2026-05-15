# termus Music Realism Plan

**Goal:** Push the generators from "harmonically valid" toward "performed and arranged like real music" by improving articulation, phrase targeting, form, drum vocabulary, and evaluation.

**Priority order:** articulation engine, bar-quantized macro scheduler, lookahead phrase resolution, form engine, drum cadence vocabulary, arrangement-first variation, deeper voice leading, listening harness.

## 1. Articulation Layer

- **Problem:** `SF2Track` currently humanizes timing and velocity, but most notes still share the same effective note length and attack/release behavior.
- **Primary files:** `internal/gen/sf2_engine.go`, `internal/gen/jazz.go`, `internal/gen/chill.go`, `internal/gen/sf2_markov.go`, `internal/gen/sf2_pentatonic.go`
- **Deliverable:** extend `SF2Track` with note-length and phrasing controls such as gate percentage, overlap/legato, velocity resolver hooks, and optional CC11 expression curves.
- **Acceptance:** sustained parts can overlap on purpose, stabs can release early, and lead voices can swell or taper without relying only on random jitter.

## 2. Bar-Quantized Macro Scheduling

- **Problem:** section toggles and key drift are currently driven by elapsed samples, so they can happen at musically awkward points.
- **Primary files:** `internal/gen/ambient.go`, `internal/gen/chill.go`, `internal/gen/jazz.go`, `internal/gen/phase.go`, `internal/gen/sf2_glass.go`, `internal/gen/sf2_markov.go`, `internal/gen/sf2_pentatonic.go`
- **Deliverable:** add shared helpers for "next bar", "next 4 bars", and "next phrase boundary" scheduling, then move macro mutation timers onto those boundaries.
- **Acceptance:** entrances, exits, key drifts, and arrangement changes happen at consistent musical seams instead of mid-phrase.

## 3. Lookahead Phrase Resolution

- **Problem:** `jazz` and `lofi` phrases now fit the current harmony better, but they still do limited forward-targeting into the next chord.
- **Primary files:** `internal/gen/jazz.go`, `internal/gen/chill.go`, `internal/gen/scale_helpers.go`
- **Deliverable:** teach phrase planners to target upcoming harmony with pickups, approach tones, enclosures, suspensions, and delayed resolutions across barlines.
- **Acceptance:** bar transitions sound intentional, and melodic lines create expectation before the next chord arrives.

## 4. Top-Level Form Engine

- **Problem:** most generators still live inside repeating 4- or 8-bar cells, with optional layers toggled on top.
- **Primary files:** `internal/gen/algorithm.go`, `internal/gen/playlist.go`, `internal/gen/chill.go`, `internal/gen/jazz.go`, `internal/gen/sf2_markov.go`, `internal/gen/ambient.go`
- **Deliverable:** add a form scheduler that can plan intros, A/A'/B sections, breakdowns, rebuilds, cadences, and outros on top of per-bar note generation.
- **Acceptance:** long playback develops in recognizable sections rather than only mutating a single loop forever.

## 5. Drum Cadence Vocabulary

- **Problem:** the drum parts are stronger than before but still rely heavily on static grids plus probability, especially at phrase endings.
- **Primary files:** `internal/gen/chill.go`, `internal/gen/jazz.go`
- **Deliverable:** add phrase-ending fills, ghost-note cells, pickup bars, crash/drop moments, and section-specific groove variants.
- **Acceptance:** every 4 or 8 bars the rhythm section acknowledges the form instead of staying metrically correct but narratively flat.

## 6. Arrangement-First Variation

- **Problem:** random GM program swaps add variety, but they can also break the identity of the ensemble.
- **Primary files:** `internal/gen/chill.go`, `internal/gen/jazz.go`, `internal/gen/sf2_engine.go`
- **Deliverable:** prefer arrangement mutations such as muting, doubling, octave shifts, density changes, expression changes, or wet/dry moves before resorting to timbre swaps.
- **Acceptance:** long-form variation feels like a band re-arranging the material, not like a different preset being loaded into the same part.

## 7. Deeper Voice Leading

- **Problem:** several generators now choose legal notes, but inner voices still use limited inversion logic.
- **Primary files:** `internal/gen/jazz.go`, `internal/gen/sf2_markov.go`, `internal/gen/sf2_pentatonic.go`, `internal/gen/scale_helpers.go`
- **Deliverable:** add nearest-inversion and common-tone retention helpers so comping and pads choose the smallest useful movement from one chord to the next.
- **Acceptance:** chord changes read as connected lines, not as isolated valid sonorities.

## 8. Listening Harness

- **Problem:** current tests mainly cover determinism and non-silence; they do not help evaluate musical regressions.
- **Primary files:** `cmd/termus-debug/main.go`, `internal/gen/*_test.go`, new harness assets under `testdata/`
- **Deliverable:** add a seed corpus renderer that produces short WAV snapshots and verifies structural invariants like cadence alignment, section timing, and repeatability.
- **Acceptance:** musical changes can be regression-tested with stable seeds and short render windows before relying on manual listening.
