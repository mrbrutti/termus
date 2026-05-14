package synth

import "math"

// Waveform selects the shape produced by an Oscillator.
type Waveform int

const (
	WaveSine Waveform = iota
	WaveSaw
	WaveTri
)

// Oscillator is a phase-accumulator oscillator. Not safe for concurrent use.
type Oscillator struct {
	wave  Waveform
	phase float64 // [0, 1)
	inc   float64 // phase increment per sample
}

func NewOscillator(w Waveform) *Oscillator {
	return &Oscillator{wave: w}
}

func (o *Oscillator) SetFrequency(hz float64) {
	o.inc = hz / float64(SampleRate)
}

// Tick advances by one sample and returns the next value in [-1, 1].
func (o *Oscillator) Tick() float64 {
	p := o.phase
	o.phase += o.inc
	if o.phase >= 1 {
		o.phase -= 1
	}
	switch o.wave {
	case WaveSine:
		return math.Sin(p * 2 * math.Pi)
	case WaveSaw:
		return 2*p - 1
	case WaveTri:
		if p < 0.5 {
			return 4*p - 1
		}
		return 3 - 4*p
	}
	return 0
}
