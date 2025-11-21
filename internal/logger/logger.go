package logger

import (
	"log/slog"
	"os"
)

var custom_logger *slog.Logger

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

func GetLogger() *slog.Logger {
	if custom_logger == nil {
		// If the logger is not initialized, then create a new one with INFO level (specially
		// usable for testing)
		InitLogger(false)
		return custom_logger
	}
	return custom_logger
}
