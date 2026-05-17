# Audio Test Harness

The `internal/audiotest/` package provides signal-level DSP assertions
(RMS, peak, spectral centroid, transient detection, pitch-modulation
depth/rate, convolution-response verification) and an in-memory renderer
(`RenderAlgorithm`) so audio changes can be regression-tested without
listening checkpoints.

## Capturing a baseline

Run:

    go run ./cmd/termus-listencheck --baseline-capture testdata/listencheck/baseline.json

This renders the default non-SF2 corpus and writes RMS / peak / centroid
per entry. Commit the resulting `baseline.json` whenever an intentional
audio change is reviewed and approved.

## Checking against the committed baseline

Run:

    go run ./cmd/termus-listencheck --baseline-check testdata/listencheck/baseline.json

Exits 0 on a clean check, 1 if any entry drifts beyond:

- RMS:      ±1.0 dB
- Peak:     ±1.5 dB
- Centroid: ±10%

The same comparison runs as a unit test
(`TestCommittedBaselineMatchesCurrentRender` in
`cmd/termus-listencheck/baseline_test.go`) so `go test ./...` catches
unintentional drift in CI.

## Writing new assertions in a sub-plan

When SP1 (wow/flutter), SP2 (IR library), SP4 (personality layer), or
SP5 (mix bus) lands, add unit tests that exercise the new processor
in isolation using `audiotest`:

    package mypkg_test

    import (
        "testing"
        "github.com/mrbrutti/termus/internal/audiotest"
    )

    func TestWowFlutterProducesExpectedPitchModulation(t *testing.T) {
        in := audiotest.Sine(440, 1.0, 44100, 8*44100)
        out := myWowFlutter(in)  // SP1 processor
        audiotest.AssertPitchModulationCents(t, out, 44100, 440, 15, 0.7, 1.0, 0.05)
    }

## Limitations

- SF2-backed algorithms (lofi, jazz) are not yet baselined. A follow-up
  will add `RenderAlgorithmWithSF2(name, seed, seconds, *meltysynth.SoundFont)`
  and extend the corpus.
- `--baseline-capture` re-renders each corpus item after the WAV
  rendering loop (one render to disk, one in-memory render to measure).
  Slightly wasteful but harmless; both paths are deterministic.
- `--baseline-check` silently ignores corpus items absent from the
  baseline JSON. A newly added algorithm needs its baseline captured
  before drift detection becomes meaningful for it.
- The pitch tracker assumes monophonic, sustained input. For wow/flutter
  verification, feed a pure sine through the processor; don't try to
  measure pitch modulation on a full track render.
