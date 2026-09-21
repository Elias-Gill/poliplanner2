package logger

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInitLogger_WritesToLogDir(t *testing.T) {
	dir := t.TempDir()

	InitLogger(Options{LogDir: dir})
	Info("hello from test", "key", "value")

	data := readFile(t, filepath.Join(dir, LogFileName))
	if !strings.Contains(data, "hello from test") {
		t.Fatalf("log file does not contain the message: %q", data)
	}
	if !strings.Contains(data, "key=value") {
		t.Fatalf("log file does not contain the structured field: %q", data)
	}
}
