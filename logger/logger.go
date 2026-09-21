package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var custom_logger *slog.Logger

// LogFileName is the name of the active log file inside Options.LogDir. Its
// rotated backup is stored next to it with a ".1" suffix.
const LogFileName = "poliplanner.log"

// Options configures the application logger. The zero value logs INFO to
// stdout with no file output, which is the behavior used by tests and tooling
// that never call InitLogger explicitly.
type Options struct {
	// Verbose lowers the minimum level to DEBUG.
	Verbose bool

	// LogDir is the directory where LogFileName is written. When empty, logs
	// are written only to stdout. The directory is created on demand.
	LogDir string

	// MaxSizeBytes rotates the file once it would grow past this size. Zero
	// disables the size limit.
	MaxSizeBytes int64

	// RotateAfter rotates the file once it is older than this duration. Zero
	// disables the age limit.
	RotateAfter time.Duration
}

// InitLogger updates the configuration of the default logger.
// It should be called after loading the application configuration.
//
// Initially, the standard Go logger is used. This function allows setting the
// log verbosity level and configuring output destinations.
func InitLogger(opts Options) {
	level := slog.LevelInfo
	if opts.Verbose {
		level = slog.LevelDebug
	}

	writers := []io.Writer{os.Stdout}
	if opts.LogDir != "" {
		writers = append(writers, NewRotatingFileWriter(
			filepath.Join(opts.LogDir, LogFileName),
			opts.MaxSizeBytes,
			opts.RotateAfter,
		))
	}

	handler := slog.NewTextHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level: level,
	})

	custom_logger = slog.New(handler)
}

func getLogger() *slog.Logger {
	if custom_logger == nil {
		// If the logger is not initialized, then create a new one with INFO level (specially
		// usable for testing)
		InitLogger(Options{})
		return custom_logger
	}
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
