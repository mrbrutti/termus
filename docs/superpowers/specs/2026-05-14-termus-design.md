# termus — design

A small terminal music player that algorithmically generates ambient deep-work / coding music and displays an oscilloscope visualization of its own output. v1 ships one generative algorithm (Eno-style loop drift) behind a pluggable interface; future versions add Pentatonic random walk, Markov-chain melody, and pure drone.

## Goals

- Single-binary Go CLI, zero external runtime dependencies (no SoX, no ffmpeg).
- Pure-Go audio synthesis — every sample is computed in-process, so the visualizer always reflects what the speakers play.
- Polished, "deep-work-vibe" terminal UI: a colored Braille oscilloscope with a soft CRT afterglow, framed by a minimal lipgloss chrome.
- Pluggable algorithm interface so v2+ can add generators without restructuring the audio or UI layers.
- Deterministic from a seed: `termus --seed 42` plays the exact same music every time.
- Optional WAV recording of the current session.

## Non-goals (v1)

- MIDI in/out, OSC, external synth control.
- Real-time effects routing exposed to the user (the algorithm wires its own effects internally).
- Music theory beyond scales/modes (no chord progressions, voice leading, or harmony engine).
- Mouse support; touch UI; remote control.
- Tracks/library/playlists — termus produces an endless stream, not files of finite length.

## Stack

- **Language:** Go (workspace at `/Users/matt/Code/go/MrBrutti/termus`).
- **Audio:** [`github.com/gopxl/beep/v2`](https://github.com/gopxl/beep), built on `hajimehoshi/oto` (CoreAudio on macOS, ALSA/PulseAudio on Linux, WASAPI on Windows). beep's `Streamer` interface and built-in `Mixer` map directly to our multi-sink architecture; its `wav` package handles recording.
- **TUI:** [`github.com/charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) for the Elm-style update loop, [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) for styling.
- **Sample rate / format:** 44100 Hz, stereo, internal processing in `float64`, converted to int16 at the speaker boundary.

## Module layout

```
termus/
├── cmd/termus/main.go         flags, seed, wire-up
├── internal/synth/            DSP primitives: oscillator, ADSR, biquad LP, delay line
├── internal/gen/              Algorithm interface + eno implementation
├── internal/audio/            beep glue: root streamer, mixer, wav sink, scope tap
├── internal/scope/            lock-free ring buffer (audio writes, UI reads)
└── internal/tui/              bubbletea model + Braille oscilloscope view
```

Each package has one responsibility and a narrow public surface:

- `synth` knows nothing about `gen`, `audio`, or `tui`. It is pure DSP math.
- `gen` depends on `synth`. It knows nothing about beep or the UI.
- `audio` depends on `gen` and `scope`. It knows nothing about the UI.
- `tui` depends on `scope` and on a small command interface published by `audio` (volume / play-pause / record). It does not import `gen` or `synth`.

The `audio` package publishes this interface for the UI:

```go
type Commander interface {
    SetVolume(pct int)                  // 0..100; clamped
    TogglePause()
    ToggleRecord() (path string, err error) // path set when starting; err if WAV open fails
}
```

## Core abstractions

### `gen.Algorithm`

```go
type Algorithm interface {
    Name() string
    Seed(s int64)                          // deterministic (re)initialization
    Next(left, right []float64)            // fill stereo buffer; len(left)==len(right)
}
```

v1 ships exactly one implementation, `gen.Eno`. The `--algo` flag selects an algorithm by name (only `eno` accepted in v1). The contract: `Next` is called repeatedly by the audio thread and must be wait-free (no locks, no allocations on the hot path after construction).

### `audio.Root`

A `beep.Streamer` that owns:
- the active `gen.Algorithm`
- a master gain (read from an `atomic.Uint32` published by the UI)
- a beep `Mixer` driving the speaker
- an optional WAV sink, started/stopped via a non-blocking command channel
- the `scope.Ring` write side

Its `Stream(samples [][2]float64) (n int, ok bool)` method is the single entry point the audio goroutine uses.

### `scope.Ring`

A lock-free ring buffer of mono samples (the average of L/R after master gain). Single-writer (audio goroutine), single-reader-at-a-time (UI goroutine). Capacity ~4096 samples, sized to fit the widest terminal we care about at our render downsample ratio. The UI calls `Snapshot(dst []float64)` which copies out the most recent `len(dst)` samples using an atomically-loaded write index.

## Data flow

```
beep speaker  ──► audio.Root.Stream(buf)
                       │
                       ▼
                gen.Algorithm.Next(L, R)
                       │
                       ▼
              apply master gain (atomic)
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
    speaker      scope.Ring     WAV writer
                       │           (if recording)
                       ▼
            bubbletea View tick (30 FPS)
                       │
                       ▼
            Braille oscilloscope render
```

The audio goroutine runs the synthesis at whatever block size beep negotiates with oto (~512–1024 frames, ~12–23 ms latency). The UI ticks at 30 FPS independently; it never blocks the audio thread.

## Concurrency rules

- Audio goroutine never takes a lock and never allocates on the hot path.
- UI → audio commands flow through:
  - `atomic.Uint32` for master gain (volume): UI writes the 0..100 integer; audio reads it once per block and maps it to a linear gain `v/100.0` (we may revisit with a perceptual curve later, but linear is fine for v1).
  - A non-blocking `chan command` for play/pause and start/stop-recording — audio drains the channel at the top of each `Stream` call. If the channel is full (UI spam), commands are dropped; the UI's own state is authoritative for display.
- Audio → UI: only the `scope.Ring`. The UI never reads anything else from audio state directly.

## Eno-drift algorithm (v1)

The "Music for Airports" trick: a small set of short tonal phrases, each on its own loop period, where the periods are mutually incommensurate so the phrases never realign and the texture never repeats.

Specifics for v1:

- **Key:** seed-derived pick from `{C, D, E, F, G, A, B}` minor (minor only in v1 — the brief is "deep work").
- **Scale:** natural minor.
- **Voices:** 5 melodic voices. Each voice is a 1–3 note phrase drawn from the scale at construction time. Voice `i` has loop period from the set `{7.0, 11.0, 13.3, 17.7, 23.1}` seconds — chosen because pairwise ratios are irrational-ish, so realignment is on the order of hours.
- **Note timbre:** `0.7 * sine(f) + 0.15 * saw(f * 1.005) + 0.15 * saw(f * 0.995)` (a sine with two slightly-detuned saw siblings for warmth), through an ADSR envelope (attack 1.8 s, decay 0.5 s, sustain 0.6, release 3.5 s) and a soft biquad lowpass at 2 kHz.
- **Drone pad:** two saw oscillators at root and root+octave, each gently detuned (±5 cents), summed at -18 dB, lowpassed with a slow LFO on the cutoff (0.05 Hz, sweeping 200–800 Hz).
- **Space:** stereo cross-delay (~300 ms / ~420 ms taps, ~25% feedback, ~30% wet). No convolution reverb in v1; the cross-delay is much cheaper and gives most of the spatial effect for ambient material.
- **Master:** soft tanh saturation on the sum to prevent clipping if too many voices coincide.

Future algorithms (`pentatonic`, `markov`, `drone`) reuse `synth/` primitives but compose them differently; they implement the same `gen.Algorithm` interface.

## TUI

Bubbletea model:

```go
type Model struct {
    width, height int
    scope         *scope.Ring
    audio         AudioCommander  // narrow interface: SetVolume, TogglePause, ToggleRecord
    volume        int             // 0..100, mirrors atomic in audio
    paused        bool
    recording     bool
    recordPath    string          // shown briefly after starting
    algoName      string          // e.g. "eno-drift"
    key           string          // e.g. "Cmin"
    seed          int64
    frame         []float64       // reusable buffer for Snapshot
    afterglow     [][]rune        // last 3 rendered frames for ghost trails
}
```

Key bindings:

| Key       | Action                |
| --------- | --------------------- |
| `space`   | toggle play/pause     |
| `↑` / `↓` | volume +5 / -5        |
| `+` / `-` | volume +5 / -5 (alt)  |
| `r`       | toggle WAV recording (file: `termus-<seed>-<unix-timestamp>.wav` in CWD) |
| `q` / `^C`| quit                  |

Render:

- **Top bar (1 row):** `termus · eno-drift · Cmin · seed=42` (left), `● REC` (right, if recording).
- **Scope (most of the screen):** Braille oscilloscope. The available rows minus chrome are the y-resolution; the width is the x-resolution. Each terminal cell is a Braille glyph with 2×4 sub-pixel dots → curves render smoothly. Coloring uses a vertical gradient (deep indigo at the rails, cyan near the centerline, near-white at amplitude peaks) plus an afterglow: the previous 2–3 frames are kept and drawn behind the current one at fading alpha so the trace appears to leave a CRT phosphor trail.
- **Bottom bar (1 row):** `[space] pause   [↑↓] vol 70%   [r] rec   [q] quit`.

Below ~40 cols or ~10 rows: replace the scope with a centered "terminal too small — resize to 40×10 minimum" message. Audio keeps playing.

## CLI surface

```
termus [--seed N] [--algo eno] [--volume 0..100]

  --seed     int      seed for the generator (default: time-based, printed at startup)
  --algo     string   algorithm name (default: eno; v1 only accepts "eno")
  --volume   int      initial output volume 0..100 (default: 70)
```

Exit codes: `0` on clean quit, `1` on audio init failure, `2` on bad flags.

## Error handling

- **Audio device unavailable** (no output, permission denied): print a one-line stderr message naming the underlying error, exit 1. The TUI never starts.
- **WAV writer failure mid-session** (disk full, perms revoked, path no longer writable): stop the recording cleanly, surface `failed to write recording: <err>` in the bottom status bar for ~3 seconds, keep playing audio.
- **Terminal too small at startup:** still launch, show the resize message, keep audio playing.
- **Panic in TUI:** deferred terminal restore in `cmd/termus/main.go`; we don't try to handle audio panics, we let the process die since they indicate a real bug.
- **Ctrl-C / `q`:** stop the speaker, flush and close any open WAV file, restore terminal.

## Testing strategy

- `internal/synth/`: unit tests with deterministic sample math. Examples: a 440 Hz sine over 1 second produces 440 zero-crossings ±tolerance; ADSR envelope sampled at `t=0, t=A, t=A+D, t=A+D+S, t=A+D+S+R` matches the expected curve; biquad LP attenuates 20 kHz by > 40 dB.
- `internal/gen/`: determinism test — `Seed(N); Next(...) → []byte` is byte-identical across runs and across two algorithm instances seeded the same way.
- `internal/scope/`: ring buffer correctness test with `go test -race`, hammering it from two goroutines.
- `internal/audio/`: an integration test that runs the root streamer for 1 second into an in-memory sink and asserts RMS > some threshold (i.e. actual audio was produced) and that the scope ring received samples.
- `internal/tui/`: golden-file tests for the Braille renderer — given a canned sample buffer and a terminal size, the produced frame must match a checked-in `.golden` file. Renderer-only; the bubbletea model is exercised manually.
- The overall feel — does the music sound good, does the scope look nice — is verified by running `termus` and listening / looking. The spec acknowledges this is not test-covered.

## Open questions / deferred

- Real reverb (convolution or Schroeder/FDN) is deferred to v2; cross-delay is enough for v1.
- Major keys / mode selection is deferred; v1 is minor-only.
- Per-algorithm config flags (`--bpm`, `--density`) are deferred until we have ≥2 algorithms and can see what they should share.
- Linux/Windows builds are expected to work via beep/oto's cross-platform support but are not validated as part of v1 acceptance — primary target is macOS (Darwin 25.4).
