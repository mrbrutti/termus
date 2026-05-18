package track

import (
	"fmt"
	"strings"
)

// FormTemplate (SP18) is one named multi-section form. Each template carries
// a default ordered list of sections — id, semantic role, bar count, harmony
// archetype, default motif treatment, and the typical transition into the
// next section. Templates are referenced by File.Form; if the authored file
// also has an explicit Sections list, the explicit list overrides the
// template (the template is then advisory only).
//
// Templates are intentionally style-flavoured so each genre has 2-3 idiomatic
// forms ready to use without elaborate authoring.
//
// All harmonies use absolute chord symbols (Cmaj7, Am7, Dm7, G7, ...). When
// users want a different key they can copy a template's section list into
// their .tm file and transpose, or pass Sections explicitly.
type FormTemplate struct {
	Name        string
	Description string
	Sections    []FormSection
	// DefaultBPM is a hint used when expanding the template's section bars
	// into duration strings if the authored file does not set a tempo.
	DefaultBPM float64
}

// FormSection is one entry in a FormTemplate's section list.
type FormSection struct {
	ID               string
	Role             string
	Bars             int
	Harmony          string // 4/4 bars, "|" separated, absolute chord symbols
	MotifTreatment   string
	PhraseStructure  string
	DynamicCurve     string
	TransitionToNext string
	Arrangement      map[string]RoleSchedule
}

var formRegistry = map[string]FormTemplate{}

func init() {
	registerForms(builtInForms()...)
}

// ResolveForm returns the named FormTemplate, or false if unknown.
func ResolveForm(name string) (FormTemplate, bool) {
	if name == "" {
		return FormTemplate{}, false
	}
	t, ok := formRegistry[strings.ToLower(strings.TrimSpace(name))]
	return t, ok
}

// FormNames returns the registered template names, sorted.
func FormNames() []string {
	names := make([]string, 0, len(formRegistry))
	for k := range formRegistry {
		names = append(names, k)
	}
	return names
}

func registerForms(templates ...FormTemplate) {
	for _, t := range templates {
		formRegistry[strings.ToLower(t.Name)] = t
	}
}

// expandFormTemplate converts a FormTemplate into a []Section list ready for
// the rest of the compile pipeline. BPM is used to convert bar counts into
// duration strings. If bpm <= 0, defaults to t.DefaultBPM (or 90 as last
// resort).
func expandFormTemplate(t FormTemplate, bpm float64) []Section {
	if bpm <= 0 {
		bpm = t.DefaultBPM
	}
	if bpm <= 0 {
		bpm = 90
	}
	out := make([]Section, 0, len(t.Sections))
	for _, fs := range t.Sections {
		sec := Section{
			ID:               fs.ID,
			Title:            fs.ID,
			Role:             fs.Role,
			Bars:             fs.Bars,
			Harmony:          fs.Harmony,
			MotifTreatment:   fs.MotifTreatment,
			PhraseStructure:  fs.PhraseStructure,
			DynamicCurve:     fs.DynamicCurve,
			TransitionToNext: fs.TransitionToNext,
			Arrangement18:    cloneSchedule(fs.Arrangement),
			Duration:         barsToDurationString(fs.Bars, bpm),
			Events:           formSectionEvents(fs),
		}
		out = append(out, sec)
	}
	return out
}

// formSectionEvents (SP19) synthesises arrangement-level Events for a
// form-driven section so review/discover tooling sees the implicit moments
// the form template describes (transitions, role entries/exits). These
// events do not change rendering — they document the form structurally.
func formSectionEvents(fs FormSection) []Event {
	var out []Event
	if kind := mapTransitionKind(fs.TransitionToNext); kind != "" {
		out = append(out, Event{Kind: kind, Bars: 1, Slot: 0})
	}
	for role, sched := range fs.Arrangement {
		if sched.EnterBar > 1 {
			out = append(out, Event{Kind: "pickup", Bar: sched.EnterBar, Slot: 0, Roles: []string{role}})
		}
		if sched.ExitBar > 0 && sched.ExitBar <= fs.Bars {
			out = append(out, Event{Kind: "tag", Bar: sched.ExitBar, Slot: 0, Roles: []string{role}})
		}
	}
	return out
}

// mapTransitionKind maps a transition_to_next value to a validateEvent-known
// event kind. Unknown / empty returns "" (skip).
func mapTransitionKind(transition string) string {
	switch strings.ToLower(strings.TrimSpace(transition)) {
	case "turnaround", "fill":
		return "fill"
	case "pickup":
		return "pickup"
	case "swell":
		return "swell"
	case "breakdown":
		return "break"
	}
	return ""
}

func cloneSchedule(in map[string]RoleSchedule) map[string]RoleSchedule {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]RoleSchedule, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// barsToDurationString returns "Ns" or "NmMs" derived from bars × beats-per-bar
// / BPM. Uses 4/4 (4 beats per bar). Returns "" when bars or bpm are zero.
func barsToDurationString(bars int, bpm float64) string {
	if bars <= 0 || bpm <= 0 {
		return ""
	}
	seconds := float64(bars) * 4.0 * 60.0 / bpm
	secs := int(seconds + 0.5)
	if secs <= 0 {
		secs = 1
	}
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	m := secs / 60
	s := secs % 60
	if s == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm%ds", m, s)
}

// builtInForms returns the SP18 default form template catalogue.
func builtInForms() []FormTemplate {
	return []FormTemplate{
		jazzAABA32(),
		jazzBlues12(),
		jazzHeadSoloHead(),
		lofiLoopForm(),
		chillABABCB(),
		chillJourney(),
		ambientEmergeDriftRecede(),
		ambientPalindrome(),
	}
}

// ------------------------- JAZZ (in C major / A minor) -------------------------

// jazzAABA32 — 32-bar AABA ballad in C major. Intro + head A1 + A2 + bridge +
// A3 + solo A1 + solo A2 + head out + outro. ~5-6 minutes at 110 BPM.
func jazzAABA32() FormTemplate {
	enterAll := map[string]RoleSchedule{
		"bass":   {EnterBar: 1},
		"kick":   {EnterBar: 1},
		"snare":  {EnterBar: 1},
		"hat":    {EnterBar: 1},
		"piano":  {EnterBar: 1},
		"keys":   {EnterBar: 1},
		"rhodes": {EnterBar: 1},
		"pad":    {EnterBar: 1},
	}
	aSection := "Cmaj7 Am7 | Dm7 G7 | Cmaj7 F7 | Em7 Am7 | Dm7 | G7 | Cmaj7 | G7"
	aSectionEnd := "Cmaj7 Am7 | Dm7 G7 | Cmaj7 F7 | Em7 Am7 | Dm7 | G7 | Cmaj7 | Cmaj7"
	bridge := "Fmaj7 | Fmaj7 | Fm7 | Bb7 | Em7 | A7 | Dm7 | G7"
	return FormTemplate{
		Name:        "jazz_aaba_32bar",
		Description: "32-bar AABA ballad in C, intro + AA bridge A + solos + head out + outro (~5-6m at 110 BPM)",
		DefaultBPM:  110,
		Sections: []FormSection{
			{
				ID: "intro", Role: "intro", Bars: 8,
				Harmony:          "Cmaj7 | Am7 | Dm7 | G7 | Cmaj7 | Am7 | Dm7 | G7",
				MotifTreatment:   "hint",
				DynamicCurve:     "crescendo",
				TransitionToNext: "pickup",
				Arrangement: map[string]RoleSchedule{
					"pad":    {EnterBar: 1},
					"piano":  {EnterBar: 1, FadeInBars: 2},
					"rhodes": {EnterBar: 1, FadeInBars: 2},
					"bass":   {EnterBar: 5},
				},
			},
			{ID: "head_a1", Role: "head_statement", Bars: 8, Harmony: aSection, MotifTreatment: "introduce", PhraseStructure: "ab", DynamicCurve: "arc", TransitionToNext: "turnaround", Arrangement: enterAll},
			{ID: "head_a2", Role: "head_variation", Bars: 8, Harmony: aSection, MotifTreatment: "vary", PhraseStructure: "ab", DynamicCurve: "arc", TransitionToNext: "turnaround", Arrangement: enterAll},
			{ID: "bridge", Role: "contrast", Bars: 8, Harmony: bridge, MotifTreatment: "fragment", PhraseStructure: "ab", DynamicCurve: "wave", TransitionToNext: "swell", Arrangement: enterAll},
			{ID: "head_a3", Role: "head_return", Bars: 8, Harmony: aSectionEnd, MotifTreatment: "return", PhraseStructure: "ab", DynamicCurve: "arc", TransitionToNext: "fill", Arrangement: enterAll},
			{ID: "solo_a1", Role: "solo", Bars: 8, Harmony: aSection, MotifTreatment: "develop", PhraseStructure: "ab", DynamicCurve: "arc", TransitionToNext: "turnaround", Arrangement: enterAll},
			{ID: "solo_a2", Role: "solo", Bars: 8, Harmony: aSection, MotifTreatment: "develop", PhraseStructure: "ab", DynamicCurve: "arc", TransitionToNext: "swell", Arrangement: enterAll},
			{ID: "head_out", Role: "head_return", Bars: 8, Harmony: aSectionEnd, MotifTreatment: "return", PhraseStructure: "ab", DynamicCurve: "decrescendo", TransitionToNext: "fill", Arrangement: enterAll},
			{
				ID: "outro", Role: "outro", Bars: 8,
				Harmony:          "Cmaj7 | Am7 | Dm7 | G7 | Cmaj7 | Cmaj7 | Cmaj7 | Cmaj7",
				MotifTreatment:   "fragment",
				DynamicCurve:     "decrescendo",
				Arrangement: map[string]RoleSchedule{
					"kick":   {ExitBar: 5, FadeOutBars: 2},
					"snare":  {ExitBar: 5, FadeOutBars: 2},
					"hat":    {ExitBar: 7},
					"bass":   {ExitBar: 7, FadeOutBars: 2},
					"piano":  {FadeOutBars: 4},
					"rhodes": {FadeOutBars: 4},
					"pad":    {EnterBar: 1},
				},
			},
		},
	}
}

// jazzBlues12 — 12-bar blues in C, 4 choruses.
func jazzBlues12() FormTemplate {
	chorus := "C7 | F7 | C7 | C7 | F7 | F7 | C7 | A7 | Dm7 | G7 | C7 A7 | Dm7 G7"
	chorusEnd := "C7 | F7 | C7 | C7 | F7 | F7 | C7 | A7 | Dm7 | G7 | C7 | C7"
	return FormTemplate{
		Name:        "jazz_blues_12bar",
		Description: "12-bar blues in C: head, two solos, head out",
		DefaultBPM:  120,
		Sections: []FormSection{
			{ID: "head_in", Role: "head_statement", Bars: 12, Harmony: chorus, MotifTreatment: "introduce", DynamicCurve: "arc", TransitionToNext: "turnaround"},
			{ID: "solo1", Role: "solo", Bars: 12, Harmony: chorus, MotifTreatment: "develop", DynamicCurve: "arc", TransitionToNext: "turnaround"},
			{ID: "solo2", Role: "solo", Bars: 12, Harmony: chorus, MotifTreatment: "develop", DynamicCurve: "crescendo", TransitionToNext: "swell"},
			{ID: "head_out", Role: "head_return", Bars: 12, Harmony: chorusEnd, MotifTreatment: "return", DynamicCurve: "decrescendo", TransitionToNext: "fill"},
		},
	}
}

// jazzHeadSoloHead — extended jazz form (32-bar chorus in C), ~6m at 120 BPM.
func jazzHeadSoloHead() FormTemplate {
	chorus := "Cmaj7 Am7 | Dm7 G7 | Cmaj7 F7 | Em7 Am7 | Dm7 | G7 | Cmaj7 | G7 | " +
		"Cmaj7 Am7 | Dm7 G7 | Cmaj7 F7 | Em7 Am7 | Dm7 | G7 | Cmaj7 | G7 | " +
		"Fmaj7 | Fmaj7 | Fm7 | Bb7 | Em7 | A7 | Dm7 | G7 | " +
		"Cmaj7 Am7 | Dm7 G7 | Cmaj7 F7 | Em7 Am7 | Dm7 | G7 | Cmaj7 | G7"
	chorusEnd := strings.Replace(chorus[:len(chorus)-2]+"Cmaj7", "Cmaj7 | G7", "Cmaj7 | Cmaj7", -1)
	_ = chorusEnd // we use the head structure directly
	return FormTemplate{
		Name:        "jazz_head_solo_head",
		Description: "Head–solo–solo–head extended jazz form (32-bar chorus in C, ~6m at 120 BPM)",
		DefaultBPM:  120,
		Sections: []FormSection{
			{
				ID: "intro", Role: "intro", Bars: 8,
				Harmony:          "Cmaj7 | Am7 | Dm7 | G7 | Cmaj7 | Am7 | Dm7 | G7",
				MotifTreatment:   "hint",
				DynamicCurve:     "crescendo",
				TransitionToNext: "pickup",
			},
			{ID: "head_in", Role: "head_statement", Bars: 32, Harmony: chorus, MotifTreatment: "introduce", PhraseStructure: "aaba", DynamicCurve: "arc", TransitionToNext: "turnaround"},
			{ID: "solo_chorus1", Role: "solo", Bars: 32, Harmony: chorus, MotifTreatment: "develop", PhraseStructure: "aaba", DynamicCurve: "arc", TransitionToNext: "turnaround"},
			{ID: "solo_chorus2", Role: "solo", Bars: 32, Harmony: chorus, MotifTreatment: "fragment", PhraseStructure: "aaba", DynamicCurve: "wave", TransitionToNext: "swell"},
			{ID: "head_out", Role: "head_return", Bars: 32, Harmony: chorus, MotifTreatment: "return", PhraseStructure: "aaba", DynamicCurve: "decrescendo", TransitionToNext: "fill"},
		},
	}
}

// ------------------------- LOFI (in D minor) -------------------------

func lofiLoopForm() FormTemplate {
	intro := "Dm9 | Dm9 | Gm7 | Gm7 | Dm9 | Dm9 | A7 | A7"
	loop := "Dm9 Gm7 | Bb6 A7 | Dm9 Gm7 | Bb6 A7 | Dm9 Gm7 | Bb6 A7 | Dm9 A7 | Dm9 A7 | Dm9 Gm7 | Bb6 A7 | Dm9 Gm7 | Bb6 A7 | Dm9 Gm7 | Bb6 A7 | Dm9 A7 | Dm9 A7"
	bridge := "Bbmaj7 | Bbmaj7 | Am7 | Am7 | Dm7 | Dm7 | G7 | G7"
	outro := "Dm9 | Dm9 | A7sus | A7sus | Dm9 | Dm9 | Dm7 | Dm7"
	return FormTemplate{
		Name:        "lofi_loop_form",
		Description: "Intro–loop–loop variation–breakdown–loop dense–outro lofi arrangement in D minor",
		DefaultBPM:  84,
		Sections: []FormSection{
			{
				ID: "intro", Role: "emerge", Bars: 8,
				Harmony:          intro,
				MotifTreatment:   "hint",
				DynamicCurve:     "crescendo",
				TransitionToNext: "pickup",
				Arrangement: map[string]RoleSchedule{
					"pad":    {EnterBar: 1},
					"rhodes": {EnterBar: 5, FadeInBars: 4},
					"bass":   {EnterBar: 1},
				},
			},
			{
				ID: "loop_a", Role: "head_statement", Bars: 16,
				Harmony:          loop,
				MotifTreatment:   "introduce",
				PhraseStructure:  "aaba",
				DynamicCurve:     "arc",
				TransitionToNext: "turnaround",
				Arrangement: map[string]RoleSchedule{
					"pad":    {ExitBar: 9, FadeOutBars: 2},
					"rhodes": {EnterBar: 1},
					"bass":   {EnterBar: 1},
					"kick":   {EnterBar: 1},
					"snare":  {EnterBar: 1},
					"hat":    {EnterBar: 1},
				},
			},
			{
				ID: "loop_b", Role: "head_variation", Bars: 16,
				Harmony:          loop,
				MotifTreatment:   "develop",
				PhraseStructure:  "aaba",
				DynamicCurve:     "arc",
				TransitionToNext: "breakdown",
				Arrangement: map[string]RoleSchedule{
					"rhodes": {EnterBar: 1},
					"bass":   {EnterBar: 1},
					"kick":   {EnterBar: 1},
					"snare":  {EnterBar: 1},
					"hat":    {EnterBar: 1},
				},
			},
			{
				ID: "bridge", Role: "contrast", Bars: 8,
				Harmony:          bridge,
				MotifTreatment:   "fragment",
				DynamicCurve:     "decrescendo",
				TransitionToNext: "swell",
				Arrangement: map[string]RoleSchedule{
					"rhodes": {ExitBar: 5, FadeOutBars: 2},
					"bass":   {EnterBar: 1},
					"kick":   {ExitBar: 1},
					"snare":  {ExitBar: 1},
					"hat":    {EnterBar: 1},
					"pad":    {EnterBar: 1},
				},
			},
			{
				ID: "loop_c", Role: "head_return", Bars: 16,
				Harmony:          loop,
				MotifTreatment:   "return",
				PhraseStructure:  "aaba",
				DynamicCurve:     "arc",
				TransitionToNext: "fill",
				Arrangement: map[string]RoleSchedule{
					"rhodes": {EnterBar: 1},
					"bass":   {EnterBar: 1},
					"kick":   {EnterBar: 1},
					"snare":  {EnterBar: 1},
					"hat":    {EnterBar: 1},
				},
			},
			{
				ID: "outro", Role: "recede", Bars: 8,
				Harmony:          outro,
				MotifTreatment:   "fragment",
				DynamicCurve:     "decrescendo",
				Arrangement: map[string]RoleSchedule{
					"kick":   {ExitBar: 5, FadeOutBars: 2},
					"snare":  {ExitBar: 5, FadeOutBars: 2},
					"hat":    {ExitBar: 7},
					"bass":   {ExitBar: 7},
					"rhodes": {FadeOutBars: 4},
					"pad":    {EnterBar: 5},
				},
			},
		},
	}
}

// ------------------------- CHILL (in C major) -------------------------

func chillABABCB() FormTemplate {
	intro := "Cmaj9 | Am7 | Fmaj9 | G7 | Cmaj9 | Am7 | Fmaj9 | G7"
	verse := "Cmaj9 | Em7 | Am7 | Fmaj9 | Cmaj9 | Em7 | Dm7 | G7 | Cmaj9 | Em7 | Am7 | Fmaj9 | Cmaj9 | Em7 | Dm7 | G7"
	verseEnd := "Cmaj9 | Em7 | Am7 | Fmaj9 | Cmaj9 | Em7 | Dm7 | G7 | Cmaj9 | Em7 | Am7 | Fmaj9 | Cmaj9 | Em7 | Dm7 | Cmaj9"
	bridge := "Fmaj9 | Fmaj9 | Fm7 | Fm7 | Bb7 | Bb7 | Em7 | A7 | Dm7 | Dm7 | G7 | G7"
	return FormTemplate{
		Name:        "chill_ababcb",
		Description: "ABABCB chill journey in C with bridge contrast and final B return",
		DefaultBPM:  90,
		Sections: []FormSection{
			{
				ID: "intro_a", Role: "emerge", Bars: 8,
				Harmony:          intro,
				MotifTreatment:   "hint",
				DynamicCurve:     "crescendo",
				TransitionToNext: "pickup",
				Arrangement: map[string]RoleSchedule{
					"pad":  {EnterBar: 1},
					"keys": {EnterBar: 5, FadeInBars: 2},
					"bass": {EnterBar: 5},
				},
			},
			{
				ID: "verse_b1", Role: "head_statement", Bars: 16,
				Harmony:          verse,
				MotifTreatment:   "introduce",
				PhraseStructure:  "aaba",
				DynamicCurve:     "arc",
				TransitionToNext: "turnaround",
				Arrangement: map[string]RoleSchedule{
					"pad":   {ExitBar: 9, FadeOutBars: 2},
					"keys":  {EnterBar: 1},
					"bass":  {EnterBar: 1},
					"kick":  {EnterBar: 1},
					"snare": {EnterBar: 1},
					"hat":   {EnterBar: 1},
				},
			},
			{
				ID: "verse_a2", Role: "head_variation", Bars: 16,
				Harmony:          verse,
				MotifTreatment:   "develop",
				PhraseStructure:  "aaba",
				DynamicCurve:     "arc",
				TransitionToNext: "swell",
				Arrangement: map[string]RoleSchedule{
					"keys":  {EnterBar: 1},
					"bass":  {EnterBar: 1},
					"kick":  {EnterBar: 1},
					"snare": {EnterBar: 1},
					"hat":   {EnterBar: 1},
				},
			},
			{
				ID: "verse_b2", Role: "head_variation", Bars: 16,
				Harmony:          verse,
				MotifTreatment:   "vary",
				PhraseStructure:  "aaba",
				DynamicCurve:     "arc",
				TransitionToNext: "breakdown",
				Arrangement: map[string]RoleSchedule{
					"keys":  {EnterBar: 1},
					"bass":  {EnterBar: 1},
					"kick":  {EnterBar: 1},
					"snare": {EnterBar: 1},
					"hat":   {EnterBar: 1},
				},
			},
			{
				ID: "bridge_c", Role: "contrast", Bars: 12,
				Harmony:          bridge,
				MotifTreatment:   "fragment",
				DynamicCurve:     "wave",
				TransitionToNext: "swell",
				Arrangement: map[string]RoleSchedule{
					"keys":  {EnterBar: 1},
					"bass":  {EnterBar: 1},
					"kick":  {ExitBar: 1},
					"snare": {ExitBar: 1},
					"hat":   {EnterBar: 1},
					"pad":   {EnterBar: 1},
				},
			},
			{
				ID: "return_b", Role: "head_return", Bars: 16,
				Harmony:          verseEnd,
				MotifTreatment:   "return",
				PhraseStructure:  "aaba",
				DynamicCurve:     "decrescendo",
				TransitionToNext: "fill",
				Arrangement: map[string]RoleSchedule{
					"keys":  {EnterBar: 1, FadeOutBars: 4},
					"bass":  {EnterBar: 1, FadeOutBars: 4},
					"kick":  {EnterBar: 1, ExitBar: 13},
					"snare": {EnterBar: 1, ExitBar: 13},
					"hat":   {EnterBar: 1, ExitBar: 15},
					"pad":   {EnterBar: 9},
				},
			},
		},
	}
}

func chillJourney() FormTemplate {
	intro := "Cmaj9 | Cmaj9 | Am7 | Am7 | Fmaj9 | Fmaj9 | G7 | G7 | Cmaj9 | Cmaj9 | Am7 | Am7 | Fmaj9 | Fmaj9 | G7 | G7"
	verse := "Cmaj9 | Em7 | Am7 | Fmaj9 | Dm7 | G7 | Cmaj9 | Cmaj9 | Cmaj9 | Em7 | Am7 | Fmaj9 | Dm7 | G7 | Cmaj9 | Cmaj9 | Cmaj9 | Em7 | Am7 | Fmaj9 | Dm7 | G7 | Cmaj9 | Cmaj9"
	bridge := "Fmaj9 | Fmaj9 | Fm7 | Fm7 | Bb7 | Bb7 | Em7 | A7 | Dm7 | Dm7 | G7 | G7 | Cmaj9 | Am7 | Dm7 | G7"
	climax := "Cmaj9 | Em7 | Am7 | Fmaj9 | Dm7 | G7 | Cmaj9 | Cmaj9 | Cmaj9 | Em7 | Am7 | Fmaj9 | Dm7 | G7 | Cmaj9 | Cmaj9"
	outro := "Cmaj9 | Cmaj9 | Am7 | Am7 | Fmaj9 | Fmaj9 | G7 | G7 | Cmaj9 | Cmaj9 | Am7 | Am7 | Fmaj9 | Fmaj9 | Cmaj9 | Cmaj9"
	return FormTemplate{
		Name:        "chill_journey",
		Description: "Long-form chill 6-part journey in C, ~7m at 88 BPM",
		DefaultBPM:  88,
		Sections: []FormSection{
			{ID: "intro", Role: "emerge", Bars: 16, Harmony: intro, MotifTreatment: "hint", DynamicCurve: "crescendo", TransitionToNext: "pickup"},
			{ID: "verse1", Role: "head_statement", Bars: 24, Harmony: verse, MotifTreatment: "introduce", PhraseStructure: "aaba", DynamicCurve: "arc", TransitionToNext: "turnaround"},
			{ID: "verse2", Role: "head_variation", Bars: 24, Harmony: verse, MotifTreatment: "develop", PhraseStructure: "aaba", DynamicCurve: "arc", TransitionToNext: "breakdown"},
			{ID: "bridge", Role: "contrast", Bars: 16, Harmony: bridge, MotifTreatment: "fragment", DynamicCurve: "wave", TransitionToNext: "swell"},
			{ID: "climax", Role: "climax", Bars: 16, Harmony: climax, MotifTreatment: "return", PhraseStructure: "aaba", DynamicCurve: "arc", TransitionToNext: "fill"},
			{ID: "outro", Role: "recede", Bars: 16, Harmony: outro, MotifTreatment: "fragment", DynamicCurve: "decrescendo"},
		},
	}
}

// ------------------------- AMBIENT (in C major) -------------------------

func ambientEmergeDriftRecede() FormTemplate {
	emerge := repeatBar("Cmaj9 | Cmaj9 | Gmaj9 | Gmaj9 | Fmaj9 | Fmaj9 | Cmaj9 | Cmaj9 | ", 7) + "Cmaj9 | Cmaj9 | Cmaj9 | Cmaj9"
	drift := repeatBar("Fmaj9 | Fmaj9 | Fm7 | Fm7 | Abmaj9 | Abmaj9 | Cmaj9 | Cmaj9 | ", 5) +
		repeatBar("Ebmaj9 | Ebmaj9 | Abmaj9 | Abmaj9 | Fm7 | Fm7 | Bb7 | Bb7 | ", 2) +
		repeatBar("Cmaj9 | Cmaj9 | Gmaj9 | Gmaj9 | Fmaj9 | Fmaj9 | Cmaj9 | Cmaj9 | ", 1) +
		repeatBar("Cmaj9 | Cmaj9 | ", 4)
	recede := repeatBar("Cmaj9 | Cmaj9 | Cmaj9 | Cmaj9 | Gmaj9 | Gmaj9 | Cmaj9 | Cmaj9 | ", 1) +
		repeatBar("Cmaj9 | Cmaj9 | Fmaj9 | Fmaj9 | Cmaj9 | Cmaj9 | Cmaj9 | Cmaj9 | ", 1) +
		repeatBar("Cmaj9 | ", 24)
	// Trim trailing " | " if present.
	emerge = trimBar(emerge)
	drift = trimBar(drift)
	recede = trimBar(recede)
	return FormTemplate{
		Name:        "ambient_emerge_drift_recede",
		Description: "Three-act ambient long-form in C: emerge (60 bars), drift (80 bars), recede (40 bars), ~3m at 60 BPM each",
		DefaultBPM:  60,
		Sections: []FormSection{
			{
				ID: "emerge", Role: "emerge", Bars: 60,
				Harmony:          emerge,
				MotifTreatment:   "hint",
				DynamicCurve:     "crescendo",
				TransitionToNext: "swell",
				Arrangement: map[string]RoleSchedule{
					"pad":   {EnterBar: 1, FadeInBars: 16},
					"drone": {EnterBar: 1, FadeInBars: 16},
					"bell":  {EnterBar: 25, FadeInBars: 8},
					"bass":  {EnterBar: 33, FadeInBars: 8},
				},
			},
			{
				ID: "drift", Role: "drift", Bars: 80,
				Harmony:          drift,
				MotifTreatment:   "develop",
				DynamicCurve:     "wave",
				TransitionToNext: "swell",
				Arrangement: map[string]RoleSchedule{
					"pad":   {EnterBar: 1},
					"drone": {EnterBar: 1},
					"bell":  {EnterBar: 1},
					"bass":  {EnterBar: 1},
				},
			},
			{
				ID: "recede", Role: "recede", Bars: 40,
				Harmony:          recede,
				MotifTreatment:   "fragment",
				DynamicCurve:     "decrescendo",
				Arrangement: map[string]RoleSchedule{
					"pad":   {EnterBar: 1, FadeOutBars: 16},
					"drone": {EnterBar: 1, FadeOutBars: 16},
					"bell":  {ExitBar: 17, FadeOutBars: 4},
					"bass":  {ExitBar: 25, FadeOutBars: 6},
				},
			},
		},
	}
}

func ambientPalindrome() FormTemplate {
	a := repeatBar("Cmaj9 | Cmaj9 | Gmaj9 | Gmaj9 | Fmaj9 | Fmaj9 | Cmaj9 | Cmaj9 | ", 4)
	b := repeatBar("Fmaj9 | Fmaj9 | Abmaj9 | Abmaj9 | Fm7 | Fm7 | Bb7 | Bb7 | ", 4)
	c := repeatBar("Ebmaj9 | Ebmaj9 | Abmaj9 | Abmaj9 | Fm7 | Fm7 | Bb7 | Bb7 | ", 2) +
		repeatBar("Cmaj9 | Cmaj9 | Gmaj9 | Gmaj9 | Fmaj9 | Fmaj9 | Cmaj9 | Cmaj9 | ", 2)
	a = trimBar(a)
	b = trimBar(b)
	c = trimBar(c)
	return FormTemplate{
		Name:        "ambient_palindrome",
		Description: "ABCBA palindrome ambient in C — sections mirror around the centre",
		DefaultBPM:  60,
		Sections: []FormSection{
			{ID: "a1", Role: "emerge", Bars: 32, Harmony: a, MotifTreatment: "introduce", DynamicCurve: "crescendo", TransitionToNext: "swell"},
			{ID: "b1", Role: "drift", Bars: 32, Harmony: b, MotifTreatment: "vary", DynamicCurve: "wave", TransitionToNext: "swell"},
			{ID: "c", Role: "climax", Bars: 32, Harmony: c, MotifTreatment: "develop", DynamicCurve: "arc", TransitionToNext: "swell"},
			{ID: "b2", Role: "drift", Bars: 32, Harmony: b, MotifTreatment: "return", DynamicCurve: "wave", TransitionToNext: "swell"},
			{ID: "a2", Role: "recede", Bars: 32, Harmony: a, MotifTreatment: "fragment", DynamicCurve: "decrescendo"},
		},
	}
}

// repeatBar repeats a chunk of "X | Y | " text N times.
func repeatBar(chunk string, times int) string {
	out := strings.Builder{}
	for i := 0; i < times; i++ {
		out.WriteString(chunk)
	}
	return out.String()
}

// trimBar removes a trailing " | " or trailing "|" if present so the result
// is a valid harmony pattern (no empty bar at the end).
func trimBar(s string) string {
	s = strings.TrimSpace(s)
	for strings.HasSuffix(s, "|") {
		s = strings.TrimSpace(s[:len(s)-1])
	}
	return s
}
