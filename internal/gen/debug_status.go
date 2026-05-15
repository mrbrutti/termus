package gen

import "fmt"

// DebugStatus is a lightweight, UI-safe snapshot of the algorithm's current
// musical state.
type DebugStatus struct {
	Chord   string
	Section string
	Bar     int
	Preset  string
}

// DebugStatusProvider lets algorithms expose a lock-free status snapshot to
// the audio thread, which then republishes it to the TUI.
type DebugStatusProvider interface {
	DebugStatus() DebugStatus
}

// WrapDebugStatus attaches a preset/source label to an algorithm while
// preserving any algorithm-provided chord/section/bar status.
func WrapDebugStatus(algo Algorithm, preset string) Algorithm {
	if algo == nil {
		return nil
	}
	return debugWrappedAlgorithm{Algorithm: algo, preset: preset}
}

// SnapshotDebugStatus resolves the current status for an algorithm, returning
// an empty struct when the algorithm does not expose any debug information.
func SnapshotDebugStatus(algo Algorithm) DebugStatus {
	if provider, ok := algo.(DebugStatusProvider); ok {
		return provider.DebugStatus()
	}
	return DebugStatus{}
}

func FormatDebugStatus(status DebugStatus) string {
	parts := make([]string, 0, 4)
	if status.Bar > 0 {
		parts = append(parts, fmt.Sprintf("bar %d", status.Bar))
	}
	if status.Section != "" {
		parts = append(parts, status.Section)
	}
	if status.Chord != "" {
		parts = append(parts, status.Chord)
	}
	if status.Preset != "" {
		if status.Preset == "synth" {
			parts = append(parts, "synth")
		} else {
			parts = append(parts, "sf2 "+status.Preset)
		}
	}
	return joinDebugParts(parts)
}

func joinDebugParts(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	out := parts[0]
	for _, part := range parts[1:] {
		out += " · " + part
	}
	return out
}

func chordOffsetLabel(offset int) string {
	switch ((offset % 12) + 12) % 12 {
	case 0:
		return "I"
	case 2:
		return "ii"
	case 3:
		return "bIII"
	case 4:
		return "III"
	case 5:
		return "IV"
	case 7:
		return "V"
	case 9:
		return "vi"
	case 10:
		return "bVII"
	default:
		return fmt.Sprintf("%+d", offset)
	}
}

type debugWrappedAlgorithm struct {
	Algorithm
	preset string
}

func (a debugWrappedAlgorithm) DebugStatus() DebugStatus {
	status := SnapshotDebugStatus(a.Algorithm)
	if status.Preset == "" {
		status.Preset = a.preset
	}
	return status
}
