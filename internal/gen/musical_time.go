package gen

import "github.com/mrbrutti/termus/internal/synth"

func secondsToSamples(sec float64) int64 {
	if sec <= 0 {
		return 0
	}
	return int64(sec * float64(synth.SampleRate))
}

func quantizeUp(sample, step int64) int64 {
	if step <= 0 {
		return sample
	}
	rem := sample % step
	if rem == 0 {
		return sample
	}
	return sample + (step - rem)
}

func scheduleQuantizedAfter(now int64, delaySec float64, step int64) int64 {
	return quantizeUp(now+secondsToSamples(delaySec), step)
}

func crossedQuantizedBoundary(prev, curr, step int64) bool {
	if step <= 0 {
		return false
	}
	return prev/step != curr/step
}

func sampleBarIndex(samples, barSamples int64) int {
	if barSamples <= 0 {
		return 0
	}
	return int(samples / barSamples)
}
