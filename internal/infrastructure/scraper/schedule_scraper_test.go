package scraper

import (
	"context"
	"strings"
	"testing"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
)

const latestScheduleURL = "https://www.pol.una.py/wp-content/uploads/Horario-de-clases-y-examenes-Segundo-Academico-2024-version-web-19122024.xlsx"

func TestScheduleScraper_FindSourcesFromHTML(t *testing.T) {
	html := readTestData(t, "page_without_drive_folders.html")

	s := NewScheduleScraper(nil, "")
	sources, err := s.FindSourcesFromHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("FindSourcesFromHTML: %v", err)
	}

	if !scheduleHasURL(sources, latestScheduleURL) {
		t.Errorf("expected %q to be discovered", latestScheduleURL)
	}

	// The laboratory files in the same page are also named "Horario ...", so the
	// schedule scraper must explicitly exclude them.
	for _, src := range sources {
		name := strings.ToLower(src.Metadata().Name)
		if strings.Contains(name, "laboratorio") {
			t.Errorf("schedule scraper must not return laboratory files, got %q", name)
		}
	}
}

func scheduleHasURL(sources []source.ScheduleSource, uri string) bool {
	for _, src := range sources {
		if src.Metadata().URI == uri {
			return true
		}
	}
	return false
}
