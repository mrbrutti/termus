package synth

import (
	"github.com/madelynnblue/go-dsp/fft"
)

// FFTConvolver is a real-time partitioned overlap-add convolver. Suitable
// for long impulse responses (1+ seconds at 44.1 kHz) where direct
// time-domain convolution can no longer keep up. CPU work per output sample
// is O(log(blockSize) + K) instead of O(N) where N = IR length, K =
// ceil(N/blockSize); for a 2-second IR at blockSize=512 this is ~170x
// cheaper than direct convolution.
//
// Trade-offs: the convolver introduces blockSize samples of latency (the
// input is buffered until a full block is available before convolution
// happens). For typical blockSize=512 at 44.1 kHz that's ~12 ms, which is
// imperceptible for ambient music.
//
// The mathematics: each IR partition is precomputed in the frequency domain
// at construction time. For every block of input, we FFT the input, shift
// the history of past input FFTs, take the dot product of the K past input
// FFTs with the K IR partition FFTs, sum them in the frequency domain,
// IFFT, and overlap-add with the previous block's tail.
type FFTConvolver struct {
	blockSize int             // N
	fftSize   int             // 2N
	irFFT     [][]complex128  // K partitions, each fftSize long, frequency domain
	histFFT   [][]complex128  // K most recent input-block FFTs (index 0 = newest)
	overlap   []float64       // length N — tail from previous block

	inputBuffer  []float64 // length N — input accumulator
	outputBuffer []float64 // length N — current block output, drained sample by sample
	inputPos     int
	outputPos    int
}

// NewFFTConvolver builds a partitioned convolver from the given IR. The
// blockSize is the FFT block size (recommended: 256, 512, or 1024 — must be
// a power of 2 for best FFT performance). Returns nil on invalid input.
func NewFFTConvolver(ir []float64, blockSize int) *FFTConvolver {
	if len(ir) == 0 || blockSize <= 0 {
		return nil
	}
	fftSize := 2 * blockSize
	numPartitions := (len(ir) + blockSize - 1) / blockSize

	// Precompute IR partitions in frequency domain. Each partition is
	// zero-padded from blockSize to fftSize (2N) to make the FFT
	// multiplication match a linear convolution of length 2N (no wrap-around).
	irFFT := make([][]complex128, numPartitions)
	for k := 0; k < numPartitions; k++ {
		buf := make([]float64, fftSize)
		start := k * blockSize
		end := start + blockSize
		if end > len(ir) {
			end = len(ir)
		}
		copy(buf, ir[start:end])
		irFFT[k] = fft.FFTReal(buf)
	}

	histFFT := make([][]complex128, numPartitions)
	for k := range histFFT {
		histFFT[k] = make([]complex128, fftSize)
	}

	return &FFTConvolver{
		blockSize:    blockSize,
		fftSize:      fftSize,
		irFFT:        irFFT,
		histFFT:      histFFT,
		overlap:      make([]float64, blockSize),
		inputBuffer:  make([]float64, blockSize),
		outputBuffer: make([]float64, blockSize),
	}
}

// Tick processes one sample. The output is the input delayed by exactly
// blockSize samples (the convolution result for an input at time t arrives
// at the output at time t + blockSize).
func (c *FFTConvolver) Tick(x float64) float64 {
	out := c.outputBuffer[c.outputPos]
	c.inputBuffer[c.inputPos] = x
	c.inputPos++
	c.outputPos++
	if c.inputPos >= c.blockSize {
		c.inputPos = 0
		c.processBlock()
		c.outputPos = 0
	}
	return out
}

func (c *FFTConvolver) processBlock() {
	// FFT the input buffer (zero-padded to fftSize).
	padded := make([]float64, c.fftSize)
	copy(padded, c.inputBuffer)
	inputFFT := fft.FFTReal(padded)

	// Shift history: drop oldest, prepend new. Each entry is a slice header;
	// the underlying arrays for displaced entries are released to the GC.
	K := len(c.histFFT)
	for i := K - 1; i > 0; i-- {
		c.histFFT[i] = c.histFFT[i-1]
	}
	c.histFFT[0] = inputFFT

	// Frequency-domain accumulation: sum over partitions of (history * IR).
	acc := make([]complex128, c.fftSize)
	for k := 0; k < K; k++ {
		ir := c.irFFT[k]
		h := c.histFFT[k]
		for j := 0; j < c.fftSize; j++ {
			acc[j] += h[j] * ir[j]
		}
	}

	// IFFT back to time domain. The first blockSize samples are this block's
	// output; the next blockSize samples are the tail that overlap-adds into
	// the next block.
	timeDomain := fft.IFFT(acc)
	for i := 0; i < c.blockSize; i++ {
		c.outputBuffer[i] = real(timeDomain[i]) + c.overlap[i]
	}
	for i := 0; i < c.blockSize; i++ {
		c.overlap[i] = real(timeDomain[i+c.blockSize])
	}
}
