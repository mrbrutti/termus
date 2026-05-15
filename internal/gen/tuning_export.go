package gen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sinshu/go-meltysynth/meltysynth"

	"github.com/mrbrutti/termus/internal/synth"
)

type TuningExporter interface {
	ExportMIDI(path string, seconds float64) error
	ExportStems(dir string, seconds float64, volume int) ([]string, error)
}

type capturedMIDIMessage struct {
	Sample int64
	Status byte
	Data1  byte
	Data2  byte
}

type midiCapture struct {
	events []capturedMIDIMessage
}

type stemDefinition struct {
	Name     string
	Channels []int32
}

func exportSF2MIDI(algo Algorithm, core *sf2Core, path string, seconds float64) error {
	events, err := captureSF2Events(algo, core, seconds)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeCapturedMIDI(path, events)
}

func exportSF2Stems(algo Algorithm, sf *meltysynth.SoundFont, core *sf2Core, dir string, seconds float64, volume int, stems []stemDefinition) ([]string, error) {
	events, err := captureSF2Events(algo, core, seconds)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	gain := EffectiveOutputGain(algo) * float64(volume) / 100.0
	written := make([]string, 0, len(stems))
	for _, stem := range stems {
		path := filepath.Join(dir, stem.Name+".wav")
		if err := renderCapturedStem(path, sf, events, stem.Channels, seconds, gain); err != nil {
			return nil, err
		}
		written = append(written, path)
	}
	return written, nil
}

func captureSF2Events(algo Algorithm, core *sf2Core, seconds float64) ([]capturedMIDIMessage, error) {
	if core == nil {
		return nil, fmt.Errorf("sf2 core unavailable")
	}
	frames := int(seconds * float64(synth.SampleRate))
	if frames <= 0 {
		return nil, fmt.Errorf("seconds must be > 0")
	}
	left := make([]float64, 512)
	right := make([]float64, 512)
	core.startMIDICapture()
	for rendered := 0; rendered < frames; {
		n := len(left)
		if rem := frames - rendered; rem < n {
			n = rem
		}
		algo.Next(left[:n], right[:n])
		rendered += n
	}
	return core.finishMIDICapture(), nil
}

func renderCapturedStem(path string, sf *meltysynth.SoundFont, events []capturedMIDIMessage, channels []int32, seconds float64, gain float64) error {
	settings := meltysynth.NewSynthesizerSettings(synth.SampleRate)
	settings.EnableReverbAndChorus = true
	settings.MaximumPolyphony = 96
	syn, err := meltysynth.NewSynthesizer(sf, settings)
	if err != nil {
		return err
	}
	totalFrames := int(seconds * float64(synth.SampleRate))
	if totalFrames <= 0 {
		return fmt.Errorf("seconds must be > 0")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := writeWAVHeader(f, 0); err != nil {
		return err
	}
	allowed := make(map[byte]bool, len(channels))
	for _, ch := range channels {
		allowed[byte(ch)] = true
	}
	bufL := make([]float32, 512)
	bufR := make([]float32, 512)
	frames := make([][2]float64, 512)
	pos := 0
	eventIdx := 0
	dataBytes := uint32(0)
	for pos < totalFrames {
		for eventIdx < len(events) && int(events[eventIdx].Sample) <= pos {
			dispatchCapturedMessage(syn, events[eventIdx], allowed)
			eventIdx++
		}
		nextEvent := totalFrames
		if eventIdx < len(events) {
			nextEvent = int(events[eventIdx].Sample)
			if nextEvent < pos {
				nextEvent = pos
			}
		}
		block := len(bufL)
		if rem := totalFrames - pos; rem < block {
			block = rem
		}
		if nextEvent > pos && nextEvent-pos < block {
			block = nextEvent - pos
		}
		if block <= 0 {
			block = 1
		}
		syn.Render(bufL[:block], bufR[:block])
		for i := 0; i < block; i++ {
			frames[i][0] = float64(bufL[i]) * gain
			frames[i][1] = float64(bufR[i]) * gain
		}
		n, err := writeWAVFrames(f, frames[:block])
		dataBytes += uint32(n)
		if err != nil {
			return err
		}
		pos += block
	}
	return patchWAVHeader(f, dataBytes)
}

func dispatchCapturedMessage(syn *meltysynth.Synthesizer, msg capturedMIDIMessage, allowed map[byte]bool) {
	channel := msg.Status & 0x0F
	if !allowed[channel] {
		return
	}
	switch msg.Status & 0xF0 {
	case 0x80:
		syn.NoteOff(int32(channel), int32(msg.Data1))
	case 0x90:
		if msg.Data2 == 0 {
			syn.NoteOff(int32(channel), int32(msg.Data1))
		} else {
			syn.NoteOn(int32(channel), int32(msg.Data1), int32(msg.Data2))
		}
	default:
		syn.ProcessMidiMessage(int32(channel), int32(msg.Status&0xF0), int32(msg.Data1), int32(msg.Data2))
	}
}

func writeCapturedMIDI(path string, events []capturedMIDIMessage) error {
	var track bytes.Buffer
	lastTick := int64(0)
	for _, ev := range events {
		tick := (ev.Sample * 1000) / int64(synth.SampleRate)
		if tick < lastTick {
			tick = lastTick
		}
		writeVarLen(&track, tick-lastTick)
		track.WriteByte(ev.Status)
		track.WriteByte(ev.Data1)
		switch ev.Status & 0xF0 {
		case 0xC0, 0xD0:
			// one-data-byte status
		default:
			track.WriteByte(ev.Data2)
		}
		lastTick = tick
	}
	writeVarLen(&track, 0)
	track.Write([]byte{0xFF, 0x2F, 0x00})

	var out bytes.Buffer
	out.Write([]byte("MThd"))
	_ = binary.Write(&out, binary.BigEndian, uint32(6))
	_ = binary.Write(&out, binary.BigEndian, uint16(0))
	_ = binary.Write(&out, binary.BigEndian, uint16(1))
	_ = binary.Write(&out, binary.BigEndian, uint16(0xE728)) // 25 fps, 40 ticks/frame => 1000 ticks/sec
	out.Write([]byte("MTrk"))
	_ = binary.Write(&out, binary.BigEndian, uint32(track.Len()))
	out.Write(track.Bytes())

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out.Bytes(), 0o644)
}

func writeVarLen(buf *bytes.Buffer, v int64) {
	if v < 0 {
		v = 0
	}
	var tmp [10]byte
	i := len(tmp) - 1
	tmp[i] = byte(v & 0x7F)
	for v >>= 7; v > 0; v >>= 7 {
		i--
		tmp[i] = byte(v&0x7F) | 0x80
	}
	buf.Write(tmp[i:])
}

func writeWAVHeader(f *os.File, dataBytes uint32) error {
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	byteRate := uint32(synth.SampleRate * 2 * 2)
	blockAlign := uint16(4)
	buf := make([]byte, 0, 44)
	buf = append(buf, []byte("RIFF")...)
	buf = binary.LittleEndian.AppendUint32(buf, 36+dataBytes)
	buf = append(buf, []byte("WAVEfmt ")...)
	buf = binary.LittleEndian.AppendUint32(buf, 16)
	buf = binary.LittleEndian.AppendUint16(buf, 1)
	buf = binary.LittleEndian.AppendUint16(buf, 2)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(synth.SampleRate))
	buf = binary.LittleEndian.AppendUint32(buf, byteRate)
	buf = binary.LittleEndian.AppendUint16(buf, blockAlign)
	buf = binary.LittleEndian.AppendUint16(buf, 16)
	buf = append(buf, []byte("data")...)
	buf = binary.LittleEndian.AppendUint32(buf, dataBytes)
	_, err := f.Write(buf)
	return err
}

func patchWAVHeader(f *os.File, dataBytes uint32) error {
	if err := writeWAVHeader(f, dataBytes); err != nil {
		return err
	}
	return f.Close()
}

func writeWAVFrames(f *os.File, frames [][2]float64) (int, error) {
	out := make([]byte, len(frames)*4)
	o := 0
	for _, fr := range frames {
		for _, v := range fr {
			if v > 1 {
				v = 1
			}
			if v < -1 {
				v = -1
			}
			s := int16(v * 32767)
			out[o] = byte(s)
			out[o+1] = byte(s >> 8)
			o += 2
		}
	}
	return f.Write(out)
}
