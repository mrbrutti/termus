package track

import (
	"strings"
)

// PhraseStructure (SP18) describes how a section's bars are organised into
// sub-phrases. Known values:
//
//	aaba — 4 phrases; A1 ≈ A2 (small variation), B contrast, A3 = elaborated A1
//	aabb — 4 phrases; A repeats, then B repeats
//	abab — 4 phrases; A and B alternate
//	abac — 4 phrases; A, B, A again, C concludes
//	throughcomposed — no internal repetition; each phrase fresh
//
// The phrase_structure module returns a per-phrase MotifTreatment suggestion
// the caller can apply when expanding section events. The base motif comes
// from Section.Motif; per-phrase treatments are layered on top of the
// section-level MotifTreatment.
type PhraseStructure string

// PhrasePlan is one phrase's slot inside a section.
type PhrasePlan struct {
	// Label is the phrase letter (a, b, c, ...).
	Label string
	// StartBeat is the 1-indexed beat where this phrase begins.
	StartBeat float64
	// Beats is the phrase's length in beats.
	Beats float64
	// MotifTreatment is the per-phrase motif transformation hint. The section
	// engine applies this on top of the section-level treatment.
	MotifTreatment string
}

// expandPhraseStructure returns the per-phrase plan for a section. totalBeats
// is the section's total beat span; structureName is the PhraseStructure code.
// If structureName is empty or unknown, returns a single phrase covering the
// whole section.
//
// Phrases are evenly subdivided in beats; the standard 8-bar phrase = 32
// beats holds even when the section's total bars don't divide evenly (the
// last phrase absorbs any remainder).
func expandPhraseStructure(structureName string, totalBeats float64) []PhrasePlan {
	if totalBeats <= 0 {
		return nil
	}
	code := strings.ToLower(strings.TrimSpace(structureName))
	letters := phraseLetters(code)
	if len(letters) == 0 {
		return []PhrasePlan{{Label: "a", StartBeat: 1.0, Beats: totalBeats}}
	}
	phraseBeats := totalBeats / float64(len(letters))
	plans := make([]PhrasePlan, 0, len(letters))
	usedLetter := map[string]int{}
	for i, letter := range letters {
		usedLetter[letter]++
		treatment := treatmentForPhraseSlot(letter, usedLetter[letter], i, len(letters))
		plans = append(plans, PhrasePlan{
			Label:          letter,
			StartBeat:      1.0 + phraseBeats*float64(i),
			Beats:          phraseBeats,
			MotifTreatment: treatment,
		})
	}
	return plans
}

// phraseLetters returns the letter sequence corresponding to a phrase code.
// "aaba" → ["a","a","b","a"]; "abab" → ["a","b","a","b"]; etc.
// "throughcomposed" → ["a","b","c","d"] (4-phrase fresh).
func phraseLetters(code string) []string {
	switch code {
	case "aaba":
		return []string{"a", "a", "b", "a"}
	case "aabb":
		return []string{"a", "a", "b", "b"}
	case "abab":
		return []string{"a", "b", "a", "b"}
	case "abac":
		return []string{"a", "b", "a", "c"}
	case "ab":
		return []string{"a", "b"}
	case "ba":
		return []string{"b", "a"}
	case "abcb":
		return []string{"a", "b", "c", "b"}
	case "abcd":
		return []string{"a", "b", "c", "d"}
	case "throughcomposed":
		return []string{"a", "b", "c", "d"}
	}
	return nil
}

// treatmentForPhraseSlot picks a motif treatment label for one phrase based
// on the phrase letter, its index within that letter's appearances, and the
// overall phrase index.
//
//	First appearance of any letter        → introduce
//	Second appearance of letter "a"        → vary
//	Third (or later) appearance of "a"     → return (elaborated)
//	First appearance of contrast letter   → fragment (B/C/D motifs derive)
//	Subsequent contrast appearances        → vary
func treatmentForPhraseSlot(letter string, occurrence, phraseIdx, totalPhrases int) string {
	if occurrence == 1 {
		if letter == "a" {
			return "introduce"
		}
		// Contrast phrases get a fragmented version of the motif so the
		// listener hears a relation without exact repetition.
		return "fragment"
	}
	if occurrence == 2 {
		if letter == "a" {
			return "vary"
		}
		return "vary"
	}
	// Third+ appearance: bring it back, elaborated.
	if letter == "a" && phraseIdx == totalPhrases-1 {
		return "return"
	}
	return "develop"
}
