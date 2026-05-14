package sf2

import (
	"testing"
)

func TestDefaultConstants(t *testing.T) {
	if DefaultURL == "" {
		t.Fatal("DefaultURL is empty")
	}
	if len(DefaultSHA256) != 64 {
		t.Fatalf("DefaultSHA256 should be 64 hex chars, got %d", len(DefaultSHA256))
	}
	if DefaultFileName == "" {
		t.Fatal("DefaultFileName is empty")
	}
}

func TestCacheDirCreates(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, err := CacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir == "" {
		t.Fatal("empty dir")
	}
}
