package logger

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RotatingFileWriter is an io.Writer that writes to a single active log file and
// keeps exactly one previous file as backup. It rotates when any of these
// happens first:
//
//   - the active file would exceed maxSizeBytes (0 disables the size limit)
//   - the active file is older than rotateAfter (0 disables the age limit)
//
// On rotation the previous backup is deleted and the active file is renamed to
// the backup, so disk usage stays bounded to roughly twice the configured size.
// This keeps the most recent logs without needing an unbounded number of files,
// which matters for resource-constrained deployments.
//
// The age of the active file is taken from its modification time when it is
// first opened and then tracked in memory. This means a process restart resets
// the age clock to the last write time of the file, while the size limit still
// guarantees that logs never grow without bound.
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

	backup := w.path + ".1"
	if err := os.Remove(backup); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(w.path, backup); err != nil && !os.IsNotExist(err) {
		return err
	}

	return w.open()
}

func (w *RotatingFileWriter) close() error {
	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}
