package logger

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// backupTimeLayout is the timestamp format embedded in rotated file names. It
// is also what deleteExpired parses back to decide whether a file is stale.
const backupTimeLayout = "20060102-150405"

// RotatingFileWriter is an io.Writer that writes to a single active log file and
// rotates it into timestamped backups. It rotates when any of these happens
// first:
//
//   - the active file would exceed maxSizeBytes (0 disables the size limit)
//   - the active file is older than rotateAfter (0 disables the age limit)
//
// rotateAfter is also the retention window: on every rotation, backups whose
// rotation timestamp is older than now-rotateAfter are deleted. This keeps the
// logs on disk bounded to the retention window instead of growing forever.
//
// The age of the active file is taken from its modification time when it is
// first opened and then tracked in memory. A process restart therefore resets
// the age clock to the last write time of the file, while the size limit still
// bounds the active file.
type RotatingFileWriter struct {
	path         string
	maxSizeBytes int64
	rotateAfter  time.Duration

	// now allows tests to control the clock. It defaults to time.Now.
	now func() time.Time

	mu       sync.Mutex
	file     *os.File
	size     int64
	openedAt time.Time
}

// NewRotatingFileWriter creates a writer for path. The parent directory is
// created on demand. A maxSizeBytes of 0 and a rotateAfter of 0 keep the single
// file unchanged.
func NewRotatingFileWriter(path string, maxSizeBytes int64, rotateAfter time.Duration) *RotatingFileWriter {
	return &RotatingFileWriter{
		path:         path,
		maxSizeBytes: maxSizeBytes,
		rotateAfter:  rotateAfter,
		now:          time.Now,
	}
}

// Write implements io.Writer.
func (w *RotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}

	if w.shouldRotate(len(p)) {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Close flushes and closes the active file. The writer can be used again after
// Close; the file is reopened on the next Write.
func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.close()
}

func (w *RotatingFileWriter) open() error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}

	w.file = f
	w.size = info.Size()

	// A brand new (empty) file starts its clock now. An existing file keeps the
	// time of its last write, so a process that starts after a long idle period
	// rotates immediately instead of reusing a stale file.
	if info.Size() == 0 {
		w.openedAt = w.now()
	} else {
		w.openedAt = info.ModTime()
	}

	return nil
}

func (w *RotatingFileWriter) shouldRotate(incoming int) bool {
	if w.maxSizeBytes > 0 && w.size+int64(incoming) > w.maxSizeBytes {
		return true
	}
	if w.rotateAfter > 0 && w.now().Sub(w.openedAt) >= w.rotateAfter {
		return true
	}
	return false
}

func (w *RotatingFileWriter) rotate() error {
	if err := w.close(); err != nil {
		return err
	}

	// Clean up expired backups before adding a new one, so the retention window
	// is enforced even if the process runs for a long time.
	if err := w.deleteExpired(); err != nil {
		return err
	}

	backup := w.nextBackupPath()
	if err := os.Rename(w.path, backup); err != nil && !os.IsNotExist(err) {
		return err
	}

	return w.open()
}

// nextBackupPath returns a unique timestamped name for the rotated file.
func (w *RotatingFileWriter) nextBackupPath() string {
	base := w.path + "." + w.now().UTC().Format(backupTimeLayout)

	candidate := base
	for i := 1; ; i++ {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
		candidate = base + "-" + strconv.Itoa(i)
	}
}

// deleteExpired removes rotated files whose rotation timestamp is older than the
// retention window.
func (w *RotatingFileWriter) deleteExpired() error {
	if w.rotateAfter <= 0 {
		return nil
	}

	cutoff := w.now().Add(-w.rotateAfter)

	matches, err := filepath.Glob(w.path + ".*")
	if err != nil {
		return err
	}

	for _, match := range matches {
		ts, ok := backupTimestamp(match, w.path)
		if !ok {
			continue
		}
		if ts.Before(cutoff) {
			if err := os.Remove(match); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}

	return nil
}

// backupTimestamp extracts the rotation time encoded in a rotated file name.
func backupTimestamp(path, activePath string) (time.Time, bool) {
	name := strings.TrimPrefix(path, activePath+".")
	if len(name) < len(backupTimeLayout) {
		return time.Time{}, false
	}

	ts, err := time.ParseInLocation(backupTimeLayout, name[:len(backupTimeLayout)], time.UTC)
	if err != nil {
		return time.Time{}, false
	}
	return ts, true
}

func (w *RotatingFileWriter) close() error {
	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}
