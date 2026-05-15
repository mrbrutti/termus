package gen

import "math/rand"

// melodicPhrases are reusable scale-degree contours that read as musical
// rather than random. Each entry is a list of scale-degree offsets (which
// can go negative to dip below the starting note). Algorithms pick one and
// transpose to a chord-relative starting degree, producing a coherent
// melodic shape rather than a per-note random walk.
//
// These are not exhaustive — pick whichever fits the mood: ascending lines
// open energy, descending lines close it, arches imply resolution, waves
// imply repose.
var melodicPhrases = [][]int{
	{0, 2, 4, 2, 0, 2, 4, 6},   // ascending → plateau
	{6, 4, 2, 0, 2, 4, 2, 0},   // descending → return
	{0, 2, 4, 6, 4, 2, 0, -2},  // peak-and-fall (classical arch)
	{0, -2, 0, 2, 4, 2, 0, -2}, // wave (repose)
	{0, 4, 2, 6, 4, 0, 2, 0},   // jazz-ish skip
	{0, 2, 0, 4, 2, 0, -2, 0},  // call-and-response
	{4, 2, 0, 2, 4, 6, 4, 2},   // recovery arch
}

// pickMelodicPhrase returns one melodic-phrase contour, randomly chosen.
// Slice is read-only — callers should not mutate it.
func pickMelodicPhrase(rng *rand.Rand) []int {
	return melodicPhrases[rng.Intn(len(melodicPhrases))]
}

// applyPhraseToScale converts an N-element melodic-phrase contour (relative
// scale-degree offsets) into N concrete MIDI keys using the given scale and
// root MIDI. octave is the base octave bump (0 = at root, 12 = +1 octave,
// etc.). startDegree is the scale degree the phrase's "0" maps to.
//
// Returns a fresh slice — caller may mutate freely.
func applyPhraseToScale(phrase []int, scale []int, rootMidi, startDegree, octave int) []int {
	if len(scale) == 0 {
		out := make([]int, len(phrase))
		for i := range out {
			out[i] = rootMidi + octave
		}
		return out
	}
	out := make([]int, len(phrase))
	for i, off := range phrase {
		deg := startDegree + off
		// Compute octave wrap so negative-going contours stay valid.
		oct := 0
		for deg < 0 {
			deg += len(scale)
			oct--
		}
		for deg >= len(scale) {
			deg -= len(scale)
			oct++
		}
		out[i] = rootMidi + scale[deg] + 12*oct + octave
	}
	return out
}

// scalePitchLoc identifies a pitch's location relative to a scale: which
// degree it is, and how many octaves above/below the reference root. Used by
// mutation closures that want to walk from "wherever we currently are"
// rather than re-rolling from scratch.
type scalePitchLoc struct {
	degreeIdx    int // index into the scale slice (0..len(scale)-1)
	octaveOffset int // octaves from the reference root
}

// findClosestScalePitch returns the closest scale-anchored pitch to `midi`,
// expressed as (degree index, octave offset). The reference root is
// rootMidi; `scale` is a slice of semitone offsets within an octave (e.g.
// pentatonic minor = {0,3,5,7,10}).
func findClosestScalePitch(midi, rootMidi int, scale []int) scalePitchLoc {
	rel := midi - rootMidi
	octave := rel / 12
	semi := rel % 12
	if semi < 0 {
		semi += 12
		octave--
	}
	bestIdx := 0
	bestDist := 1 << 30
	for i, s := range scale {
		d := semi - s
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return scalePitchLoc{degreeIdx: bestIdx, octaveOffset: octave}
}
