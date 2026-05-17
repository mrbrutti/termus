package gen

func (a *Ambient) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *Ambient) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "pads", Channels: []int32{0, 1, 2}},
		{Name: "bells", Channels: []int32{3, 4}},
		{Name: "bass", Channels: []int32{5}},
	})
}

func (a *SF2Glass) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *SF2Glass) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "bells", Channels: []int32{0, 1, 2, 3}},
		{Name: "textures", Channels: []int32{4, 5, 6}},
		{Name: "bass", Channels: []int32{7}},
	})
}

func (a *SF2Drone) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *SF2Drone) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "drones", Channels: []int32{0, 1, 2}},
		{Name: "shimmer", Channels: []int32{3}},
		{Name: "bass", Channels: []int32{4}},
	})
}

func (a *SF2Pentatonic) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *SF2Pentatonic) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "bass_harp", Channels: []int32{0}},
		{Name: "music_box", Channels: []int32{1}},
		{Name: "ornaments", Channels: []int32{2, 3}},
		{Name: "choir", Channels: []int32{4}},
	})
}

func (a *SF2Markov) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *SF2Markov) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "lead", Channels: []int32{0, 4}},
		{Name: "bass", Channels: []int32{1}},
		{Name: "comp", Channels: []int32{2}},
		{Name: "texture", Channels: []int32{3}},
	})
}

func (a *Phase) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *Phase) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "phase", Channels: []int32{0, 1}},
		{Name: "texture", Channels: []int32{2, 3}},
		{Name: "bass", Channels: []int32{4}},
		{Name: "sparkle", Channels: []int32{5}},
	})
}

func (a *Jazz) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *Jazz) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "piano", Channels: []int32{0}},
		{Name: "bass", Channels: []int32{1}},
		{Name: "sax", Channels: []int32{2}},
		{Name: "drums", Channels: []int32{drumChannel}},
	})
}

func (a *Chill) ExportMIDI(path string, seconds float64) error {
	return exportSF2MIDI(a, a.core, path, seconds)
}

func (a *Chill) ExportStems(dir string, seconds float64, volume int) ([]string, error) {
	return exportSF2Stems(a, a.sf, a.core, dir, seconds, volume, []stemDefinition{
		{Name: "ep", Channels: []int32{0}},
		{Name: "bass", Channels: []int32{1}},
		{Name: "vibes", Channels: []int32{2}},
		{Name: "sax", Channels: []int32{3}},
		{Name: "guitar", Channels: []int32{4}},
		{Name: "drums", Channels: []int32{drumChannel}},
	})
}
