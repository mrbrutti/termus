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
		Summary:  "SGM v2.01 NicePianosGuitarsBass — substantially better piano/guitar/bass; best for piano-led genres",
	},
	"tyros4": {
		Name:     "tyros4",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/yamaha%20tyros%204_just_t4_fixed.sf2",
		SHA256:   "de5b1404630840a2a897e77c6083e2231c5d1523e6571a8afd972a4f684d36ce",
		FileName: "yamaha tyros 4_just_t4_fixed.sf2",
		SizeMB:   502,
		Summary:  "Yamaha Tyros 4 workstation rip (\"just_t4_fixed\") — best for jazz brass / sax sections / show-band reeds and rich GS strings",
	},
	"dsound4": {
		Name:     "dsound4",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/DSoundFontV4.sf2",
		SHA256:   "01815e87817138cc30e798152d3cc66991f853305eac34660bb588437c09fee4",
		FileName: "DSoundFontV4.sf2",
		SizeMB:   553,
		Summary:  "DSoundFont V4 — large balanced GM/GS bank, drop-in alternative when GeneralUser feels thin",
	},
	"fatboy": {
		Name:     "fatboy",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/FatBoy-v0.786.sf2",
		SHA256:   "8aac3471ea8873b6526325918fff5c83008dccecdc89ecc88d32a7205289307e",
		FileName: "FatBoy-v0.786.sf2",
		SizeMB:   315,
		Summary:  "FatBoy v0.786 — every instrument hand-metered to equal loudness; clean, MIT-style license; good for clean lo-fi and baroque",
	},
	// roland-sc55-up was researched and downloaded but is incompatible with
	// go-meltysynth — its "Halo Pad" preset has no zone, which makes the
	// parser reject the entire file. Removed from the catalog. If a Roland-
	// character alternative is wanted, `arachno` covers similar ground
	// (D-50/M1/MU/Fairlight blend) and parses cleanly.
	"timbres-of-heaven": {
		Name:     "timbres-of-heaven",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/Timbres%20Of%20Heaven%20GM_GS_XG_SFX%20V%203.4%20Final.sf2",
		SHA256:   "a94524fc660ce203dd9216dae008d75c1375c15701eaaa67b95f3aad1dea9384",
		FileName: "Timbres Of Heaven GM_GS_XG_SFX V 3.4 Final.sf2",
		SizeMB:   377,
		Summary:  "Timbres of Heaven — orchestral/classical workhorse; rich strings, brass, woodwinds; best for classical",
	},
	"merlin-symphony": {
		Name:     "merlin-symphony",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/merlin_symphony%28v1.21%29.sf2",
		SHA256:   "521a953c5983de7bd2997b08ca5fe69afcf010d9358b5675f8fc8932b897f95e",
		FileName: "merlin_symphony(v1.21).sf2",
		SizeMB:   163,
		Summary:  "Merlin Symphony v1.21 — strong orchestral strings/winds; alt classical bank (drums are weak)",
	},
	"fairy-tale": {
		Name:     "fairy-tale",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/The_Fairy_Tale_Bank.sf2",
		SHA256:   "8a331057ff24b6238717533dfefe1fa58faf1f093468789e36dfe640a00dccd2",
		FileName: "The_Fairy_Tale_Bank.sf2",
		SizeMB:   200,
		Summary:  "The Fairy Tale Bank — celesta, music-box, glockenspiel + bells; CC-BY-NC-SA — best for bells / lullaby aesthetics",
	},
	"fm-dx": {
		Name:     "fm-dx",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/FMSynthesis1.40.sf2",
		SHA256:   "26c4d802da653782c62dd398c5c4f57aa24f75233f93e6a7a828583fa84f939b",
		FileName: "FMSynthesis1.40.sf2",
		SizeMB:   124,
		Summary:  "FM Synthesis v1.40 — DX-style FM EPs, metallic bells, leads, basses — '80s tones, perfect for phase/drone textures",
	},
	"musescore-general": {
		Name:     "musescore-general",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/MuseScore_General%28v0.1.3%29.sf2",
		SHA256:   "8520f85bd115d51be327736584fd2b0ccced1ec786636fc2139efea2d714a5b4",
		FileName: "MuseScore_General(v0.1.3).sf2",
		SizeMB:   208,
		Summary:  "MuseScore General v0.1.3 — MIT-licensed, balanced, polite GM rendering; safe legal status",
	},
	"arachno": {
		Name:     "arachno",
		URL:      "https://archive.org/download/free-soundfonts-sf2-2019-04/Arachno_SoundFont_Version_1.0.sf2",
		SHA256:   "9a57fb3b6714e69dda12390e351b087e81fc3b1eca15c6b4bbe172799f4cf3cd",
		FileName: "Arachno_SoundFont_Version_1.0.sf2",
		SizeMB:   148,
		Summary:  "Arachno SoundFont — D-50/M1/MU/Fairlight blend; retro-game/'80s feel; great pads, bells, leads; CC-BY-NC-SA (non-commercial)",
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
