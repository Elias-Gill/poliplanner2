package scraper

import (
	"context"
	"testing"
	"time"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper/engine"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

func TestLabScraper_FindSourcesFromHTML_viaDriveFolder(t *testing.T) {
	html := readTestData(t, "page_with_drive_folders.html")

	const labFileName = "Horario-de-Practicas-del-Laboratorio-Electronica-2026-01012026.xlsx"

	drive := &fakeDriveHelper{
		folderSources: []*engine.WebSource{
			{
				URL:        "https://drive.google.com/uc?export=download&id=lab",
				Name:       labFileName,
				UploadDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Semester:   academic.FirstSemester,
			},
		},
	}

	l := NewLabScraper(drive, "")
	sources, err := l.FindSourcesFromHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("FindSourcesFromHTML: %v", err)
	}

	if drive.folderCalls != 1 {
		t.Errorf("expected the drive helper to be called once, got %d", drive.folderCalls)
	}

	if len(sources) != 1 {
		t.Fatalf("expected 1 laboratory source, got %d", len(sources))
	}

	if got := sources[0].Metadata().Name; got != labFileName {
		t.Errorf("unexpected source name: %q", got)
	}
}
