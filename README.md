# termus

A terminal music player that **generates ambient and lofi music from scratch in real time**. No tracks, no samples-on-disk, no playlist — every second of audio is synthesized or rendered through a SoundFont on the fly by one of nine generative algorithms. Run it in your terminal, hit play, and it produces music that never repeats and can keep going for hours.

```
termus · chill · seed=42 · vol 70%

⠀⠀⠀⠀⠠⢤⡀⠀⠀⠀⠀⠀⠀⢀⡠⠤⠀⠀⠀⠀⠀⠀⠀⠐⠂⢄⡀⠀⠀⠀⠀⠀⠀⠀⠠⠴⠉⠀⠀⠀⠀⠀⠀⠀⠐⠈⠂⢄⡀⠀
⠀⠀⠀⠀⠀⠀⠉⠐⠉⠉⠁⠈⠁⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠊⠘⠉⠁⠁⠐⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁

[space] play  [↑↓] vol 70%  [r] rec  [c] indigo  [q] quit
```

(Above: a colored Braille oscilloscope tracking the synthesizer's output, with selectable themes.)

## Quick start

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

First run with a SoundFont-based algorithm (`--algo sf2`, `eno-sf2`, etc.) auto-downloads a SoundFont into `~/Library/Caches/termus/soundfonts/` (or the OS equivalent). The catalog:

| Preset | Size | Best for |
|-|-|-|
| `general` (default) | 32 MB | Balanced GM — fallback for everything. MIT. |
| `sgm` | 325 MB | Piano / guitar / bass focus. |
| `tyros4` | 502 MB | **Jazz** — Tyros 4 brass, sax, walking bass, brushed kits. |
| `dsound4` | 553 MB | Large balanced alt to `general`. |
| `fatboy` | 315 MB | Loudness-matched GM. Good for clean lo-fi / baroque. MIT-ish. |
| `roland-sc55-up` | 177 MB | '90s Roland GS pads / FM-EP — ambient, lo-fi authenticity. |
| `timbres-of-heaven` | 377 MB | **Classical** — strings, brass, woodwinds. |
| `merlin-symphony` | 163 MB | Alt orchestral. |
| `fairy-tale` | 200 MB | **Bells / lullaby** — celesta, music-box, glockenspiel. CC-BY-NC-SA. |
| `fm-dx` | 124 MB | **Drone / phase** — DX-style FM EPs, metallic bells. |
| `musescore-general` | 208 MB | Polite, neutral GM. Safe legal status. |
| `arachno` | 148 MB | D-50 / M1 / MU / Fairlight blend. CC-BY-NC-SA. |

Use `--sf2-strategy optimal` to download each genre's preferred SoundFont automatically.

```bash
termus --algo chill              # default: GeneralUser-GS (32 MB)
termus --algo chill --sf2-preset sgm   # audiophile: SGM (325 MB on first run)
termus --algo chill --sf2 ~/Music/MyFavorite.sf2   # your own file
```

## Algorithms

Termus ships eight genre-named algorithms, each with an SF2 sampled version (default) and a `-synth` pure-synthesis fallback (no download needed). All produce indefinitely-long output via per-note mutation, macro key drift, instrument swaps, and section toggling.

| `--algo` | Genre | What it is |
|-|-|-|
| `ambient` | Ambient | Music for Airports — pad-bell on incommensurate loops, sampled |
| `drone` | Drone | Stars of the Lid — held strings + flute shimmer over deep bed |
| `bells` | Bells | Tubular bells + crystal pad — bright, late-night focus |
| `lullaby` | Lullaby | Pentatonic random walk — piano, harp, kalimba, never clashes |
| `classical` | Classical | Markov melody on piano + strings + clarinet — feels composed |
| `phase` | Phase | Reich-style — two vibraphones drift in tempo, ever-changing pattern |
| `lofi` | Lo-fi | Hip-hop drums + Rhodes EP + walking bass + sax + nylon guitar |
| `jazz` | Jazz | Medium-swing small group — walking bass, ride pattern, Charleston comp, brushed kit, alto-sax solo over ii-V-I cycles |

Each genre name has a `-synth` variant (e.g. `ambient-synth`, `lofi-synth` (n/a — only ambient/drone/bells/lullaby/classical have synth variants)) that uses pure synthesis instead of a SoundFont — useful if you want to skip the SoundFont download.

Legacy algorithm names (`eno`, `eno-sf2`, `chill`, `glass`, `markov-sf2`, etc.) still work and resolve to the corresponding genre name.

## Usage

```bash
termus [--algo NAME] [--seed N] [--volume 0..100]
       [--sf2 PATH] [--ir NAME-OR-PATH] [--ir-wet 0..1]
```

Examples:

```bash
# Default — sampled ambient (Music for Airports style)
termus

# Pick a genre + seed (same seed = same music)
termus --algo lofi --seed 42
termus --algo classical --seed 99

# Jazz in a cathedral
termus --algo jazz --ir cathedral

# Bring your own impulse response
termus --algo ambient --ir ~/Downloads/concert-hall.wav --ir-wet 0.4

# Use the 325 MB high-quality SoundFont
termus --algo lofi --sf2-preset sgm

# Skip the SoundFont download (pure synthesis)
termus --algo ambient-synth
```

### `--ir` presets

| Preset | What it is |
|-|-|
| `room` | ~80 ms early reflections — close, intimate |
| `hall` | ~1.5 s concert-hall tail |
| `cathedral` | ~3.5 s long cathedral tail |
| `plate` | ~2 s dense plate-reverb |
| any path | Load a 16-bit PCM WAV file as the IR |

Synthetic IRs are generated from a deterministic xorshift seed, so the same preset produces the same impulse response every run. Long IRs (> 1024 samples) automatically use FFT-based partitioned convolution; shorter ones use direct time-domain convolution with zero latency.

## Controls

| Key | Action |
|-|-|
| `space` | Play / pause |
| `↑` `↓` `+` `-` | Volume ±5 |
| `r` | Toggle WAV recording (writes `termus-<seed>-<timestamp>.wav` to CWD) |
| `c` | Cycle color theme (indigo / amber / matrix / magenta / mono / rainbow) |
| `C` | Cycle visualization style (scope / spectrum / bars / mirror) |
| `n` / `p` | Next / previous algorithm — hot-swap with ~200 ms crossfade |
| `s` | Skip to next playlist track (only when a playlist is active) |
| `q` `Ctrl-C` | Quit |

## Playlists

Termus can build a playlist that auto-advances tracks with a 2-second
crossfade, so it keeps producing music for hours without intervention. Every
playlist gets a stylized random name derived from the seed
("Velvet Sessions Vol. 7", "Late Atlas").

```bash
# Six different seeds of lo-fi, 5 minutes each
termus --algo lofi --playlist same --playlist-tracks 6 --playlist-duration 5m

# Random genres throughout, 10 tracks
termus --playlist mixed --playlist-tracks 10
```

## SoundFont strategy

By default termus uses one SoundFont for everything (`--sf2-strategy=single`,
balanced 32 MB GeneralUser-GS). For best quality across all genres, opt in
to per-algorithm preferences:

```bash
# Downloads GeneralUser-GS (32 MB) + SGM (325 MB) on first run with a
# progress bar; cycling/playlist switches between them automatically.
termus --sf2-strategy optimal
```

Piano-heavy genres (`lullaby`, `classical`, `lofi`, `jazz`) prefer SGM;
shimmery genres (`ambient`, `drone`, `bells`, `phase`) stay on GeneralUser-GS.

## How it works

Every algorithm pushes audio at 44.1 kHz stereo into a beep `Streamer`. The pure-synthesis algorithms build voices from oscillators, ADSR envelopes, biquad lowpass filters, and delay lines. The SoundFont-based algorithms emit MIDI NoteOn/NoteOff events to a `go-meltysynth` synthesizer.

Long-form variety comes from layered mutation:

1. **Per-slot mutation** — at each note trigger, a small chance to re-roll one of the *other* slots in the cycle so the figure gradually evolves.
2. **Macro key drift** — every 4–7 minutes the key shifts by ±1–2 semitones, taking effect gradually as mutations roll in.
3. **Instrument swaps** — every 3–9 minutes one MIDI channel rotates to a musically-compatible different GM program.
4. **Section toggles** — algorithms with optional ornament layers (e.g. chill's sax + nylon guitar) flip them in/out every 90–240 s, producing verse/chorus-like form.
5. **Per-track humanization** — velocity jitter (±N MIDI velocity) and timing jitter (±N ms) on every note. Lofi swing is a separate deterministic offset.

The TUI is built with bubbletea + lipgloss; the oscilloscope renders to a Braille glyph grid with 2×4 sub-pixel resolution per terminal cell, in your choice of color theme.

## Dependencies

| Library | License | Purpose |
|-|-|-|
| [`gopxl/beep/v2`](https://github.com/gopxl/beep) | MIT | Audio output (CoreAudio / ALSA / WASAPI via `oto`) |
| [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) | MIT | TUI framework |
| [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) | MIT | Terminal styling |
| [`sinshu/go-meltysynth`](https://github.com/sinshu/go-meltysynth) | MIT | Pure-Go SoundFont synthesizer |
| [`madelynnblue/go-dsp`](https://github.com/madelynnblue/go-dsp) | BSD | FFT for partitioned convolution reverb |

The auto-downloaded SoundFont is **GeneralUser-GS.sf2** by S. Christian Collins (~32 MB), MIT-licensed and fetched from [`mrbumpy409/GeneralUser-GS`](https://github.com/mrbumpy409/GeneralUser-GS). It's not bundled with the binary — termus downloads it on first use to your OS cache directory, where you can replace or delete it freely. Verified via SHA-256 after download to reject tampered files.

## License

MIT — see [LICENSE](LICENSE).
