package scrapper

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	directDownloadURL = "https://www.pol.una.py/wp-content/uploads/Horario-de-clases-y-examenes-Segundo-Academico-2024-version-web-19122024.xlsx"
	driveDownloadURL  = "https://drive.google.com/uc?export=download&id=1BVbZHZ6w01MLzGYBRBx2mbDkJ3-7QtLZ"
)

var (
	testdataPath      = filepath.Join("..", "..", "testdata", "webscrapper")
	htmlNoDrivePath   = filepath.Join(testdataPath, "page_without_drive_folders.html")
	htmlWithDrivePath = filepath.Join(testdataPath, "page_with_drive_folders.html")
	expectedDirectURL = directDownloadURL
	expectedDriveURL  = driveDownloadURL
)

func TestFindLatestExcelUrlFromLocalHtml(t *testing.T) {
	html, err := os.ReadFile(htmlNoDrivePath)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}

	s := NewWebScrapper(nil)
	start := time.Now()
	src, err := s.FindLatestSourceFromHTML(string(html))
	end := time.Now()
	t.Logf("Scrapping concluded in: %dms", end.Sub(start).Milliseconds())

	if err != nil {
		t.Fatalf("find source: %v", err)
	}
	if src.URL != expectedDirectURL {
		t.Errorf("url mismatch\nwant: %s\ngot:  %s", expectedDirectURL, src.URL)
	}
}

func TestFindLatestExcelUrlWithDriveFolders(t *testing.T) {
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	html, err := os.ReadFile(htmlWithDrivePath)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}

	helper := NewGoogleDriveHelper()
	s := NewWebScrapper(helper)
	src, err := s.FindLatestSourceFromHTML(string(html))
	if err != nil {
		t.Fatalf("find source: %v", err)
	}
	if src.URL != expectedDriveURL {
		t.Errorf("url mismatch\nwant: %s\ngot:  %s", expectedDriveURL, src.URL)
	}
}
