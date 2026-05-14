package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// ReadIR reads an IR (impulse response) WAV file and returns it as a single
// mono float64 slice in [-1, 1]. Mono input is returned as-is; stereo and
// multichannel files are downmixed (mean of channels).
//
// Only 16-bit PCM WAV is supported. This is sufficient for the Voxengo and
// similar free IR libraries which all distribute in this format.
func ReadIR(path string) ([]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hdr := make([]byte, 12)
	if _, err := io.ReadFull(f, hdr); err != nil {
		return nil, fmt.Errorf("wav %q: read RIFF header: %w", path, err)
	}
	if string(hdr[0:4]) != "RIFF" || string(hdr[8:12]) != "WAVE" {
		return nil, fmt.Errorf("wav %q: not a RIFF/WAVE file", path)
	}

	var (
		numChannels   uint16
		bitsPerSample uint16
		dataBytes     uint32
		audioFormat   uint16
	)
	// Walk chunks until we find both "fmt " and "data".
	chunkHdr := make([]byte, 8)
	var sawFmt, sawData bool
	var dataStart int64
	for !sawData {
		if _, err := io.ReadFull(f, chunkHdr); err != nil {
			return nil, fmt.Errorf("wav %q: read chunk header: %w", path, err)
		}
		chunkID := string(chunkHdr[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHdr[4:8])
		switch chunkID {
		case "fmt ":
			fmtBuf := make([]byte, chunkSize)
			if _, err := io.ReadFull(f, fmtBuf); err != nil {
				return nil, fmt.Errorf("wav %q: read fmt chunk: %w", path, err)
			}
			if len(fmtBuf) < 16 {
				return nil, fmt.Errorf("wav %q: short fmt chunk", path)
			}
			audioFormat = binary.LittleEndian.Uint16(fmtBuf[0:2])
			numChannels = binary.LittleEndian.Uint16(fmtBuf[2:4])
			bitsPerSample = binary.LittleEndian.Uint16(fmtBuf[14:16])
			sawFmt = true
		case "data":
			dataBytes = chunkSize
			pos, _ := f.Seek(0, io.SeekCurrent)
			dataStart = pos
			sawData = true
		default:
			if _, err := f.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("wav %q: skip chunk %q: %w", path, chunkID, err)
			}
		}
	}
	if !sawFmt {
		return nil, fmt.Errorf("wav %q: missing fmt chunk", path)
	}
	if audioFormat != 1 {
		return nil, fmt.Errorf("wav %q: audioFormat=%d, only PCM (1) supported", path, audioFormat)
	}
	if bitsPerSample != 16 {
		return nil, fmt.Errorf("wav %q: bitsPerSample=%d, only 16-bit supported", path, bitsPerSample)
	}
	if numChannels < 1 {
		return nil, fmt.Errorf("wav %q: numChannels=%d", path, numChannels)
	}
	if _, err := f.Seek(dataStart, io.SeekStart); err != nil {
		return nil, err
	}

	// Read all 16-bit samples and downmix to mono.
	bytesPerFrame := int(numChannels) * 2
	numFrames := int(dataBytes) / bytesPerFrame
	out := make([]float64, numFrames)
	rawFrame := make([]byte, bytesPerFrame)
	for i := 0; i < numFrames; i++ {
		if _, err := io.ReadFull(f, rawFrame); err != nil {
			return nil, fmt.Errorf("wav %q: read sample %d: %w", path, i, err)
		}
		var sum float64
		for c := 0; c < int(numChannels); c++ {
			s := int16(binary.LittleEndian.Uint16(rawFrame[c*2 : c*2+2]))
			sum += float64(s) / 32768.0
		}
		out[i] = sum / float64(numChannels)
	}
	return out, nil
}
