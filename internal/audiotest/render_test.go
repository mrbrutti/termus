// internal/audiotest/render_test.go
package audiotest

import (
	"testing"

	"github.com/mrbrutti/termus/internal/synth"
)

func TestRenderAlgorithmProducesNonSilentBuffer(t *testing.T) {
	const seconds = 0.5
	buf, err := RenderAlgorithm("ambient-synth", 42, seconds)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	wantFrames := int(seconds * float64(synth.SampleRate))
	if len(buf) != wantFrames {
		t.Fatalf("len(buf) = %d, want %d", len(buf), wantFrames)
	}
	mono := ToMono(buf)
	if RMS(mono) < 1e-6 {
		t.Fatalf("RMS = %g, buffer appears silent", RMS(mono))
	}
}

func TestRenderAlgorithmIsDeterministicForSameSeed(t *testing.T) {
	a, err := RenderAlgorithm("ambient-synth", 7, 0.25)
	if err != nil {
		t.Fatalf("render a: %v", err)
	}
	b, err := RenderAlgorithm("ambient-synth", 7, 0.25)
	if err != nil {
		t.Fatalf("render b: %v", err)
	}
	if len(a) != len(b) {
		t.Fatalf("length differs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("differ at frame %d: %v vs %v", i, a[i], b[i])
		}
	}
}

func TestRenderAlgorithmReturnsErrorForUnknownName(t *testing.T) {
	if _, err := RenderAlgorithm("not-a-real-algo", 1, 0.1); err == nil {
		t.Fatal("expected error for unknown algorithm name")
	}
}

func TestRenderAlgorithmRejectsSF2AlgorithmWithoutSoundFont(t *testing.T) {
	// "lofi" requires SF2; should error rather than crash.
	if _, err := RenderAlgorithm("lofi", 1, 0.1); err == nil {
		t.Fatal("expected error when rendering SF2 algorithm without sound font")
	}
}
