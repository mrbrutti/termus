// Package sf2 manages SoundFont (.sf2) files: locating one on disk, falling
// back to auto-download of TimGM6mb.sf2 (6 MB, GPL-2, complete GM bank) into
// the user's cache directory, and opening it as a *meltysynth.SoundFont.
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

// DefaultURL is the GitHub-hosted copy of TimGM6mb.sf2 by Tim Brechbill.
const DefaultURL = "https://github.com/arbruijn/TimGM6mb/raw/master/TimGM6mb.sf2"

// DefaultSHA256 is the expected hash of the file at DefaultURL. Used to
// verify downloads; if the URL ever serves a different file we reject it.
const DefaultSHA256 = "c5378b62028c920cb11e4803327983fee2f2cdff5dc89c708e39da417e51c854"

// DefaultFileName is what the cached file is called on disk.
const DefaultFileName = "TimGM6mb.sf2"

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

// EnsureDefault returns the path to a usable SF2 file, downloading the default
// one if no cached copy is present. If progress is non-nil it's called with
// (bytesDownloaded, totalBytes) updates during a download.
func EnsureDefault(progress func(done, total int64)) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, DefaultFileName)

	if _, err := os.Stat(path); err == nil {
		// File exists. Verify hash; if mismatched, re-download.
		ok, err := verifyHash(path, DefaultSHA256)
		if err != nil {
			return "", fmt.Errorf("verify cached sf2: %w", err)
		}
		if ok {
			return path, nil
		}
		// Hash mismatch: stale or corrupt. Re-download.
		_ = os.Remove(path)
	}

	if err := downloadTo(DefaultURL, path, progress); err != nil {
		return "", err
	}
	ok, err := verifyHash(path, DefaultSHA256)
	if err != nil {
		return "", fmt.Errorf("verify downloaded sf2: %w", err)
	}
	if !ok {
		_ = os.Remove(path)
		return "", errors.New("downloaded sf2 failed checksum — refusing to use")
	}
	return path, nil
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
	client := &http.Client{Timeout: 5 * time.Minute}
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
	buf := make([]byte, 32*1024)
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
