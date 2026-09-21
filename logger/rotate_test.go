package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s not to exist, got err=%v", path, err)
	}
}

func TestRotatingFileWriter_AppendsAndCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "poliplanner.log")

	w := NewRotatingFileWriter(path, 0, 0)
	if _, err := w.Write([]byte("first\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := w.Write([]byte("second\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	if got := readFile(t, path); got != "first\nsecond\n" {
		t.Errorf("content = %q, want %q", got, "first\nsecond\n")
	}
}

func TestRotatingFileWriter_RotatesBySize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "poliplanner.log")

	w := NewRotatingFileWriter(path, 10, 0)
	if _, err := w.Write([]byte("0123456789")); err != nil {
		t.Fatalf("write: %v", err)
	}
	// This write would push the file past maxSizeBytes, so it rotates first.
	if _, err := w.Write([]byte("a")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	if got := readFile(t, path); got != "a" {
		t.Errorf("active content = %q, want %q", got, "a")
	}
	if got := readFile(t, path+".1"); got != "0123456789" {
		t.Errorf("backup content = %q, want %q", got, "0123456789")
	}
}

func TestRotatingFileWriter_RotatesByAge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "poliplanner.log")

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	w := NewRotatingFileWriter(path, 0, 15*24*time.Hour)
	w.now = func() time.Time { return now }

	if _, err := w.Write([]byte("day one\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	now = now.Add(16 * 24 * time.Hour)
	if _, err := w.Write([]byte("day sixteen\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	if got := readFile(t, path); got != "day sixteen\n" {
		t.Errorf("active content = %q, want %q", got, "day sixteen\n")
	}
	if got := readFile(t, path+".1"); got != "day one\n" {
		t.Errorf("backup content = %q, want %q", got, "day one\n")
	}
}

func TestRotatingFileWriter_KeepsOnlyTwoFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "poliplanner.log")

	// Each 10-byte block triggers a rotation, so after three writes there have
	// been two rotations but only the last two files should remain.
	w := NewRotatingFileWriter(path, 10, 0)
	for _, chunk := range []string{"aaaaaaaaaa", "bbbbbbbbbb", "cccccccccc"} {
		if _, err := w.Write([]byte(chunk)); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	mustExist(t, path)
	mustExist(t, path+".1")
	mustNotExist(t, path+".2")

	if got := readFile(t, path); got != "cccccccccc" {
		t.Errorf("active content = %q, want %q", got, "cccccccccc")
	}
	if got := readFile(t, path+".1"); got != "bbbbbbbbbb" {
		t.Errorf("backup content = %q, want %q", got, "bbbbbbbbbb")
	}
}

func TestRotatingFileWriter_RotatesStaleFileOnOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "poliplanner.log")

	if err := os.WriteFile(path, []byte("stale\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	old := time.Now().Add(-20 * 24 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	w := NewRotatingFileWriter(path, 0, 15*24*time.Hour)
	if _, err := w.Write([]byte("fresh\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	if got := readFile(t, path); got != "fresh\n" {
		t.Errorf("active content = %q, want %q", got, "fresh\n")
	}
	if got := readFile(t, path+".1"); got != "stale\n" {
		t.Errorf("backup content = %q, want %q", got, "stale\n")
	}
}
