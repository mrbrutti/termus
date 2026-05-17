// internal/audiotest/measure.go
//
// One-shot measurement bundle used by baseline-capture / baseline-check.
package audiotest

// Measurement is the regression-tracked summary of a rendered buffer.
type Measurement struct {
	Frames     int     `json:"frames"`
	SampleRate float64 `json:"sample_rate"`
	RMSDb      float64 `json:"rms_db"`
	PeakDb     float64 `json:"peak_db"`
	CentroidHz float64 `json:"centroid_hz"`
}

// MeasureMono computes the standard set of metrics on a mono buffer.
func MeasureMono(buf []float64, sampleRate float64) Measurement {
	return Measurement{
		Frames:     len(buf),
		SampleRate: sampleRate,
		RMSDb:      ToDB(RMS(buf)),
		PeakDb:     ToDB(Peak(buf)),
		CentroidHz: SpectralCentroidHz(buf, sampleRate),
	}
}
