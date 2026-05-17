// internal/audiotest/render.go
//
// Drive the existing audio.Root → Stream pipeline directly into a buffer
// without writing to disk. SF2-requiring algorithms are rejected; use the
// listencheck path with --baseline-* flags for those.
package audiotest

import (
	"fmt"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/scope"
	"github.com/mrbrutti/termus/internal/synth"
)

// RenderAlgorithm renders a non-SF2 algorithm to a stereo float64 buffer.
// Deterministic for a given (algoName, seed, seconds).
func RenderAlgorithm(algoName string, seed int64, seconds float64) ([][2]float64, error) {
	spec, ok := gen.Resolve(algoName)
	if !ok {
		return nil, fmt.Errorf("audiotest: unknown algorithm %q", algoName)
	}
	if spec.RequiresSF2 {
		return nil, fmt.Errorf("audiotest: %q requires SF2 (load through listencheck path)", algoName)
	}
	algo := spec.Build(nil)
	algo.Seed(seed)

	frames := int(seconds * float64(synth.SampleRate))
	if frames <= 0 {
		return nil, fmt.Errorf("audiotest: seconds %g produces no frames", seconds)
	}
	out := make([][2]float64, frames)

	root := audio.NewRoot(algo, scope.NewRing(64))
	root.SetVolume(100)

	const chunk = 4410
	scratch := make([][2]float64, chunk)
	pos := 0
	for pos < frames {
		n := chunk
		if remain := frames - pos; remain < n {
			n = remain
		}
		if _, ok := root.Stream(scratch[:n]); !ok {
			return out[:pos], fmt.Errorf("audiotest: stream ended at frame %d", pos)
		}
		copy(out[pos:pos+n], scratch[:n])
		pos += n
	}
	return out, nil
}
