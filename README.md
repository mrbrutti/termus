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

First run with a SoundFont-based algorithm (`--algo sf2`, `eno-sf2`, etc.) auto-downloads a SoundFont into `~/Library/Caches/termus/soundfonts/` (or the OS equivalent). Two curated presets:

| Preset | Size | What it is |
|-|-|-|
| `general` (default) | 32 MB | **GeneralUser-GS** by S. Christian Collins. MIT. 261 instruments, 13 drum kits. Balanced. |
| `sgm` | 325 MB | **SGM v2.01 NicePianosGuitarsBass** by Sergio Acuña. Substantially better piano/guitar/bass — best for `chill`, `pentatonic-sf2`, `markov-sf2`. ~10× the download. |

```bash
termus --algo chill              # default: GeneralUser-GS (32 MB)
termus --algo chill --sf2-preset sgm   # audiophile: SGM (325 MB on first run)
termus --algo chill --sf2 ~/Music/MyFavorite.sf2   # your own file
```

## Algorithms

Termus ships nine algorithms. Six are pure-synthesis (~3 MB binary, no downloads); the others run through a SoundFont sampled-instrument engine. All produce indefinitely-long output via per-note mutation, macro key drift, instrument swaps, and section toggling.

| Algorithm | What it is |
|-|-|
| `eno` | Brian Eno "Music for Airports" — pad-bell voices on incommensurate loop periods over a slow drone, pure synthesis |
| `drone` | Stars of the Lid / pure drone — multiple long-sustained voices with a shimmer voice on top |
| `glass` | FM-synthesis bells with subtle vibrato — bright, crystalline, late-night focus music |
| `pentatonic` | Random walk through pentatonic minor — never clashes, perpetually melodic |
| `markov` | Hand-crafted minor-key transition matrix — feels "composed" rather than random |
| `sf2` | Hi-fi chord-progression algorithm — piano arpeggio + strings + warm pad + bass + flute melody over an 8-chord A-B form with modal interchange |
| `eno-sf2` / `drone-sf2` / `glass-sf2` / `pentatonic-sf2` / `markov-sf2` | Same scheduling logic as the pure-synth versions, but each voice plays a real sampled instrument with stereo placement, reverb sends, filter LFOs, and macro instrument swaps over time |
| `phase` | Reich-style phase-shift — two vibraphone voices play the same 4-note pentatonic figure at slightly different tempos, drifting continuously in and out of rhythmic alignment |
| `chill` | Lofi hip-hop — drum kit + Rhodes EP chord stabs + walking bass + vibe melody + sparse sax + nylon guitar comping, with swing, tape saturation, vinyl crackle, sidechain ducking, and the typical "muffled tape" low-pass |

## Usage

```bash
termus [--algo NAME] [--seed N] [--volume 0..100]
       [--sf2 PATH] [--ir NAME-OR-PATH] [--ir-wet 0..1]
```

Examples:

```bash
# Default — pure-synthesis Eno
termus

# Pick an algorithm + seed (same seed = same music)
termus --algo chill --seed 42

# Use the SoundFont chord-progression algorithm in a cathedral
termus --algo sf2 --ir cathedral

# Bring your own impulse response
termus --algo eno-sf2 --ir ~/Downloads/concert-hall.wav --ir-wet 0.4

# Start louder
termus --algo phase --volume 90
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
| `c` | Cycle oscilloscope color theme (indigo / amber / matrix / magenta / mono / rainbow) |
| `q` `Ctrl-C` | Quit |

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
