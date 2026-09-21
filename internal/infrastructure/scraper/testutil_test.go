package scraper

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper/engine"
)

// fakeDriveHelper mirrors the engine fake. Each package keeps its own so the
// tests do not depend on each other.
type fakeDriveHelper struct {
	folderSources []*engine.WebSource
	spreadsheet   *engine.WebSource

	folderCalls int
}

func (f *fakeDriveHelper) ListSourcesInURL(_ context.Context, _ string) ([]*engine.WebSource, error) {
	f.folderCalls++
	return f.folderSources, nil
}

func (f *fakeDriveHelper) GetSourceFromSpreadsheetLink(_ context.Context, _ string) (*engine.WebSource, error) {
	return f.spreadsheet, nil
}

func readTestData(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(repoRoot(t), "test_data", "webscraper", name))
	if err != nil {
		t.Fatalf("read test data %q: %v", name, err)
	}
	return string(data)
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root (go.mod not found)")
		}
		dir = parent
	}
}
