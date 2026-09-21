package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

const (
	noDriveHTMLFile = "page_without_drive_folders.html"
	driveHTMLFile   = "page_with_drive_folders.html"

	latestScheduleURL = "https://www.pol.una.py/wp-content/uploads/Horario-de-clases-y-examenes-Segundo-Academico-2024-version-web-19122024.xlsx"
)

// fakeDriveHelper lets us exercise the Google Drive branch without network or
// API key, which is the whole point of the DriveHelper interface.
type fakeDriveHelper struct {
	folderSources []*WebSource
	spreadsheet   *WebSource

	folderCalls      int
	spreadsheetCalls int
}

func (f *fakeDriveHelper) ListSourcesInURL(_ context.Context, _ string) ([]*WebSource, error) {
	f.folderCalls++
	return f.folderSources, nil
}

func (f *fakeDriveHelper) GetSourceFromSpreadsheetLink(_ context.Context, _ string) (*WebSource, error) {
	f.spreadsheetCalls++
	return f.spreadsheet, nil
}

func TestScrapeEngine_FindSourcesFromHTML_DirectLinks(t *testing.T) {
	html := readTestData(t, noDriveHTMLFile)

	// A filter that only accepts schedule-like files.
	filter := func(name string) bool {
		return strings.Contains(strings.ToLower(name), "horario") && !strings.Contains(strings.ToLower(name), "laboratorio")
	}

	e := NewScrapeEngine(nil, "https://www.pol.una.py/")
	sources, err := e.FindSourcesFromHTML(context.Background(), html, filter)
	if err != nil {
		t.Fatalf("FindSourcesFromHTML: %v", err)
	}

	if !containsURL(sources, latestScheduleURL) {
		t.Errorf("expected source %q to be discovered", latestScheduleURL)
	}

	for _, s := range sources {
		if !strings.HasSuffix(strings.ToLower(s.Name), ".xlsx") {
			t.Errorf("unexpected non-excel source: %q", s.Name)
		}
	}
}

func TestScrapeEngine_FindSourcesFromHTML_DriveFolder(t *testing.T) {
	html := readTestData(t, driveHTMLFile)

	const driveFileName = "Horario-de-clases-y-examenes-Drive-2026-01012026.xlsx"

	drive := &fakeDriveHelper{
		folderSources: []*WebSource{
			{
				URL:        "https://drive.google.com/uc?export=download&id=abc",
				Name:       driveFileName,
				UploadDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Semester:   academic.FirstSemester,
			},
		},
	}

	// Restrict the filter to the fake file so the direct links in the page do not
	// pollute the result.
	filter := func(name string) bool { return name == driveFileName }

	e := NewScrapeEngine(drive, "https://www.pol.una.py/")
	sources, err := e.FindSourcesFromHTML(context.Background(), html, filter)
	if err != nil {
		t.Fatalf("FindSourcesFromHTML: %v", err)
	}

	if drive.folderCalls != 1 {
		t.Errorf("expected the drive helper to be called once, got %d", drive.folderCalls)
	}

	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}

	if sources[0].Name != driveFileName {
		t.Errorf("unexpected source name: %q", sources[0].Name)
	}
}

func TestScrapeEngine_FindSourcesFromHTML_NoMatches(t *testing.T) {
	html := readTestData(t, noDriveHTMLFile)

	e := NewScrapeEngine(nil, "https://www.pol.una.py/")
	_, err := e.FindSourcesFromHTML(context.Background(), html, func(string) bool { return false })
	if err == nil {
		t.Fatal("expected an error when nothing matches")
	}
	if err != ErrorNoSourceFound {
		t.Fatalf("expected ErrorNoSourceFound, got %v", err)
	}
}

func TestNewScrapeEngine_DefaultsTargetURL(t *testing.T) {
	e := NewScrapeEngine(nil, "")
	if e.targetURL != DefaultTargetURL {
		t.Errorf("expected default target url %q, got %q", DefaultTargetURL, e.targetURL)
	}
}

func containsURL(sources []*WebSource, url string) bool {
	for _, s := range sources {
		if s.URL == url {
			return true
		}
	}
	return false
}

// readTestData locates the repository test_data/webscraper directory regardless
// of the working directory used by the test binary.
func readTestData(t *testing.T, name string) string {
	t.Helper()

	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "test_data", "webscraper", name))
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
