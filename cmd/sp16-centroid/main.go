// sp16-centroid is a quick CLI that reports the average spectral centroid
// of a directory of WAV files. Used to verify genre ordering after SP16:
// ambient < lofi < chill < jazz.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mrbrutti/termus/internal/audiotest"
)

func main() {
	root := flag.String("root", "wavs/sp16", "root directory containing genre subdirs")
	flag.Parse()
	dirs, err := os.ReadDir(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	results := map[string]float64{}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		genre := d.Name()
		entries, err := os.ReadDir(filepath.Join(*root, genre))
		if err != nil {
			continue
		}
		var centroids []float64
		for _, f := range entries {
			if filepath.Ext(f.Name()) != ".wav" {
				continue
			}
			samples, sr, err := readWAV(filepath.Join(*root, genre, f.Name()))
			if err != nil {
				continue
			}
			c := audiotest.SpectralCentroidHz(samples, sr)
			centroids = append(centroids, c)
		}
		if len(centroids) == 0 {
			continue
		}
		sum := 0.0
		for _, c := range centroids {
			sum += c
		}
		results[genre] = sum / float64(len(centroids))
	}
	type entry struct {
		Genre    string
		Centroid float64
	}
	list := []entry{}
	for g, c := range results {
		list = append(list, entry{Genre: g, Centroid: c})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Centroid < list[j].Centroid
	})
	for _, e := range list {
		fmt.Printf("%-12s  %8.1f Hz\n", e.Genre, e.Centroid)
	}
}

// readWAV reads a 16-bit PCM stereo or mono WAV file and returns the samples
// as float64 in [-1, 1] (mono mix-down if stereo).
func readWAV(path string) ([]float64, float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	if len(data) < 44 {
		return nil, 0, fmt.Errorf("file too short")
	}
	channels := binary.LittleEndian.Uint16(data[22:24])
	sampleRate := binary.LittleEndian.Uint32(data[24:28])
	bitsPerSample := binary.LittleEndian.Uint16(data[34:36])
	if bitsPerSample != 16 {
		return nil, 0, fmt.Errorf("unsupported bits per sample: %d", bitsPerSample)
	}
	// Find "data" chunk.
	idx := 36
	for idx+8 <= len(data) {
		chunkID := string(data[idx : idx+4])
		chunkSize := binary.LittleEndian.Uint32(data[idx+4 : idx+8])
		if chunkID == "data" {
			idx += 8
			payload := data[idx : idx+int(chunkSize)]
			samples := readSamples(payload, int(channels))
			return samples, float64(sampleRate), nil
		}
		idx += 8 + int(chunkSize)
	}
	return nil, 0, fmt.Errorf("data chunk not found")
}

func readSamples(b []byte, channels int) []float64 {
	nFrames := len(b) / (2 * channels)
	out := make([]float64, nFrames)
	for i := 0; i < nFrames; i++ {
		var sum int32
		for c := 0; c < channels; c++ {
			off := i*2*channels + c*2
			v := int16(binary.LittleEndian.Uint16(b[off : off+2]))
			sum += int32(v)
		}
		avg := float64(sum) / float64(channels)
		out[i] = avg / 32768.0
	}
	return out
}
