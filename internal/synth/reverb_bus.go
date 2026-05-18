package synth

import (
	"fmt"
	"math"
)

// ReverbBus combines a convolution IR with pre-delay and wet/dry levels
// so a caller can route signal through one of several named reverbs and
// control how present it sits in the mix.
//
// The bus is stereo: the mono input is convolved through two slightly
// different IRs (seeded with seed and seed+1) to create a wide stereo image.
// Dry pass-through is the caller's responsibility — Tick only returns wet.
type ReverbBus struct {
	predelay []float64 // ring buffer for pre-delay
	pdWrite  int       // write index into predelay
	pdLen    int       // pre-delay length in samples
	convL    *FFTConvolver
	convR    *FFTConvolver
	wetLin   float64 // wet level in linear amplitude
}

// ReverbBusConfig holds all parameters needed to build a ReverbBus.
type ReverbBusConfig struct {
	IRName     string  // preset name from IRLibrary
	PreDelayMs float64 // delay before the convolver (e.g. 10–50 ms)
	WetDB      float64 // wet level in dBFS (e.g. -12)
	SampleRate float64
	Seed       int64
}

// NewReverbBus constructs a ReverbBus for the named IR preset.
// Returns an error if the name is not in IRLibrary().
func NewReverbBus(cfg ReverbBusConfig) (*ReverbBus, error) {
	preset := IRByName(cfg.IRName)
	if preset == nil {
		return nil, fmt.Errorf("reverb bus: IR preset %q not found", cfg.IRName)
	}
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = float64(SampleRate)
	}

	// Generate two correlated-but-distinct IRs for stereo width.
	irL := preset.Generate(cfg.SampleRate, cfg.Seed)
	irR := preset.Generate(cfg.SampleRate, cfg.Seed+1)

	// Normalize each IR independently (cube-root normalization as in sf2_engine).
	normalizeIR(irL)
	normalizeIR(irR)

	const blockSize = 512
	convL := NewFFTConvolver(irL, blockSize)
	convR := NewFFTConvolver(irR, blockSize)

	// Pre-delay ring buffer. Length must be at least 1.
	pdSamples := int(math.Round(cfg.PreDelayMs * 0.001 * cfg.SampleRate))
	if pdSamples < 1 {
		pdSamples = 1
	}

	wetLin := math.Pow(10, cfg.WetDB/20)

	return &ReverbBus{
		predelay: make([]float64, pdSamples),
		pdLen:    pdSamples,
		convL:    convL,
		convR:    convR,
		wetLin:   wetLin,
	}, nil
}

// Tick processes one mono input sample and returns the wet stereo output.
// Dry pass-through is the caller's responsibility.
func (b *ReverbBus) Tick(inL, inR float64) (wetL, wetR float64) {
	// Mix to mono for the pre-delay line.
	mono := 0.5 * (inL + inR)

	// Pre-delay: read out the oldest sample, write the new one.
	delayed := b.predelay[b.pdWrite]
	b.predelay[b.pdWrite] = mono
	b.pdWrite = (b.pdWrite + 1) % b.pdLen

	// Convolve through L and R independently for stereo width.
	wetL = b.convL.Tick(delayed) * b.wetLin
	wetR = b.convR.Tick(delayed) * b.wetLin
	return wetL, wetR
}

// normalizeIR scales an IR in-place using cube-root normalization so that a
// dense long IR doesn't blow up the convolved output. Cube-root is a
// perceptual compromise: full power-normalization makes long IRs too quiet.
func normalizeIR(ir []float64) {
	var sumSq float64
	for _, x := range ir {
		sumSq += x * x
	}
	if sumSq <= 0 {
		return
	}
	norm := math.Pow(1.0/sumSq, 1.0/3.0)
	for i := range ir {
		ir[i] *= norm
	}
}
