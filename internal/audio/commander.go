// Package audio glues the synth output to beep, the scope ring buffer, and an
// optional WAV recording sink.
package audio

// Commander is the narrow interface the TUI uses to control audio.
type Commander interface {
	SetVolume(pct int)
	TogglePause()
	// ToggleRecord starts or stops recording. When starting, returns the
	// output path. When stopping (or on failure), path is "".
	ToggleRecord() (path string, err error)
}
