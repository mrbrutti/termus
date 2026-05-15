// Package sf2 manages SoundFont (.sf2) files: locating one on disk, falling
// back to auto-download of a curated preset into the user's cache directory,
// and opening it as a *meltysynth.SoundFont.
package sf2

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

// Preset describes one of the curated SoundFont download options. The user
// picks a preset via --sf2-preset (or accepts the default "general"); the
// engine downloads the corresponding file on first use, verifies its hash,
// and caches it.
type Preset struct {
	Name     string
	URL      string
	SHA256   string
	FileName string
	SizeMB   int    // approximate, for progress messages
	Summary  string // one-liner shown to the user
}

// Presets — the curated SF2 options. Order matters: "general" is the default
// shown first in --help.
var Presets = map[string]Preset{
	"general": {
		Name:     "general",
		URL:      "https://github.com/mrbumpy409/GeneralUser-GS/raw/main/GeneralUser-GS.sf2",
		SHA256:   "9575028c7a1f589f5770fccc8cff2734566af40cd26ed836944e9a5152688cfe",
		FileName: "GeneralUser-GS.sf2",
		SizeMB:   32,
		Summary:  "GeneralUser-GS by S. Christian Collins — MIT, 261 instruments + 13 drum kits, balanced quality",
	},
	"sgm": {
		Name:     "sgm",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/SGM-v2.01-NicePianosGuitarsBass-V1.2.sf2",
		SHA256:   "1b999795b8006c323e490596eb8c41acc0c50598748cb8bee9e71d75b54a6088",
		FileName: "SGM-v2.01-NicePianosGuitarsBass.sf2",
		SizeMB:   325,
		Summary:  "SGM v2.01 NicePianosGuitarsBass — substantially better piano/guitar/bass samples; best for chill/pentatonic/markov; ~10× the download",
	},
}

// DefaultPreset is the preset used when --sf2-preset is unspecified.
const DefaultPreset = "general"

// DefaultURL is retained as a compatibility alias for callers that grabbed
// it directly; new code should use Presets[DefaultPreset].URL.
const DefaultURL = "https://github.com/mrbumpy409/GeneralUser-GS/raw/main/GeneralUser-GS.sf2"

// DefaultSHA256 is the expected hash for DefaultURL.
const DefaultSHA256 = "9575028c7a1f589f5770fccc8cff2734566af40cd26ed836944e9a5152688cfe"

// DefaultFileName is the name DefaultURL is cached under.
const DefaultFileName = "GeneralUser-GS.sf2"

// CacheDir returns the directory we use to cache downloaded soundfonts.
// Creates it if it doesn't exist.
func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "termus", "soundfonts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// EnsureDefault returns the path to the default-preset SF2, downloading it
// if no cached copy is present. Equivalent to EnsurePreset(DefaultPreset).
func EnsureDefault(progress func(done, total int64)) (string, error) {
	return EnsurePreset(DefaultPreset, progress)
}

// EnsurePreset returns the path to the named preset's SF2 file, downloading
// it on demand if not cached. progress receives byte-count updates during a
// download (nil to disable). Returns an error for unknown presets.
func EnsurePreset(presetName string, progress func(done, total int64)) (string, error) {
	p, ok := Presets[presetName]
	if !ok {
		return "", fmt.Errorf("unknown sf2 preset %q (available: %s)",
			presetName, presetNamesJoined())
	}
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, p.FileName)

	if _, err := os.Stat(path); err == nil {
		ok, err := verifyHash(path, p.SHA256)
		if err != nil {
			return "", fmt.Errorf("verify cached sf2: %w", err)
		}
		if ok {
			return path, nil
		}
		// Hash mismatch — stale or corrupt, redownload.
		_ = os.Remove(path)
	}

	if err := downloadTo(p.URL, path, progress); err != nil {
		return "", err
	}
	ok2, err := verifyHash(path, p.SHA256)
	if err != nil {
		return "", fmt.Errorf("verify downloaded sf2: %w", err)
	}
	if !ok2 {
		_ = os.Remove(path)
		return "", errors.New("downloaded sf2 failed checksum — refusing to use")
	}
	return path, nil
}

func presetNamesJoined() string {
	out := ""
	for k := range Presets {
		if out != "" {
			out += ", "
		}
		out += k
	}
	return out
}

// Open loads an SF2 file from disk into a meltysynth SoundFont.
func Open(path string) (*meltysynth.SoundFont, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sf, err := meltysynth.NewSoundFont(f)
	if err != nil {
		return nil, fmt.Errorf("parse sf2 %q: %w", path, err)
	}
	return sf, nil
}

func verifyHash(path, want string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	got := hex.EncodeToString(h.Sum(nil))
	return got == want, nil
}

func downloadTo(url, dst string, progress func(done, total int64)) error {
	client := &http.Client{Timeout: 15 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download %q: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %q: HTTP %d", url, resp.StatusCode)
	}

	tmp := dst + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	total := resp.ContentLength
	var done int64
	buf := make([]byte, 64*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				_ = f.Close()
				_ = os.Remove(tmp)
				return werr
			}
			done += int64(n)
			if progress != nil {
				progress(done, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return rerr
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
