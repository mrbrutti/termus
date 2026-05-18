# termus

`termus` is a terminal-native generative music instrument. It can run live for hours, render long-form pieces to disk, build playlists, export stems and MIDI, and surface deeper controls through a minimal TUI plus a control center.

- Live stations with curated identities like `Night Drift`, `Soft Tape`, and `Dusty Swing`
- Long-form listening modes: `endless`, `album-side`, `hour-stream`, and `radio`
- Live playback, direct WAV rendering, playlist rendering, stems, and MIDI export
- Minimal play view with a deeper control center for music, seeds, library, export, audio, and debug

## Quick Start

```bash
go install github.com/mrbrutti/termus/cmd/termus@latest
termus
```

Or from source:

```bash
git clone https://github.com/mrbrutti/termus
cd termus
go run ./cmd/termus
```

Sample invocations:

```bash
# Default live station
termus

# Long-form jazz station with preferred per-algorithm SoundFonts
termus --algo jazz --listen-mode hour-stream --sf2-strategy pro

# Render an album-side ambient piece to disk
termus --algo ambient --listen-mode album-side --out ambient-side-a.wav

# Render a radio-style mixed playlist with WAVs + manifest
termus --listen-mode radio --playlist-out ./radio-set

# Render stems and MIDI when supported
termus --algo lofi --out soft-tape.wav --stems --midi
```

## Stations And Algorithms

The CLI still uses genre-style `--algo` names. The TUI and playlists surface curated station labels, and the top bar / control center show both where useful, for example `Night Drift · ambient`.

| `--algo` | Station label | Character | Synth fallback |
| - | - | - | - |
| `ambient` | `Night Drift` | Slow ambient chord drift, bells, and hovering textures | `ambient-synth` |
| `drone` | `Deep Field` | Sustained low-motion beds and metallic shimmer | `drone-synth` |
| `bells` | `Glass Chapel` | Bright bells, celesta-like figures, reflective space | `bells-synth` |
| `lullaby` | `Sleep Walk` | Pentatonic lullaby textures that stay consonant | `lullaby-synth` |
| `classical` | `Chamber Loop` | Chamber-like melodic writing and evolving form | `classical-synth` |
| `phase` | `Slow Signal` | Reich-style phased motion and repeating pattern drift | sampled only |
| `lofi` | `Soft Tape` | Beat-driven tape-warm grooves and melodic pocket | sampled only |
| `jazz` | `Dusty Swing` | Walking bass, comping, drums, and lead phrasing | sampled only |

Notes:

- Legacy names such as `eno`, `chill`, `glass`, `markov-sf2`, and `sf2` still resolve.
- The sampled algorithms can run with a curated SoundFont preset or your own `.sf2` file.
- The long-form engine now evolves by episode and movement instead of short fixed loops.

## Engines

termus has two render engines. The engine is selected per-track in the `.tm` file via the `render_engine:` field.

| `render_engine:` | Description |
| - | - |
| `sf2` (default, or unset) | The original procedural engine. Authored compiler + algorithmic generators + SoundFont sampler. Pure Go. No external dependencies beyond bundled `.sf2` files. Used by every track up through v2 of the `.tm` schema and remains the default. |
| `acestep` (opt-in, SP21) | AI generation via the ACE-Step diffusion model running in a local Python service. Streaming playback with look-ahead queueing and equal-power crossfade between successive renders. Requires `services/acestep/install.sh` (5-10 GB model download) and the daemon to be running. Played via `termus-stream` rather than `termus`. |

Old tracks with no `render_engine:` field continue to compile and render unchanged on the SF2 path. The ACE-Step path is documented in [`docs/acestep-engine.md`](docs/acestep-engine.md); the reference v3 track is `tracks/lofi/bookstore-rainy-night-v3.tm`.

## Listening Modes

Listening modes shape the session profile and some export defaults.

| `--listen-mode` | What it does |
| - | - |
| `endless` | Default live mode. Single evolving stream. Offline `--out` defaults to 180 seconds. |
| `album-side` | Longer, more deliberate arc. Offline `--out` defaults to 24 minutes. |
| `hour-stream` | One continuous hour-scale piece. Offline `--out` defaults to 60 minutes. |
| `radio` | Auto-configures a mixed playlist feel. Defaults to `--playlist mixed`, 8 tracks, ~7m30s per track. |

Notes:

- `--listen-mode radio` works for live playback and `--playlist-out`.
- `--listen-mode radio` is not valid with direct single-file `--out`.

## CLI Options

### Core playback

| Flag | Description |
| - | - |
| `--algo NAME` | Select algorithm/station source. |
| `--seed N` | Set deterministic seed. Same seed recreates the same starting world. |
| `--volume 0..100` | Initial live/output volume. |
| `--listen-mode MODE` | `endless`, `album-side`, `hour-stream`, `radio`. |
| `--debug` | Start the TUI with the musical debug inspector visible. |

### Sound and space

| Flag | Description |
| - | - |
| `--sf2 PATH` | Use a custom SoundFont file. Overrides preset selection. |
| `--sf2-preset NAME` | Choose a curated preset such as `general` or `sgm`. |
| `--sf2-strategy single\|pro\|max` | `single` uses one preset everywhere, `pro` loads each algorithm's preferred preset, and `max` preloads the full curated SoundFont catalog. |
| `--ir room\|hall\|cathedral\|plate\|PATH` | Apply convolution reverb from a preset or WAV impulse response. |
| `--ir-wet 0..1` | Wet mix for `--ir`. |

### Playlists and rendering

| Flag | Description |
| - | - |
| `--playlist same\|mixed` | Live or batch playlist mode. `same` varies seeds of the chosen algorithm; `mixed` rotates genres. |
| `--playlist-tracks N` | Playlist length. |
| `--playlist-duration DURATION` | Per-track duration before crossfade. |
| `--out FILE.wav` | Render one piece to a WAV instead of starting live playback. |
| `--seconds N` | Duration for `--out`. If omitted, the listening mode default is used. |
| `--playlist-out DIR` | Render a playlist to a directory of WAVs plus `manifest.json`. |
| `--stems` | With `--out` or `--playlist-out`, also export per-stem WAVs when supported. |
| `--midi` | With `--out` or `--playlist-out`, also export MIDI when supported. |

## TUI Model

The live app now has a two-layer model:

1. Play view: visualizer first, minimal chrome.
2. Control center: the place for nearly everything deeper.

### Global keys

| Key | Action |
| - | - |
| `space` | Play / pause |
| `↑` `↓` `+` `-` | Volume up / down |
| `m` | Open or close the control center |
| `?` | Open or close help |
| `q` or `Ctrl-C` | Quit |

The footer intentionally stays minimal. Most other actions live in the control center instead of the main play view.

Power-user note:

- Older direct shortcuts still work even though they are no longer advertised in the footer. If you already know keys like `c`, `C`, `n`, `p`, `d`, `l`, `i`, `e`, `z`, `[`, `]`, `a`, `b`, `tab`, `k`, or `x`, you can still use them directly.

### Control center

Inside the control center:

- `↑` `↓` browse rows
- `←` `→` adjust the current value
- `Enter` apply, toggle, or open
- `Tab` move to the next section

Sections:

- `Now`: playback, listening mode summary, track status, recording, playlist skip
- `Look`: visual mode, theme, chrome, help state
- `Music`: density, brightness, motion, reverb, swing, drone depth, tempo, phrase length, seed morph
- `Seeds`: algorithm switching, seed browsing, A/B slots, compare, keep, reject
- `Library`: saved seeds, rating, favorites, tags, recent history, best takes, saved sessions
- `Export`: 60-second live export drawer for WAV, MIDI, stems, and recording control
- `Audio`: backend status, retry live audio, render-only fallback
- `Debug`: toggle overlay plus bar / section / chord / preset state

### Help, chrome, and debug

- `?` shows the current two-layer help overlay.
- The debug overlay can be toggled from the control center or enabled on startup with `--debug`.
- Reduced chrome / zen behavior is handled through the `Look` section instead of adding more always-visible keys.

## Visualizers And Themes

Current visual modes:

- `scope`: the default thin waveform trace
- `contour`: a center-weighted spectral horizon
- `vector`: an expanded stereo phase portrait
- `signal`: a `termus`-native carrier trace with restrained ghost echoes
- `drift`: string-like horizontal vibration lines

Current themes:

- `indigo`
- `amber`
- `matrix`
- `magenta`
- `mono`
- `rainbow`

The non-default visualizers use the same minimal braille line language as the default view, with pulse-based color breathing and smoother visual transitions.

## Rendering And Export

### Direct offline renders

```bash
termus --algo ambient --out piece.wav
termus --algo jazz --listen-mode album-side --out side-b.wav
termus --algo lofi --out beat.wav --stems --midi
```

Behavior:

- `--out` writes one WAV file.
- `--seconds` controls duration unless the listening mode supplies a longer default.
- When possible, renders snap to a nearby cadence/outro instead of hard-cutting.
- `--stems` writes a sibling `-stems/` directory.
- `--midi` writes a sibling `.mid` file when the algorithm supports MIDI capture/export.

### Playlist renders

```bash
termus --playlist same --algo bells --playlist-out ./bells-set
termus --listen-mode radio --playlist-out ./radio-set --stems --midi
```

Behavior:

- `--playlist-out` renders one file per track.
- A `manifest.json` is written alongside the audio.
- `--stems` and `--midi` can be added for per-track artifacts when supported.

### In-TUI export drawer

The live app also has a built-in export drawer in the control center:

- WAV export
- MIDI export
- stem export
- live recording toggle

Live exports currently render 60-second artifacts into `./exports/`.

## Seeds, Library, And Sessions

The app now includes built-in curation tools:

- browse seeds for the current algorithm
- store A/B seeds and compare them
- keep or reject takes
- rate and favorite seeds
- tag seeds
- browse recent history and best takes
- save and reload sessions that capture algo, seed, visual, theme, volume, and related state

This makes `termus` usable as a real exploration and capture workflow, not only a passive generator.

## SoundFonts

If you use sampled algorithms, `termus` can auto-download curated SoundFonts into your user cache directory on first use.

| Preset | Size | Best for |
| - | - | - |
| `general` | 32 MB | Balanced default GM bank |
| `sgm` | 325 MB | Piano, guitar, bass, and warm groove work |
| `tyros4` | 502 MB | Jazz brass, sax, show-band color |
| `dsound4` | 553 MB | Large balanced alternative |
| `fatboy` | 315 MB | Loudness-matched GM, clean lo-fi / baroque |
| `timbres-of-heaven` | 377 MB | Classical / orchestral writing |
| `merlin-symphony` | 163 MB | Alternate orchestral bank |
| `fairy-tale` | 200 MB | Bells, celesta, lullaby palettes |
| `fm-dx` | 124 MB | FM textures for drone / phase |
| `musescore-general` | 208 MB | Polite neutral GM |
| `arachno` | 148 MB | Retro pad / bell / lead color |

Recommended:

```bash
termus --sf2-strategy pro
termus --sf2-strategy max
```

- `pro` loads the preferred preset for each sampled algorithm.
- `max` preloads the entire curated catalog for instant switching and experimentation.

## Reverb / IR Presets

`--ir` accepts either a preset name or a WAV path:

| Value | Description |
| - | - |
| `room` | Tight early reflections |
| `hall` | Concert-hall tail |
| `cathedral` | Long, spacious tail |
| `plate` | Dense synthetic plate |
| `PATH.wav` | Custom 16-bit PCM WAV impulse response |

## How It Works

`termus` combines several layers:

- algorithm-specific note and phrase generation
- long-form episode and movement planning
- anti-repetition memory
- voice-leading, motif recall, and arrangement changes
- live TUI state and export tooling

Under the hood it uses:

- pure synthesis for the synth variants
- `go-meltysynth` for sampled SoundFont playback
- optional convolution reverb
- Bubble Tea + Lip Gloss for the TUI
- braille-grid rendering for the visual system

## Dependencies

| Library | License | Purpose |
| - | - | - |
| [`gopxl/beep/v2`](https://github.com/gopxl/beep) | MIT | Audio output and streaming |
| [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) | MIT | TUI framework |
| [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) | MIT | Terminal styling |
| [`sinshu/go-meltysynth`](https://github.com/sinshu/go-meltysynth) | MIT | Pure-Go SoundFont synthesizer |
| [`madelynnblue/go-dsp`](https://github.com/madelynnblue/go-dsp) | BSD | FFT support for DSP / visuals |

## License

MIT — see [LICENSE](LICENSE).
