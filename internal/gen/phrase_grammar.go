package gen

// PhraseShape captures a simple melodic sentence arc that can be reused
// across generators: a pickup into the idea, a statement, a peak, and a
// release. Any component may be empty.
type PhraseShape struct {
	Pickup    []int
	Statement []int
	Peak      []int
	Release   []int
}

func (p PhraseShape) Phrase() []int {
	return stitchPhrase(p.Pickup, p.Statement, p.Peak, p.Release)
}

func buildPhraseMotifs(aQuestion, aAnswer PhraseShape, aprimeSub map[int]int, bQuestion, bAnswer, cadence, outro PhraseShape) MotifMemory {
	aPhrase := stitchPhrase(aQuestion.Phrase(), aAnswer.Phrase())
	return MotifMemory{
		A:       aPhrase,
		Aprime:  sequencePhrase(aPhrase, aprimeSub),
		B:       stitchPhrase(bQuestion.Phrase(), bAnswer.Phrase()),
		Cadence: cadence.Phrase(),
		Outro:   outro.Phrase(),
	}
}

