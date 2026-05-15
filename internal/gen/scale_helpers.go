package gen

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
