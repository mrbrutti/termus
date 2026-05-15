// Package audio glues the synth output to beep, the scope ring buffer, and an
// optional WAV recording sink.
package audio

import "github.com/mrbrutti/termus/internal/gen"

// Commander is the narrow interface the TUI uses to control audio.
type Commander interface {
	SetVolume(pct int)
	TogglePause()
	// ToggleRecord starts or stops recording. When starting, returns the
	// output path. When stopping (or on failure), path is "".
	ToggleRecord() (path string, err error)
	// SwapAlgorithm hot-swaps the running algorithm. Picked up by the audio
	// thread at the start of the next Stream call.
	SwapAlgorithm(algo gen.Algorithm)
}
