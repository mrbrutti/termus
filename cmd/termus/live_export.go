package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mrbrutti/termus/internal/audio"
	"github.com/mrbrutti/termus/internal/gen"
	"github.com/mrbrutti/termus/internal/tui"
)

const tuiExportSeconds = 60.0

func makeTUIExporter(build func(gen.AlgoSpec, int64) gen.Algorithm, volume int) *tui.ExportController {
	return &tui.ExportController{
		Seconds: tuiExportSeconds,
		WAV: func(spec gen.AlgoSpec, seed int64) (string, error) {
			path := liveExportBase(spec, seed) + ".wav"
			algo := build(spec, seed)
			plan := audio.PlanRender(algo, tuiExportSeconds)
			_, err := audio.RenderToWAVWithPlan(path, algo, plan, volume)
			return path, err
		},
		MIDI: func(spec gen.AlgoSpec, seed int64) (string, error) {
			path := liveExportBase(spec, seed) + ".mid"
			algo := build(spec, seed)
			plan := audio.PlanRender(algo, tuiExportSeconds)
			exporter, ok := algo.(gen.TuningExporter)
			if !ok {
				return "", fmt.Errorf("algorithm does not support MIDI export")
			}
			return path, exporter.ExportMIDI(path, plan.DurationSeconds())
		},
		Stems: func(spec gen.AlgoSpec, seed int64) (string, error) {
			dir := liveExportBase(spec, seed) + "-stems"
			algo := build(spec, seed)
			plan := audio.PlanRender(algo, tuiExportSeconds)
			exporter, ok := algo.(gen.TuningExporter)
			if !ok {
				return "", fmt.Errorf("algorithm does not support stem export")
			}
			_, err := exporter.ExportStems(dir, plan.DurationSeconds(), volume)
			return dir, err
		},
	}
}

func liveExportBase(spec gen.AlgoSpec, seed int64) string {
	stamp := time.Now().Format("20060102-150405")
	file := fmt.Sprintf("%s-seed%d-%s", spec.Name, seed, stamp)
	return filepath.Join("exports", file)
}
