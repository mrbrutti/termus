package gen

import (
	"math/rand"
	"testing"

	"github.com/mrbrutti/termus/internal/synth"
)

type sf2Event struct {
	kind string
	key  int32
	cc   int32
	val  int32
}

type fakeSF2Sink struct {
	events []sf2Event
}

func (f *fakeSF2Sink) NoteOn(channel int32, key int32, velocity int32) {
	f.events = append(f.events, sf2Event{kind: "on", key: key})
}

func (f *fakeSF2Sink) NoteOff(channel int32, key int32) {
	f.events = append(f.events, sf2Event{kind: "off", key: key})
}

func (f *fakeSF2Sink) ProcessMidiMessage(channel int32, command int32, data1 int32, data2 int32) {
	f.events = append(f.events, sf2Event{kind: "midi", cc: command, key: data1, val: data2})
}

func testTrackState(cfg SF2Track, periodSamples int64) *sf2TrackState {
	return &sf2TrackState{
		cfg:           cfg,
		periodSamples: periodSamples,
		notesLen:      int64(len(cfg.Notes)),
		curSlot:       -1,
		curKey:        -1,
		overlapKey:    -1,
	}
}

func TestSF2TrackStateTieRepeatsKeepsSingleNoteOn(t *testing.T) {
	sink := &fakeSF2Sink{}
	state := testTrackState(SF2Track{
		Channel:     0,
		Velocity:    90,
		Notes:       []int{60, 60},
		PeriodSec:   2,
		Gate:        1.0,
		Legato:      true,
		TieRepeats:  true,
		OverlapSec:  0.02,
		ReleaseSec:  0,
		ResolveNote: nil,
	}, 200)

	state.fireTransition(0, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	if len(sink.events) != 1 || sink.events[0] != (sf2Event{kind: "on", key: 60}) {
		t.Fatalf("first fire events = %+v", sink.events)
	}
	firstRelease := state.releaseT

	state.fireTransition(100, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	if len(sink.events) != 1 {
		t.Fatalf("tied repeat re-articulated: %+v", sink.events)
	}
	if state.curKey != 60 {
		t.Fatalf("curKey = %d, want 60", state.curKey)
	}
	if state.releaseT <= firstRelease {
		t.Fatalf("releaseT = %d, want extension beyond %d", state.releaseT, firstRelease)
	}
}

func TestSF2TrackStateOverlapDelaysPriorNoteOff(t *testing.T) {
	sink := &fakeSF2Sink{}
	state := testTrackState(SF2Track{
		Channel:    0,
		Velocity:   90,
		Notes:      []int{60, 62},
		PeriodSec:  2,
		Gate:       1.0,
		Legato:     true,
		OverlapSec: 0.01,
	}, 2*int64(synth.SampleRate))

	state.fireTransition(0, sink, rand.New(rand.NewSource(1)))                       //nolint:gosec
	state.fireTransition(int64(synth.SampleRate), sink, rand.New(rand.NewSource(1))) //nolint:gosec
	if got, want := len(sink.events), 2; got != want {
		t.Fatalf("event count after overlap fire = %d, want %d (%+v)", got, want, sink.events)
	}
	if sink.events[0] != (sf2Event{kind: "on", key: 60}) || sink.events[1] != (sf2Event{kind: "on", key: 62}) {
		t.Fatalf("unexpected overlap events: %+v", sink.events)
	}
	if state.overlapKey != 60 || state.overlapOffT <= int64(synth.SampleRate) {
		t.Fatalf("overlap state = key %d off %d", state.overlapKey, state.overlapOffT)
	}

	state.handleDueEvents(state.overlapOffT-1, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	if len(sink.events) != 2 {
		t.Fatalf("overlap note ended too early: %+v", sink.events)
	}
	state.handleDueEvents(state.overlapOffT, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	if got, want := sink.events[len(sink.events)-1], (sf2Event{kind: "off", key: 60}); got != want {
		t.Fatalf("final overlap event = %+v, want %+v", got, want)
	}
}

func TestSF2TrackStateRestDoesNotForceHeldNoteOff(t *testing.T) {
	sink := &fakeSF2Sink{}
	state := testTrackState(SF2Track{
		Channel:   0,
		Velocity:  90,
		Notes:     []int{60, -1},
		PeriodSec: 2,
		Gate:      1.5,
		Legato:    true,
	}, 200)

	state.fireTransition(0, sink, rand.New(rand.NewSource(1)))   //nolint:gosec
	state.fireTransition(100, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	if len(sink.events) != 1 {
		t.Fatalf("rest slot forced note off: %+v", sink.events)
	}
	if state.curKey != 60 {
		t.Fatalf("curKey = %d, want 60 held through rest", state.curKey)
	}

	state.handleDueEvents(150, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	foundOff := false
	for _, ev := range sink.events {
		if ev == (sf2Event{kind: "off", key: 60}) {
			foundOff = true
			break
		}
	}
	if !foundOff {
		t.Fatalf("release after held rest missing note-off: %+v", sink.events)
	}
}

func TestSF2TrackStateDeterministicTimingOffsetShapesNextFire(t *testing.T) {
	sink := &fakeSF2Sink{}
	state := testTrackState(SF2Track{
		Channel:                0,
		Velocity:               90,
		Notes:                  []int{60, 62},
		PeriodSec:              2,
		Gate:                   1.0,
		ResolveTimingOffsetSec: cyclicTimingOffset(5),
	}, 2*int64(synth.SampleRate))

	state.fireTransition(0, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	want := int64(synth.SampleRate) + secondsToSamples(0.005)
	if state.nextFireT != want {
		t.Fatalf("nextFireT = %d, want %d", state.nextFireT, want)
	}
}

func TestSF2TrackStatePerformanceCurvesEmitControlChanges(t *testing.T) {
	sink := &fakeSF2Sink{}
	state := testTrackState(SF2Track{
		Channel:   0,
		Velocity:  90,
		Notes:     []int{60},
		PeriodSec: 1,
		Gate:      1.0,
		ResolveExpression: func(slot int, key int) SF2ExpressionCurve {
			return SF2ExpressionCurve{Start: 80, Peak: 100, End: 84, PeakAt01: 0.25}
		},
		ResolveModWheel: func(slot int, key int) SF2ExpressionCurve {
			return SF2ExpressionCurve{Start: 0, Peak: 18, End: 6, PeakAt01: 0.5}
		},
		ResolveBrightness: func(slot int, key int) SF2ExpressionCurve {
			return SF2ExpressionCurve{Start: 90, Peak: 110, End: 94, PeakAt01: 0.2}
		},
	}, 100)

	state.fireTransition(0, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	if len(sink.events) < 4 {
		t.Fatalf("initial events = %+v", sink.events)
	}

	state.handleDueEvents(25, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	state.handleDueEvents(50, sink, rand.New(rand.NewSource(1))) //nolint:gosec
	state.handleDueEvents(99, sink, rand.New(rand.NewSource(1))) //nolint:gosec

	foundExpr := false
	foundMod := false
	foundBright := false
	for _, ev := range sink.events {
		if ev.kind != "midi" {
			continue
		}
		if ev.cc == ccControlChange && ev.key == 11 && ev.val == 100 {
			foundExpr = true
		}
		if ev.cc == ccControlChange && ev.key == 1 && ev.val == 18 {
			foundMod = true
		}
		if ev.cc == ccControlChange && ev.key == 74 && ev.val == 110 {
			foundBright = true
		}
	}
	if !foundExpr || !foundMod || !foundBright {
		t.Fatalf("missing control peaks: expr=%v mod=%v bright=%v events=%+v", foundExpr, foundMod, foundBright, sink.events)
	}
}

func TestSF2TrackStateDetuneSendsPitchBendAndReset(t *testing.T) {
	sink := &fakeSF2Sink{}
	state := testTrackState(SF2Track{
		Channel:   0,
		Velocity:  90,
		Notes:     []int{60},
		PeriodSec: 1,
		Gate:      1.0,
		ResolveDetuneCents: func(slot int, key int) int32 {
			return 20
		},
	}, 100)

	state.fireTransition(0, sink, rand.New(rand.NewSource(1)))    //nolint:gosec
	state.handleDueEvents(100, sink, rand.New(rand.NewSource(1))) //nolint:gosec

	foundDetune := false
	foundReset := false
	for _, ev := range sink.events {
		if ev.kind != "midi" || ev.cc != 0xE0 {
			continue
		}
		if ev.key == pitchBendLSB(20) && ev.val == pitchBendMSB(20) {
			foundDetune = true
		}
		if ev.key == pitchBendLSB(0) && ev.val == pitchBendMSB(0) {
			foundReset = true
		}
	}
	if !foundDetune || !foundReset {
		t.Fatalf("pitch bend events missing: detune=%v reset=%v events=%+v", foundDetune, foundReset, sink.events)
	}
}
