package audio

import (
	"encoding/binary"
	"fmt"
	"os"
)

// WAVWriter writes int16 PCM samples to a WAV file. The header is patched
// with the final data size on Close.
type WAVWriter struct {
	f          *os.File
	sampleRate int
	channels   int
	dataBytes  uint32
}

// NewWAVWriter creates a new WAV file with a placeholder header.
func NewWAVWriter(path string, sampleRate, channels int) (*WAVWriter, error) {
	if channels < 1 || channels > 2 {
		return nil, fmt.Errorf("channels must be 1 or 2, got %d", channels)
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := &WAVWriter{f: f, sampleRate: sampleRate, channels: channels}
	if err := w.writeHeader(0); err != nil {
		_ = f.Close()
		return nil, err
	}
	return w, nil
}

func (w *WAVWriter) writeHeader(dataBytes uint32) error {
	if _, err := w.f.Seek(0, 0); err != nil {
		return err
	}
	byteRate := uint32(w.sampleRate * w.channels * 2)
	blockAlign := uint16(w.channels * 2)
	buf := make([]byte, 0, 44)
	buf = append(buf, []byte("RIFF")...)
	buf = binary.LittleEndian.AppendUint32(buf, 36+dataBytes)
	buf = append(buf, []byte("WAVEfmt ")...)
	buf = binary.LittleEndian.AppendUint32(buf, 16)         // fmt chunk size
	buf = binary.LittleEndian.AppendUint16(buf, 1)          // PCM
	buf = binary.LittleEndian.AppendUint16(buf, uint16(w.channels))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(w.sampleRate))
	buf = binary.LittleEndian.AppendUint32(buf, byteRate)
	buf = binary.LittleEndian.AppendUint16(buf, blockAlign)
	buf = binary.LittleEndian.AppendUint16(buf, 16)         // bits per sample
	buf = append(buf, []byte("data")...)
	buf = binary.LittleEndian.AppendUint32(buf, dataBytes)
	_, err := w.f.Write(buf)
	return err
}

// Write encodes float64 stereo frames as int16 little-endian PCM.
func (w *WAVWriter) Write(frames [][2]float64) error {
	out := make([]byte, len(frames)*w.channels*2)
	o := 0
	for _, fr := range frames {
		for c := 0; c < w.channels; c++ {
			v := fr[c]
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
	n, err := w.f.Write(out)
	w.dataBytes += uint32(n)
	return err
}

// Close patches the header with the final data size and closes the file.
func (w *WAVWriter) Close() error {
	if err := w.writeHeader(w.dataBytes); err != nil {
		_ = w.f.Close()
		return err
	}
	return w.f.Close()
}
