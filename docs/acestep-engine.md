# ACE-Step engine (SP21)

The ACE-Step engine is the second render path in termus, alongside the
default SF2 procedural engine. It is opt-in per track via the
`render_engine: acestep` field in a v3 `.tm` file.

## Architecture

```
                    +-----------------------------+
                    | termus-stream  (Go binary)  |
                    |                             |
                    |  track.ParseFile(.tm)       |
                    |  acestep.CompileV3()        |
                    |  audio.Streamer (queue+xf)  |
                    +--------------+--------------+
                                   |
                       HTTP/JSON   |   audio/wav
                                   v
                    +-----------------------------+
                    | services/acestep/server.py  |
                    |                             |
                    |  FastAPI                    |
                    |  ACE-Step model (MLX/PyTorch)|
                    |  diffusion -> WAV bytes     |
                    +-----------------------------+
```

The Go side stays Go. The Python side runs as a separate long-lived process
because the ACE-Step model is a Python project with native Apple Silicon
acceleration via MLX (or CPU/CUDA via PyTorch). The HTTP/JSON boundary lets
us keep the Go codebase free of Python dependencies and lets the Python
daemon hold the loaded model warm across many renders.

## Setup

Full instructions live in `services/acestep/README.md`. The short form:

1. `cd services/acestep && ./install.sh` — sets up a Python virtualenv,
   installs FastAPI + ACE-Step + MLX/PyTorch, and downloads the model
   weights (the 2B turbo checkpoint is several GB).
2. `services/acestep/.venv/bin/python services/acestep/server.py` — starts
   the daemon on `http://localhost:7790` by default.
3. `curl http://localhost:7790/health` — should respond
   `{"loaded": true, ...}` after warmup.
4. `termus-stream tracks/lofi/bookstore-rainy-night-v3.tm` — start
   streaming. SIGINT to stop.

## Performance (honest)

The numbers below are taken from the upstream ACE-Step project's
documentation and the model card; they have not been independently
benchmarked in this repository. Measure on your hardware before depending on
them.

- Model load (cold): tens of seconds on M-series Macs; longer on first run
  due to weight download (5-10 GB depending on checkpoint).
- Per-render time (2B turbo, 8 inference steps): roughly 10x faster than
  real-time on Apple M-series silicon and modern NVIDIA GPUs; slower on
  CPU. A 3-minute track is generated in roughly 15-30 seconds on M2/M3.
- Memory: keep the model resident in the daemon; reloading per render is
  prohibitive.

These are not termus-side numbers. The Go streamer adds negligible overhead:
its loop is HTTP I/O + a WAV decode + a beep playback path.

## How streaming works

`internal/audio/streamer.go` runs two goroutines:

1. **Producer loop** — calls `AudioProducer.Produce(ctx, seq)` for
   `seq = 0, 1, 2, ...` and pushes the resulting WAV bytes onto a bounded
   queue. The queue depth (default 2) is the look-ahead window: while
   track N is playing, the producer is generating N+1 (and possibly N+2),
   so playback never has to wait on the model.
2. **Player loop** — pulls the next track off the queue, decodes the WAV
   bytes via `gopxl/beep/v2/wav`, and hands a `beep.Streamer` to the
   `AudioSink` for playback. At the natural end of each track the loop
   crossfades into the next.

### Look-ahead math

With queue depth `Q`, per-render time `T_r`, and per-track length `T_l`:

- The producer stays ahead of the player as long as `T_r < T_l`. For a
  3-minute track that renders in 20 seconds, the producer finishes 9 tracks
  in the time the player consumes 1, so the queue is always full and
  playback never stalls.
- The first track is paid for in latency. After
  `T_r` seconds the first track plays; thereafter playback is continuous.
- `Q = 2` is enough headroom for ACE-Step's turbo settings. Increase `Q` if
  you need more cushion for high-inference-step renders.

### Crossfade

The streamer uses `effects.TransitionEqualPower` from `gopxl/beep`: a
cosine ramp paired with its inverse, mixed together by a `beep.Mixer`.
Total perceived volume stays constant across the overlap window. Default
`CrossfadeSec = 3`.

The last track in a `--max-tracks N` run gets a tail fade-out instead of a
crossfade so playback ends cleanly.

## Authoring a v3 .tm

A minimal v3 track looks like:

```yaml
title: My ACE-Step Track
render_engine: acestep
key: Cmin
tempo: 88
total_duration: 3m

acestep:
  style: >
    warm lo-fi rhodes with vinyl crackle in a quiet bookstore.
    No vocals, no lead percussion.
  tags: [lofi, rhodes, instrumental]
  motif: stepwise descent from the fifth to the root
  sections:
    - id: intro
      bars: 8
      description: soft Rhodes, no drums
      harmony: "Cm9"
      dynamic: soft
    - id: head
      bars: 16
      description: brushed kick and bass enter
      harmony: "Cm9 Fm7 Abmaj7 G7sus"
      dynamic: building
```

Fields:

- `render_engine: acestep` — required, selects the engine.
- `acestep.style` — natural-language paragraph; becomes the bulk of the
  ACE-Step prompt.
- `acestep.tags` — rank-ordered; the first tag should be the genre.
- `acestep.sections[].description` — per-section guidance; each is folded
  into the prompt and surfaced to the model as a separate
  `section_descriptions[]` entry.
- `acestep.sections[].harmony` — chord symbols, concatenated across all
  sections into a single `harmony_chain`.
- `acestep.motif` — natural-language motif description.
- `acestep.scale` — optional override; otherwise inferred from `key`.
- `acestep.time_signature` — defaults to `4/4`.
- `acestep.seed` — optional override of the top-level `seed`.
- `acestep.inference_steps` — diffusion step count override; 0 uses the
  turbo default (8). Higher = slower + (typically) cleaner.

`tracks/lofi/bookstore-rainy-night-v3.tm` is the reference fixture and is
parsed by the prompt-compiler test suite.

## Known limitations

- **Model download size.** The ACE-Step 2B turbo checkpoint is several GB
  and is downloaded on first install. There is no smaller drop-in
  replacement supported by this engine.
- **First-render latency.** Even after the model is warm, the first
  render after process start incurs an additional 10-30s of JIT/compile
  cost on some backends. The streamer's `--queue-depth 2` largely hides
  this from the listener after track 0.
- **Python dependency.** termus itself stays pure Go, but the engine
  requires a separately-installed Python virtualenv. The install script
  is documented but it is the largest single piece of state in this
  feature.
- **Mocked tests only.** The Go test suite uses `httptest` mocks of the
  service and an in-memory recording sink for the streamer. End-to-end
  audio generation requires the real Python daemon — the user must
  validate audio quality and timing on their hardware.
- **Fixed sample rate per session.** The streamer initialises the OS
  speaker with the first track's sample rate and does not resample. In
  practice ACE-Step outputs a fixed rate per model checkpoint, so this
  doesn't bite, but it's a hard fail if a track ever returns a different
  rate.
- **No restart of the streamer.** Once `Stop()` returns, the `Streamer`
  cannot be reused. Build a new one. This is consistent with the
  goroutine-life-cycle pattern used elsewhere in termus.
