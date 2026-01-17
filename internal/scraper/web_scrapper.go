package scraper

import (
	"context"
	"crypto/tls"
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

// ==================================
// =        Data Structures         =
// ==================================

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

const default_target = "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/"

// ================================
// =        Public API            =
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
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	},
}

func NewWebScraper(googleHelper *GoogleDriveHelper) *WebScrapper {
	base, err := url.Parse(default_target)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse uri: %s\n%+v", default_target, err))
	}

	if googleHelper == nil {
		log.Warn("No Google Drive helper configured")
	}

	return &WebScrapper{
		targetURL:    default_target,
		baseURL:      base,
		googleHelper: googleHelper,
	}
}

func (ws *WebScrapper) FindLatestDownloadSource(
	ctx context.Context,
) (*ExcelDownloadSource, error) {
	log.Info("Finding latest download source", "target_url", ws.targetURL)

	sources, err := ws.extractSourcesFromURL(ctx, ws.targetURL)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	var latest *ExcelDownloadSource
	for _, s := range sources {
		if latest == nil || s.UploadDate.After(latest.UploadDate) {
			latest = s
		}
	}

	return latest, nil
}

// For debugging / testing
func (ws *WebScrapper) FindLatestSourceFromHTML(
	ctx context.Context,
	htmlContent string,
) (*ExcelDownloadSource, error) {
	log.Debug("Finding latest source from HTML", "content_length", len(htmlContent))

	sources, err := ws.extractSourcesFromHTML(ctx, htmlContent)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	var latest *ExcelDownloadSource
	for _, s := range sources {
		if latest == nil || s.UploadDate.After(latest.UploadDate) {
			latest = s
		}
	}

	return latest, nil
}

// DownloadThisSource downloads the Excel file to a temporary file.
// The request lifetime is fully controlled by ctx.
func (s *ExcelDownloadSource) DownloadThisSource(
	ctx context.Context,
) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "poliplanner-bot/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	cleanName := strings.TrimSuffix(s.FileName, filepath.Ext(s.FileName))
	tmp, err := os.CreateTemp("", "horario_"+cleanName+"__*.xlsx")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

// =====================================
// =        Private methods            =
// =====================================

func (ws *WebScrapper) extractSourcesFromURL(
	ctx context.Context,
	targetURL string,
) ([]*ExcelDownloadSource, error) {
	sources := make([]*ExcelDownloadSource, 0, 16)

	collector := colly.NewCollector(
		colly.AllowedDomains("www.pol.una.py"),
		colly.MaxDepth(1),
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
	)
	collector.WithTransport(&http.Transport{
		// Thanks FPUNA :[. Disable TLS verification.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	collector.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			r.Abort()
		default:
		}
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		href := e.Attr("href")
		absolute := e.Request.AbsoluteURL(href)
		ws.processURL(ctx, absolute, &sources)
	})

	collector.OnError(func(r *colly.Response, err error) {
		log.Warn("Scraper error", "url", r.Request.URL, "error", err)
	})

	if err := collector.Visit(targetURL); err != nil {
		return nil, err
	}

	collector.Wait()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return sources, nil
}

func (ws *WebScrapper) extractSourcesFromHTML(
	ctx context.Context,
	htmlContent string,
) ([]*ExcelDownloadSource, error) {
	sources := make([]*ExcelDownloadSource, 0, 16)

	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.IgnoreRobotsTxt(),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		href := e.Attr("href")
		absolute := ws.makeAbsoluteURL(href)
		ws.processURL(ctx, absolute, &sources)
	})

	if err := c.PostRaw(ws.targetURL, []byte(htmlContent)); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return sources, nil
}

func (ws *WebScrapper) processURL(
	ctx context.Context,
	absoluteURL string,
	sources *[]*ExcelDownloadSource,
) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if ws.isDirectExcelDownloadURL(absoluteURL) {
		if s := ws.extractDirectSource(absoluteURL); s != nil {
			*sources = append(*sources, s)
		}
		return
	}

	if !strings.Contains(absoluteURL, "google.com") {
		return
	}

	if ws.googleHelper == nil {
		return
	}

	if ws.isGoogleDriveFolderURL(absoluteURL) {
		list, err := ws.googleHelper.ListSourcesInURL(ctx, absoluteURL)
		if err == nil && len(list) > 0 {
			*sources = append(*sources, list...)
		}
		return
	}

	if ws.isGoogleSpreadsheetURL(absoluteURL) {
		src, err := ws.googleHelper.GetSourceFromSpreadsheetLink(ctx, absoluteURL)
		if err == nil && src != nil {
			*sources = append(*sources, src)
		}
	}
}

func (ws *WebScrapper) isDirectExcelDownloadURL(u string) bool {
	return directDownloadPattern.MatchString(u)
}

func (ws *WebScrapper) isGoogleDriveFolderURL(u string) bool {
	return googleDriveFolderPattern.MatchString(u)
}

func (ws *WebScrapper) isGoogleSpreadsheetURL(u string) bool {
	return googleSpreadsheetPattern.MatchString(u)
}

func (ws *WebScrapper) extractDirectSource(uri string) *ExcelDownloadSource {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil
	}

	name := filepath.Base(parsed.Path)
	date, err := extractDateFromFilename(name)
	if err != nil {
		return nil
	}

	return &ExcelDownloadSource{
		URL:        uri,
		FileName:   name,
		UploadDate: date,
	}
}

func (ws *WebScrapper) makeAbsoluteURL(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	rel, err := url.Parse(href)
	if err != nil {
		return href
	}
	return ws.baseURL.ResolveReference(rel).String()
}
