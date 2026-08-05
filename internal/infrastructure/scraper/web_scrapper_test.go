package scraper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
)

const (
	directDownloadURL = "https://www.pol.una.py/wp-content/uploads/Horario-de-clases-y-examenes-Segundo-Academico-2024-version-web-19122024.xlsx"
)

var (
	testPath          = filepath.Join(config.Get().Paths.BaseDir, "test_data", "webscraper")
	htmlNoDrivePath   = filepath.Join(testPath, "page_without_drive_folders.html")
	htmlWithDrivePath = filepath.Join(testPath, "page_with_drive_folders.html")
	expectedDirectURL = directDownloadURL
	expectedDriveURL  = driveDownloadURL
)

func TestFindLatestExcelUrlFromLocalHtml(t *testing.T) {
	html, err := os.ReadFile(htmlNoDrivePath)
	if err != nil {
		t.Fatalf("read html: %+v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Scrap and measure execution times
	s := NewWebScraper(nil)
	start := time.Now()
	src, err := s.FindSourcesFromHTML(ctx, string(html))
	end := time.Now()
	t.Logf("Scrapping concluded in: %dms", end.Sub(start).Milliseconds())

	if err != nil {
		t.Fatalf("find source: %+v", err)
	}
	if src[0].URL != expectedDirectURL {
		t.Errorf("url mismatch\nwant: %s\ngot:  %s", expectedDirectURL, src[0].URL)
	}
}

func TestFindLatestExcelUrlWithDriveFolders(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
		return
	}

	html, err := os.ReadFile(htmlWithDrivePath)
	if err != nil {
		t.Fatalf("read html: %+v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	helper := NewGoogleDriveHelper(apiKey)
	s := NewWebScraper(helper)
	src, err := s.FindSourcesFromHTML(ctx, string(html))
	if err != nil {
		t.Fatalf("find source: %+v", err)
	}
	if src[0].URL != expectedDriveURL {
		t.Errorf("url mismatch\nwant: %s\ngot:  %s", expectedDriveURL, src[0].URL)
	}
}
