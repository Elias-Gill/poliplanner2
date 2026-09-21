package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogFileName is the name of the active log file inside Options.LogDir. Rotated
// files are stored next to it with a timestamp suffix.
const LogFileName = "poliplanner.log"

// Options configures the application logger. The zero value logs INFO to
// stdout as text, with no file output, which is the behavior used by tests and
// tooling that never call InitLogger explicitly.
type Options struct {
	// Level is the minimum level to emit. The zero value is LevelInfo.
	Level slog.Level

	// LogDir is the directory where LogFileName is written. When empty, logs
	// are written only to stdout. The directory is created on demand.
	LogDir string

	// MaxSizeBytes rotates the file once it would grow past this size. Zero
	// disables the size limit.
	MaxSizeBytes int64

	// RotateAfter rotates the file once it is older than this duration and also
	// bounds how long rotated files are kept. Zero disables both.
	RotateAfter time.Duration

	// Format selects the handler. Use "json" for JSON output; anything else
	// (including the empty value) produces text.
	Format string

	// Location renders timestamps. When nil, time.Local is used.
	Location *time.Location
}

// custom_logger is initialized with a sane default so logging never depends on
// an unsynchronized lazy initialization. InitLogger replaces it once, from the
// composition root, before any concurrent goroutine is started.
var custom_logger = slog.New(newHandler(os.Stdout, slog.LevelInfo, "text", nil))

// InitLogger updates the configuration of the default logger. It should be
// called after loading the application configuration.
func InitLogger(opts Options) {
	writers := []io.Writer{os.Stdout}
	if opts.LogDir != "" {
		writers = append(writers, NewRotatingFileWriter(
			filepath.Join(opts.LogDir, LogFileName),
			opts.MaxSizeBytes,
			opts.RotateAfter,
		))
	}

	custom_logger = slog.New(newHandler(
		io.MultiWriter(writers...),
		opts.Level,
		opts.Format,
		opts.Location,
	))
}

func newHandler(w io.Writer, level slog.Level, format string, loc *time.Location) slog.Handler {
	handlerOpts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if loc != nil && a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
				return slog.Time(slog.TimeKey, a.Value.Time().In(loc))
			}
			return a
		},
	}

	if strings.EqualFold(format, "json") {
		return slog.NewJSONHandler(w, handlerOpts)
	}
	return slog.NewTextHandler(w, handlerOpts)
}

func getLogger() *slog.Logger {
	return custom_logger
}

func Debug(msg string, args ...any) {
	getLogger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	getLogger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	getLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	getLogger().Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	getLogger().Error(msg, args...)
	os.Exit(1)
}
