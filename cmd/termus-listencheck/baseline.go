package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/mrbrutti/termus/internal/audiotest"
)

const (
	rmsToleranceDB     = 1.0
	peakToleranceDB    = 1.5
	centroidToleranceR = 0.10
)

type measurement struct {
	Frames     int     `json:"frames"`
	SampleRate float64 `json:"sample_rate"`
	RMSDb      float64 `json:"rms_db"`
	PeakDb     float64 `json:"peak_db"`
	CentroidHz float64 `json:"centroid_hz"`
}

type baselineEntry struct {
	Name        string      `json:"name"`
	Algo        string      `json:"algo"`
	Seed        int64       `json:"seed"`
	Seconds     float64     `json:"seconds"`
	Measurement measurement `json:"measurement"`
}

func measurementFromAudiotest(m audiotest.Measurement) measurement {
	return measurement{
		Frames:     m.Frames,
		SampleRate: m.SampleRate,
		RMSDb:      m.RMSDb,
		PeakDb:     m.PeakDb,
		CentroidHz: m.CentroidHz,
	}
}

func writeBaseline(path string, entries []baselineEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func readBaseline(path string) ([]baselineEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []baselineEntry
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type driftReport struct {
	Name   string
	Metric string
	Want   float64
	Got    float64
}

func (d driftReport) String() string {
	return fmt.Sprintf("%s: %s drift %.3f → %.3f", d.Name, d.Metric, d.Want, d.Got)
}

func compareBaselines(base, current []baselineEntry) []driftReport {
	byName := map[string]baselineEntry{}
	for _, b := range base {
		byName[b.Name] = b
	}
	var drifts []driftReport
	for _, c := range current {
		b, ok := byName[c.Name]
		if !ok {
			continue
		}
		if math.Abs(c.Measurement.RMSDb-b.Measurement.RMSDb) > rmsToleranceDB {
			drifts = append(drifts, driftReport{c.Name, "rms_db", b.Measurement.RMSDb, c.Measurement.RMSDb})
		}
		if math.Abs(c.Measurement.PeakDb-b.Measurement.PeakDb) > peakToleranceDB {
			drifts = append(drifts, driftReport{c.Name, "peak_db", b.Measurement.PeakDb, c.Measurement.PeakDb})
		}
		if b.Measurement.CentroidHz > 0 {
			ratio := math.Abs(c.Measurement.CentroidHz-b.Measurement.CentroidHz) / b.Measurement.CentroidHz
			if ratio > centroidToleranceR {
				drifts = append(drifts, driftReport{c.Name, "centroid_hz", b.Measurement.CentroidHz, c.Measurement.CentroidHz})
			}
		}
	}
	return drifts
}
