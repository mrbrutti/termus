package gen

import "testing"

type debugStatusStub struct {
	name   string
	status DebugStatus
}

func (a debugStatusStub) Name() string { return a.name }
func (a debugStatusStub) Seed(int64)   {}
func (a debugStatusStub) Next(left, right []float64) {
}
func (a debugStatusStub) DebugStatus() DebugStatus { return a.status }

func TestWrapDebugStatusAddsPreset(t *testing.T) {
	wrapped := WrapDebugStatus(debugStatusStub{
		name:   "stub",
		status: DebugStatus{Chord: "Dm7", Bar: 3},
	}, "tyros4")
	status := SnapshotDebugStatus(wrapped)
	if status.Preset != "tyros4" || status.Chord != "Dm7" || status.Bar != 3 {
		t.Fatalf("wrapped status = %+v", status)
	}
}

func TestFormatDebugStatus(t *testing.T) {
	got := FormatDebugStatus(DebugStatus{
		Bar:     5,
		Section: "A'",
		Chord:   "G7",
		Preset:  "sgm",
	})
	want := "bar 5 · A' · G7 · sf2 sgm"
	if got != want {
		t.Fatalf("FormatDebugStatus = %q, want %q", got, want)
	}
}
