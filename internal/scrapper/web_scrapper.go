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

	log "github.com/elias-gill/poliplanner2/internal/logger"
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
	baseURL      *url.URL
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

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	},
}

func NewWebScrapper(googleHelper *GoogleDriveHelper) *WebScrapper {
	uri := "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/"
	base, err := url.Parse(uri)
	if err != nil {
		panic(err.Error())
	}
	log.GetLogger().Debug("Creating web scrapper", "target_url", uri)

	if googleHelper == nil {
		log.GetLogger().Warn("No Google drive helper configured", "target_url", uri)
	}

	return &WebScrapper{
		targetURL:    uri,
		baseURL:      base,
		googleHelper: googleHelper,
	}
}

func (ws *WebScrapper) FindLatestDownloadSource() (*ExcelDownloadSource, error) {
	log.GetLogger().Info("Finding latest download source", "target_url", ws.targetURL)

	sources, err := ws.extractSourcesFromURL(ws.targetURL)
	if err != nil {
		return nil, fmt.Errorf("error scraping URL: %v", err)
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	log.GetLogger().Info("Found potential sources", "count", len(sources))
	var latestSource *ExcelDownloadSource
	for _, source := range sources {
		log.GetLogger().Debug("Evaluating source",
			"file", source.FileName,
			"date", source.UploadDate.Format("2006-01-02"),
			"url", source.URL)
		if latestSource == nil || source.UploadDate.After(latestSource.UploadDate) {
			latestSource = source
		}
	}

	if latestSource != nil {
		log.GetLogger().Info("Selected latest source",
			"file", latestSource.FileName,
			"date", latestSource.UploadDate.Format("2006-01-02"))
	}
	return latestSource, nil
}

// For debugging/testing
func (ws *WebScrapper) FindLatestSourceFromHTML(htmlContent string) (*ExcelDownloadSource, error) {
	log.GetLogger().Debug("Finding latest source from HTML content", "content_length", len(htmlContent))

	sources, err := ws.extractSourcesFromHTML(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("error scraping HTML: %v", err)
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	log.GetLogger().Debug("Found sources from HTML", "count", len(sources))
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
	log.GetLogger().Info("Downloading source", "file", s.FileName, "url", s.URL)

	req, _ := http.NewRequest("GET", s.URL, nil)
	req.Header.Set("User-Agent", "poliplanner-bot/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.GetLogger().Error("HTTP request failed", "error", err, "url", s.URL)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.GetLogger().Error("HTTP request returned non-OK status", "status_code", resp.StatusCode, "url", s.URL)
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	cleanName := strings.TrimSuffix(s.FileName, filepath.Ext(s.FileName))
	tempFile, err := os.CreateTemp("", "horario_"+cleanName+"__*.xlsx")
	if err != nil {
		log.GetLogger().Error("Failed to create temporary file", "error", err)
		return "", err
	}
	defer tempFile.Close()

	bytesCopied, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		log.GetLogger().Error("Failed to copy response body to file", "error", err)
		os.Remove(tempFile.Name())
		return "", err
	}

	log.GetLogger().Info("Download completed successfully",
		"file", tempFile.Name(),
		"size_bytes", bytesCopied,
		"original_name", s.FileName)
	return tempFile.Name(), nil
}

// =====================================
// ======== Private methods ============
// =====================================

func (ws *WebScrapper) extractSourcesFromURL(targetURL string) ([]*ExcelDownloadSource, error) {
	log.GetLogger().Debug("Extracting sources from URL", "url", targetURL)

	sources := make([]*ExcelDownloadSource, 0, 20)
	collector := colly.NewCollector(
		colly.AllowedDomains("www.pol.una.py"),
		colly.Async(true),
		colly.MaxDepth(1),
		colly.IgnoreRobotsTxt(),
	)
	collector.SetRequestTimeout(10 * time.Second)

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(href)
		log.GetLogger().Debug("Found link", "href", href, "absolute_url", absoluteURL)
		ws.processURL(absoluteURL, &sources)
	})

	collector.OnError(func(r *colly.Response, err error) {
		log.GetLogger().Warn("Error scraping URL", "url", r.Request.URL, "error", err)
	})

	collector.OnScraped(func(r *colly.Response) {
		log.GetLogger().Debug("Scraping completed", "url", r.Request.URL, "sources_found", len(sources))
	})

	err := collector.Visit(targetURL)
	if err != nil {
		log.GetLogger().Error("Failed to visit target URL", "url", targetURL, "error", err)
		return nil, err
	}
	collector.Wait()

	log.GetLogger().Info("URL scraping completed", "total_sources_found", len(sources))
	return sources, nil
}

func (ws *WebScrapper) extractSourcesFromHTML(htmlContent string) ([]*ExcelDownloadSource, error) {
	log.GetLogger().Debug("Extracting sources from HTML", "html_length", len(htmlContent))

	sources := make([]*ExcelDownloadSource, 0, 20)
	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.IgnoreRobotsTxt(),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		abs := ws.makeAbsoluteURL(href)
		log.GetLogger().Debug("Processing HTML link", "href", href, "absolute_url", abs)
		ws.processURL(abs, &sources)
	})

	if err := c.PostRaw(ws.targetURL, []byte(htmlContent)); err != nil {
		log.GetLogger().Error("Failed to parse HTML content", "error", err)
		return nil, fmt.Errorf("parse html: %v", err)
	}

	log.GetLogger().Debug("HTML parsing completed", "sources_found", len(sources))
	return sources, nil
}

func (ws *WebScrapper) processURL(absoluteURL string, sources *[]*ExcelDownloadSource) {
	log.GetLogger().Debug("Processing URL", "url", absoluteURL)

	if ws.isDirectExcelDownloadURL(absoluteURL) {
		log.GetLogger().Debug("URL matches direct Excel download pattern")
		if source := ws.extractDirectSource(absoluteURL); source != nil {
			*sources = append(*sources, source)
			log.GetLogger().Debug("Added direct download source", "file", source.FileName)
		}
		return
	}

	if !strings.Contains(absoluteURL, "google.com") {
		log.GetLogger().Debug("URL is not a Google service, skipping", "url", absoluteURL)
		return
	}

	if ws.googleHelper == nil {
		log.GetLogger().Debug("Google helper not available, skipping Google URL")
		return
	}

	if ws.googleHelper != nil {
		if ws.isGoogleDriveFolderURL(absoluteURL) {
			log.GetLogger().Debug("URL is Google Drive folder", "url", absoluteURL)
			if driveSources, _ := ws.googleHelper.ListSourcesInURL(absoluteURL); len(driveSources) > 0 {
				*sources = append(*sources, driveSources...)
				log.GetLogger().Debug("Added Google Drive folder sources", "count", len(driveSources))
			}
		} else if ws.isGoogleSpreadsheetURL(absoluteURL) {
			log.GetLogger().Debug("URL is Google Spreadsheet", "url", absoluteURL)
			if source, _ := ws.googleHelper.GetSourceFromSpreadsheetLink(absoluteURL); source != nil {
				*sources = append(*sources, source)
				log.GetLogger().Debug("Added Google Spreadsheet source", "file", source.FileName)
			}
		}
	} else {
		log.GetLogger().Debug("Google drive URL skipped", "url", absoluteURL)
	}
}

func (ws *WebScrapper) isDirectExcelDownloadURL(url string) bool {
	matches := directDownloadPattern.MatchString(url)
	log.GetLogger().Debug("Checking direct download pattern", "url", url, "matches", matches)
	return matches
}

func (ws *WebScrapper) isGoogleDriveFolderURL(url string) bool {
	matches := googleDriveFolderPattern.MatchString(url)
	log.GetLogger().Debug("Checking Google Drive folder pattern", "url", url, "matches", matches)
	return matches
}

func (ws *WebScrapper) isGoogleSpreadsheetURL(url string) bool {
	matches := googleSpreadsheetPattern.MatchString(url)
	log.GetLogger().Debug("Checking Google Spreadsheet pattern", "url", url, "matches", matches)
	return matches
}

func (ws *WebScrapper) extractDirectSource(uri string) *ExcelDownloadSource {
	log.GetLogger().Debug("Extracting direct source", "url", uri)

	parsedURL, err := url.Parse(uri)
	if err != nil {
		log.GetLogger().Warn("Failed to parse URL", "url", uri, "error", err)
		return nil
	}

	fileName := parsedURL.Path
	if idx := strings.LastIndex(fileName, "/"); idx != -1 {
		fileName = fileName[idx+1:]
	}

	date, err := extractDateFromFilename(fileName)
	if err != nil {
		log.GetLogger().Debug("Could not extract date from filename", "file", fileName, "error", err)
		// TODO: start working on log.Logger.n patterns
		return nil
	}

	source := &ExcelDownloadSource{
		URL:        uri,
		FileName:   fileName,
		UploadDate: date,
	}

	log.GetLogger().Debug("Direct source extracted", "file", fileName, "date", date.Format("2006-01-02"))
	return source
}

func (ws *WebScrapper) makeAbsoluteURL(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	relative, err := url.Parse(href)
	if err != nil {
		log.GetLogger().Debug("Failed to parse relative URL", "href", href, "error", err)
		return href
	}
	absolute := ws.baseURL.ResolveReference(relative).String()
	log.GetLogger().Debug("Converted relative to absolute URL", "relative", href, "absolute", absolute)
	return absolute
}
