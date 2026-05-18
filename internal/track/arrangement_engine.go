package track

import (
	"strings"

	"github.com/mrbrutti/termus/internal/gen"
)

// ArrangementBeatWindow describes the time window during which a role is
// active inside a section. enterBeat and exitBeat are 1-indexed beats inside
// the section (beat 1.0 = start of bar 1). fadeIn and fadeOut are durations
// in beats. When a role has no schedule, the window is "the whole section"
// (no gating).
type ArrangementBeatWindow struct {
	EnterBeat float64
	ExitBeat  float64 // exclusive end-beat; events at >= this beat are gated
	FadeIn    float64
	FadeOut   float64
	Prominent bool
	// HasSchedule is true when the section had an explicit entry for this role.
	HasSchedule bool
}

// computeArrangementWindows turns a Section.Arrangement18 (role → schedule)
// into per-role beat windows. totalBeats is the total beat count of the
// section. bpm is provided so we can convert bars → beats (4 beats/bar in 4/4).
func computeArrangementWindows(section Section, totalBeats float64) map[string]ArrangementBeatWindow {
	if len(section.Arrangement18) == 0 {
		return nil
	}
	const beatsPerBar = 4.0
	totalBars := int(totalBeats / beatsPerBar)
	if totalBars <= 0 {
		totalBars = 1
	}
	out := make(map[string]ArrangementBeatWindow, len(section.Arrangement18))
	for name, sched := range section.Arrangement18 {
		win := ArrangementBeatWindow{
			EnterBeat:   1.0,
			ExitBeat:    totalBeats + 1.0, // beats are 1-indexed
			FadeIn:      float64(sched.FadeInBars) * beatsPerBar,
			FadeOut:     float64(sched.FadeOutBars) * beatsPerBar,
			Prominent:   sched.Prominent,
			HasSchedule: true,
		}
		if sched.EnterBar > 1 {
			win.EnterBeat = float64(sched.EnterBar-1)*beatsPerBar + 1.0
		}
		if sched.ExitBar > 0 && sched.ExitBar <= totalBars+1 {
			win.ExitBeat = float64(sched.ExitBar-1)*beatsPerBar + 1.0
		}
		out[name] = win
	}
	return out
}

// applyArrangementToEvents filters and shapes a NoteEvent slice according to
// the role's window. Events before enterBeat or at/after exitBeat are dropped.
// Events within fadeIn ramp linearly up from ~20% to 100% velocity. Events
// within fadeOut ramp 100% → ~20% velocity. The original slice is not
// mutated; a new slice is returned.
func applyArrangementToEvents(events []NoteEvent, win ArrangementBeatWindow) []NoteEvent {
	if !win.HasSchedule {
		return events
	}
	if len(events) == 0 {
		return events
	}
	out := make([]NoteEvent, 0, len(events))
	fadeInEnd := win.EnterBeat + win.FadeIn
	fadeOutStart := win.ExitBeat - win.FadeOut
	for _, ev := range events {
		if ev.Beat < win.EnterBeat {
			continue
		}
		if ev.Beat >= win.ExitBeat {
			continue
		}
		v := ev.Vel
		if v == 0 {
			v = 80
		}
		// Fade-in ramp.
		if win.FadeIn > 0 && ev.Beat < fadeInEnd {
			ratio := (ev.Beat - win.EnterBeat) / win.FadeIn
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			scale := 0.2 + 0.8*ratio
			v = int(float64(v) * scale)
		}
		// Fade-out ramp.
		if win.FadeOut > 0 && ev.Beat >= fadeOutStart {
			ratio := (win.ExitBeat - ev.Beat) / win.FadeOut
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			scale := 0.2 + 0.8*ratio
			v = int(float64(v) * scale)
		}
		if v < 1 {
			v = 1
		}
		if v > 127 {
			v = 127
		}
		copy := ev
		copy.Vel = v
		out = append(out, copy)
	}
	return out
}

// applyArrangementGatingToTrack scales the slot-velocity array of an
// authored render track according to the schedule for the role named.
// Used for algorithm-generated tracks (those that did not go through the
// event-driven path). The slot vector is per-bar-slot at authoredSlotsPerBar
// resolution.
func applyArrangementGatingToTrack(track *gen.AuthoredRenderTrack, win ArrangementBeatWindow, totalBars int) {
	if !win.HasSchedule || track == nil {
		return
	}
	const beatsPerBar = 4.0
	slotsPerBeat := float64(authoredSlotsPerBar) / beatsPerBar
	enterSlot := int((win.EnterBeat - 1.0) * slotsPerBeat)
	exitSlot := int((win.ExitBeat - 1.0) * slotsPerBeat)
	if enterSlot < 0 {
		enterSlot = 0
	}
	if exitSlot > totalBars*authoredSlotsPerBar {
		exitSlot = totalBars * authoredSlotsPerBar
	}
	if len(track.VelocityPattern) == 0 {
		return
	}
	fadeInSlots := int(win.FadeIn * slotsPerBeat)
	fadeOutSlots := int(win.FadeOut * slotsPerBeat)
	for i := range track.VelocityPattern {
		if i < enterSlot || i >= exitSlot {
			track.VelocityPattern[i] = 0
			continue
		}
		v := float64(track.VelocityPattern[i])
		// Fade-in.
		if fadeInSlots > 0 && i < enterSlot+fadeInSlots {
			ratio := float64(i-enterSlot) / float64(fadeInSlots)
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			v *= 0.2 + 0.8*ratio
		}
		// Fade-out.
		if fadeOutSlots > 0 && i >= exitSlot-fadeOutSlots {
			ratio := float64(exitSlot-i) / float64(fadeOutSlots)
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			v *= 0.2 + 0.8*ratio
		}
		track.VelocityPattern[i] = int32(v + 0.5)
		if track.VelocityPattern[i] < 0 {
			track.VelocityPattern[i] = 0
		}
	}
}

// matchTrackToRole finds the role name in the schedule map that this
// AuthoredRenderTrack belongs to. The match is loose: exact role name or
// case-insensitive prefix match against the track's Track ID or label.
func matchTrackToRole(track gen.AuthoredRenderTrack, schedules map[string]ArrangementBeatWindow) (string, bool) {
	if len(schedules) == 0 {
		return "", false
	}
	candidates := []string{track.Name}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if win, ok := schedules[c]; ok {
			_ = win
			return c, true
		}
		lc := strings.ToLower(c)
		for k := range schedules {
			lk := strings.ToLower(k)
			if lk == lc {
				return k, true
			}
			// Drum sub-name match: "kick" → "kick_drum", etc.
			if strings.HasPrefix(lc, lk) || strings.HasPrefix(lk, lc) {
				return k, true
			}
		}
	}
	return "", false
}
