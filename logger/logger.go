package logger

import (
	"log/slog"
	"os"
)

var custom_logger *slog.Logger

// InitLogger updates the configuration of the default logger.
// It should be called after loading the application configuration.
//
// Initially, the standard Go logger is used.
// This function allows setting the log verbosity level and configuring output destinations.
func InitLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	custom_logger = slog.New(handler)
}

func getLogger() *slog.Logger {
	if custom_logger == nil {
		// If the logger is not initialized, then create a new one with INFO level (specially
		// usable for testing)
		InitLogger(false)
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
