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
- [x] Add live seed morphing so the app can glide between takes instead of only hard-switching seeds.
- [x] Add tempo-aware rhythmic controls for groove-based genres and phrase-length controls for ambient-style genres behind the same panel.
- [x] Add explicit audio recovery controls such as retry live audio and render-only fallback actions behind the panel.
- [x] Improve endings with graceful cadences and export outros so rendered tracks stop feeling abruptly truncated.

## Control Center Redesign

- [x] Make the default bottom status bar nearly empty, showing only the current music type plus very light `?` / `m` affordances and transient center status.
- [x] Rewrite the `?` help overlay so it teaches the two-layer model (`play view` + `control center`) instead of listing the full hotkey inventory.
- [x] Expand the control center from the current tabbed overlay into a left-nav general menu with sections for now, look, music, seeds, library, export, audio, and debug.
- [x] Move current overlay-first actions such as visual/theme switching, library browsing, export actions, debug toggles, and seed curation under the control center so the main interface stays minimal.
- [x] Reduce the visible global interaction model to the essential keys (`space`, `↑↓`, `m`, `?`, `q`) while keeping deeper actions discoverable inside the control center.

## Long-Form Listening

### Episode Structure
- [x] Replace fixed looping forms with a streaming episode planner so `jazz`, `lofi`, and `classical` stop wrapping on one short bar cycle.

### Long Horizon State
- [x] Add a long-horizon conductor state that carries harmony family, motif family, texture scene, density bias, and movement identity across episodes.

### Episode Regeneration
- [x] Regenerate progressions, motifs, comp/drum plans, and orchestration choices at episode boundaries instead of replaying one frozen cycle forever.

### Anti-Repetition Memory
- [x] Add recent-history memory and penalties for reused progressions, motif shapes, fills, and cadence types so long playback avoids obvious returns.

### Time Layers
- [x] Split scheduling into explicit note, bar, section, and episode layers so local phrasing and long-form reinvention stop fighting each other.

### Ambient Evolution
- [x] Give the ambient-family generators slow contour, register, loop-length, and foreground/background regeneration so they evolve over hours instead of circling one constellation.

### Movement Mode
- [x] Add a higher-level movement sequence for long listening so one piece can drift through establish, deepen, subtract, brighten/darken, and near-return phases.

### Motif Transformation
- [x] Revisit motifs by transforming them over time (transpose, invert, stretch, reharmonize, re-orchestrate) instead of only repeating or replacing them outright.

## Product Identity

### Stations
- [x] Replace the raw genre-facing presentation with a curated station pass so the app surfaces named moods like `night drift`, `glass chapel`, and `soft tape` instead of only algorithm labels.

### Listening Modes
- [x] Add explicit long-form listening modes such as `endless`, `album side`, `hour stream`, and `radio`, and wire them into live playback plus offline render/export defaults.

### Visual Polish
- [x] Give the TUI one more signature visualization pass with subtle color breathing, smoother visual switching, and at least one uniquely `termus` visual mode that still matches the minimal line language.

## Style Rule Pass

### MAX Palette Rules
- [x] Reframe `max` as a curated role-based SoundFont pool per algorithm instead of always-on full-bank layering, so each station can borrow a wider palette without losing its core identity.

### Glass Chapel
- [x] Fix `Glass Chapel`'s shared-channel pad interference, reduce trashy texture interactions, and retune its `max` palette / voicing so the bell station stays clean and luminous instead of brittle.

### Jazz Guide Rules
- [x] Encode jazz style-guide rules into `jazz`: stronger guide-tone motion, 2-bar / 4-bar phrase sentences, better turnaround targeting, and more conversational comping/solo interaction.

### Lo-fi Guide Rules
- [x] Encode lo-fi style-guide rules into `lofi`: groove pocket over jitter, richer extension vocabulary, more soulful loop mutation, and calmer/supportive texture use.

### Shared Phrase Grammar
- [x] Add a shared melodic phrase-grammar layer with question/answer, pickup/peak/release, and cadence-aware endings so melodic genres sound more composed and less slot-generated.

### Genre Density Policy
- [x] Add genre-specific density policies so note activity increases in `jazz` / `lofi` while `ambient` / `bells` / `drone` stay selective and spacious over long listening.

## TM Composition Language

### Schema
- [x] Add a richer `.tm` composition language with structured metadata plus embedded mini-languages for rhythm, melody, harmony, and arrangement so AI can author auditable long-form pieces instead of only steering generators by seed.

### Parser And Linter
- [x] Implement `.tm` parsing, validation, and linting for section durations, algorithm names, listening modes, pattern syntax, and structural contrast checks.

### Compilation
- [x] Compile `.tm` compositions into the existing playback/export pipeline as authored scored playlists with per-section control profiles and section titles.

### CLI Support
- [x] Add CLI support for loading `.tm` files in live playback and batch export paths without exposing a new user REPL.

### Authored Pack
- [x] Add an initial authored `.tm` pack for `lofi`, `jazz`, and `bells` so we can test whether authored long-form pieces sound more like real music than pure seed-driven generation.

### Score-Driven Lo-fi
- [x] Make `.tm` audit fields actually drive the `lofi` generator's harmony, lead contour, comp rhythm, groove density, and role activation instead of only acting as annotations on top of seed/profile changes.
- [x] Expand the authored lo-fi score library with longer, more contrast-heavy studies so we can audition multiple believable tape-era moods instead of one thin example.

## Mature Authored Engine

### Phase 1: Make Tracks Authoritative
- [x] Replace the remaining generator-first track path with one unified authored playback engine so `-track` always renders from the score IR instead of falling back to genre-specific composition habits.
- [x] Compile each section into explicit role timelines and phrase blocks rather than one continuous slot stream per role, so verse / bridge / breakdown / outro can contain genuinely different written material.
- [x] Preserve full chord quality, extensions, slash bass, suspensions, and borrowed colors in the IR instead of flattening harmony into coarse major/minor/dominant buckets.
- [x] Add section-local phrase ownership for bass, comp, lead, texture, and drums so a track can write different material per section instead of reusing one motif family all song.
- [x] Support section inheritance and transforms in `.tm` (`derive`, `sequence`, `invert`, `thin`, `lift-register`, `cadence rewrite`) so authors can write A / A' / B relationships without duplicating raw material.

### Phase 2: Expand The `.tm` Language
- [x] Add a first-class arrangement block to `.tm` with explicit scene events such as fills, stop-time bars, pedal holds, swells, doubles, breaks, tags, pickups, and endings.
- [x] Add phrase-block syntax to `.tm` for section-local melody, comp, bass, and drum writing instead of only global role motifs and patterns.
- [x] Add orchestration directives to `.tm` so roles can change instrument family, register, articulation, or prominence by section without redefining the whole track.
- [x] Add track-level variation budgets and anti-repetition constraints to `.tm` so authors can say how much mutation is allowed per section, phrase, or return.
- [x] Add linter rules for `.tm` that flag weak contrast, over-dense writing, missing cadence shape, too many simultaneous bright attacks, and sections that are too similar.

### Phase 3: Soundfont And Instrument Intelligence
- [x] Evolve `sf2_inventory.json` from a rough bank catalog into a role-aware program inventory with family, articulation, tone, realism, and blend tags for actual instrument selection.
- [x] Move `pro` and `max` selection onto an orchestration resolver that picks bank + program by role intent, style pack, register, and ensemble cohesion instead of mostly static mappings.
- [ ] Support section-local instrument substitutions and doubles so one track can move from piano trio to organ combo, or from celesta lead to choir/pad answer, without changing styles entirely.
- [ ] Add cohesion scoring to prevent `max` from sounding like unrelated banks piled together, especially for soft/wet genres like `bells`, `ambient`, and `lullaby`.

### Phase 4: Performance Realism
- [ ] Replace generic slot repetition with phrase-aware performer logic for drummer, bassist, comp, and lead roles so timing, accent, and articulation respond to phrase purpose.
- [ ] Add richer section-local dynamics: crescendos, decrescendos, breath points, held endings, phrase peaks, and drop-to-silence moments that feel performed rather than enumerated.
- [ ] Build deeper drum vocabularies per style with bar-start pickups, turnaround bars, fills, ghost-note patterns, ride/hat switching, stop choruses, and ending cadences.
- [ ] Build deeper bass vocabularies per style with pedal notes, walks, anticipations, chromatic approaches, descents, and sparse anchor modes instead of one generalized line behavior.
- [ ] Build deeper comp vocabularies per style with real voicing families, rhythmic cells, sectional stabs, held pads, and answer figures instead of one generalized support stream.

### Phase 5: Style Packs And Genre Identity
- [ ] Define explicit style packs that interpret the same `.tm` language differently for `lofi`, `jazz`, `ambient`, `bells`, `classical`, `drone`, `phase`, and `lullaby`.
- [ ] Give each style pack multiple substyles so tracks within a genre can diverge meaningfully, for example multiple lofi bands, multiple jazz ensembles, and multiple bell/ambient schools.
- [ ] Add genre-specific phrase and arrangement libraries drawn from real musical practice: heads, shout choruses, turnarounds, suspended bridges, tape-beat breakdowns, nocturne cadences, and devotional bell scenes.
- [ ] Curate at least 10 clearly differentiated authored tracks per genre once the unified engine is in place, and reject or rewrite tracks that still collapse into one house sound.

### Phase 6: Evaluation And Curation
- [ ] Add renderer-side metrics for repetition fatigue, section contrast, cadence spacing, lead occupancy, register spread, harmonic color retention, and ensemble diversity.
- [ ] Add bundled listening corpora and A/B snapshots per genre so we can hear whether engine changes make tracks more song-like or more synthetic.
- [ ] Add a track-review workflow that renders one canonical example per authored track and surfaces the arrangement map for faster curation.
- [ ] Add a strict authored-track test gate so new `tracks/<genre>/*.tm` files must compile, include arrangement moments, show section contrast, and meet minimum diversity thresholds.

### Phase 7: Product Fit
- [ ] Make the track browser show compact but useful authored metadata such as section count, ensemble, substyle glyphs, and arrangement complexity without wasting vertical space.
- [ ] Add a track-structure inspector in the control center so we can see section order, events, harmony, and active ensemble while a score is playing.
- [ ] Make default startup favor the authored track library once the unified engine is mature enough that curated tracks consistently outperform raw procedural playback.
