package scrapper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// ================================
// ======== Data Structures =======
// ================================

type ExcelDownloadSource struct {
	URL        string
	FileName   string
	UploadDate time.Time
}

type WebScrapper struct {
	targetURL    string
	googleHelper *GoogleDriveHelper
}

// ================================
// ======== Public API ============
// ================================

var (
	directDownloadPattern = regexp.MustCompile(
		`.*(?i)(horario|clases|examen(?:es)?|exame|exam)[\w\-_.]*\.xlsx$`)
	googleDriveFolderPattern = regexp.MustCompile(
		`^https://drive\.google\.com/(?:drive/(?:u/\d+/)?folders|folders)/[\w-]+`)
	googleSpreadsheetPattern = regexp.MustCompile(
		`^https://docs\.google\.com/spreadsheets/d/[\w-]+`)
)

func NewWebScrapper(googleHelper *GoogleDriveHelper) *WebScrapper {
	return &WebScrapper{
		targetURL:    "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/",
		googleHelper: googleHelper,
	}
}

func (ws *WebScrapper) FindLatestDownloadSource() (*ExcelDownloadSource, error) {
	sources, err := ws.extractSourcesFromURL(ws.targetURL)
	if err != nil {
		return nil, fmt.Errorf("error scraping URL: %v", err)
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	var latestSource *ExcelDownloadSource
	for _, source := range sources {
		fmt.Printf("Found source: %s - %s\n", source.FileName, source.UploadDate.Format("2006-01-02"))
		if latestSource == nil || source.UploadDate.After(latestSource.UploadDate) {
			latestSource = source
		}
	}
	return latestSource, nil
}

// For debugging/testing
func (ws *WebScrapper) FindLatestSourceFromHTML(htmlContent string) (*ExcelDownloadSource, error) {
	sources, err := ws.extractSourcesFromHTML(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("error scraping HTML: %v", err)
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	var latestSource *ExcelDownloadSource
	for _, source := range sources {
		if latestSource == nil || source.UploadDate.After(latestSource.UploadDate) {
			latestSource = source
		}
	}
	return latestSource, nil
}

// DownloadThisSource downloads the Excel file to a temporary file
func (s *ExcelDownloadSource) DownloadThisSource() (string, error) {
	resp, err := http.Get(s.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	cleanName := strings.TrimSuffix(s.FileName, filepath.Ext(s.FileName))
	tempFile, err := os.CreateTemp("", "horario_"+cleanName+"__*.xlsx")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}
	return tempFile.Name(), nil
}

// =====================================
// ======== Private methods ============
// =====================================

func (ws *WebScrapper) extractSourcesFromURL(targetURL string) ([]*ExcelDownloadSource, error) {
	var sources []*ExcelDownloadSource
	collector := colly.NewCollector(
		colly.AllowedDomains("www.pol.una.py"),
		colly.Async(true),
	)

	collector.SetRequestTimeout(10 * time.Second)

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(href)
		ws.processURL(absoluteURL, &sources)
	})

	collector.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error scraping %s: %v\n", r.Request.URL, err)
	})

	err := collector.Visit(targetURL)
	if err != nil {
		return nil, err
	}
	collector.Wait()
	return sources, nil
}

// extractSourcesFromHTML (reemplaza la versión anterior)
func (ws *WebScrapper) extractSourcesFromHTML(htmlContent string) ([]*ExcelDownloadSource, error) {
	var sources []*ExcelDownloadSource
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		abs := ws.makeAbsoluteURL(href)
		ws.processURL(abs, &sources)
	})

	if err := c.PostRaw(ws.targetURL, []byte(htmlContent)); err != nil {
		return nil, fmt.Errorf("parse html: %v", err)
	}
	return sources, nil
}

func (ws *WebScrapper) processURL(absoluteURL string, sources *[]*ExcelDownloadSource) {
	switch {
	case ws.isDirectExcelDownloadURL(absoluteURL):
		if source := ws.extractDirectSource(absoluteURL); source != nil {
			*sources = append(*sources, source)
		}
	case ws.isGoogleDriveFolderURL(absoluteURL):
		if ws.googleHelper != nil {
			if driveSources, err := ws.googleHelper.ListSourcesInURL(absoluteURL); err == nil {
				*sources = append(*sources, driveSources...)
			}
		}
	case ws.isGoogleSpreadsheetURL(absoluteURL):
		if ws.googleHelper != nil {
			if source, err := ws.googleHelper.GetSourceFromSpreadsheetLink(absoluteURL); err == nil {
				*sources = append(*sources, source)
			}
		}
	}
}

func (ws *WebScrapper) isDirectExcelDownloadURL(url string) bool {
	return directDownloadPattern.MatchString(url)
}

func (ws *WebScrapper) isGoogleDriveFolderURL(url string) bool {
	return googleDriveFolderPattern.MatchString(url)
}

func (ws *WebScrapper) isGoogleSpreadsheetURL(url string) bool {
	return googleSpreadsheetPattern.MatchString(url)
}

func (ws *WebScrapper) extractDirectSource(uri string) *ExcelDownloadSource {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil
	}

	fileName := parsedURL.Path
	if idx := strings.LastIndex(fileName, "/"); idx != -1 {
		fileName = fileName[idx+1:]
	}

	date := extractDateFromFilename(fileName)
	if date.IsZero() {
		return nil
	}

	return &ExcelDownloadSource{
		URL:        uri,
		FileName:   fileName,
		UploadDate: date,
	}
}

func (ws *WebScrapper) makeAbsoluteURL(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	base, _ := url.Parse(ws.targetURL)
	relative, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(relative).String()
}
