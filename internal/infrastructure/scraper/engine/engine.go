package engine

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

	log "github.com/elias-gill/poliplanner2/logger"
)

const DefaultTargetURL = "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/"

var (
	ErrorNoSourceFound      = errors.New("no sources found")
	ErrorCannotParseURI     = errors.New("cannot parse target uri")
	ErrorGoogleDriveNotInit = errors.New("google drive helper not configured")

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

// Filter decides whether a discovered Excel file name is relevant for a given
// domain scraper. The engine is agnostic to what "relevant" means, which keeps
// schedule and laboratory discovery completely independent.
type Filter func(name string) bool

// ScrapeEngine is the shared, domain-agnostic scraping machinery: it walks a
// target page, resolves links and delegates Google Drive/Sheets resolution to a
// DriveHelper. Domain scrapers configure it with their own filter.
type ScrapeEngine struct {
	targetURL   string
	baseURL     *url.URL
	driveHelper DriveHelper
}

// NewScrapeEngine builds the engine. When targetURL is empty DefaultTargetURL is
// used. A nil DriveHelper disables Google Drive/Sheets handling.
func NewScrapeEngine(drive DriveHelper, targetURL string) *ScrapeEngine {
	if targetURL == "" {
		targetURL = DefaultTargetURL
	}

	base, err := url.Parse(targetURL)
	if err != nil {
		panic(fmt.Errorf("%w: %s %+v", ErrorCannotParseURI, targetURL, err))
	}

	if drive == nil {
		log.Warn(ErrorGoogleDriveNotInit.Error())
	}

	return &ScrapeEngine{
		targetURL:   targetURL,
		baseURL:     base,
		driveHelper: drive,
	}
}

// Discover walks the configured target page and returns every relevant source.
func (e *ScrapeEngine) Discover(ctx context.Context, filter Filter) ([]*WebSource, error) {
	log.Info("Finding web sources", "target_url", e.targetURL)

	sources := make([]*WebSource, 0, 16)

	collector := colly.NewCollector(
		colly.AllowedDomains("www.pol.una.py"),
		colly.MaxDepth(1),
		colly.IgnoreRobotsTxt(),
		colly.Async(false),
	)

	collector.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	collector.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			r.Abort()
		default:
		}
	})

	collector.OnHTML("a[href]", func(elt *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		absolute := elt.Request.AbsoluteURL(elt.Attr("href"))
		e.processURL(ctx, absolute, filter, &sources)
	})

	collector.OnError(func(r *colly.Response, err error) {
		log.Warn("Scraper error", "url", r.Request.URL, "error", err)
	})

	if err := collector.Visit(e.targetURL); err != nil {
		return nil, err
	}

	collector.Wait()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(sources) == 0 {
		return nil, ErrorNoSourceFound
	}

	return sources, nil
}

// FindSourcesFromHTML runs the same link extraction over an in-memory HTML
// document. It performs no network request and is meant for tests and offline
// inspection of a saved page.
func (e *ScrapeEngine) FindSourcesFromHTML(ctx context.Context, htmlContent string, filter Filter) ([]*WebSource, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	sources := make([]*WebSource, 0, 4)

	doc.Find("a[href]").Each(func(_ int, selection *goquery.Selection) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		href, _ := selection.Attr("href")
		absolute := e.makeAbsoluteURL(href)
		e.processURL(ctx, absolute, filter, &sources)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(sources) == 0 {
		return nil, ErrorNoSourceFound
	}

	return sources, nil
}

func (e *ScrapeEngine) processURL(ctx context.Context, absoluteURL string, filter Filter, sources *[]*WebSource) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Google Drive folders and spreadsheets are resolved through the helper.
	if e.driveHelper != nil && strings.Contains(absoluteURL, "google.com") {
		e.processGoogleURL(ctx, absoluteURL, filter, sources)
		return
	}

	// Direct Excel links.
	parsed, err := url.Parse(absoluteURL)
	if err != nil {
		return
	}

	name := filepath.Base(parsed.Path)
	if !filter(name) {
		return
	}

	if s := buildDirectSource(absoluteURL, name); s != nil {
		*sources = append(*sources, s)
	}
}

func (e *ScrapeEngine) processGoogleURL(ctx context.Context, absoluteURL string, filter Filter, sources *[]*WebSource) {
	if googleDriveFolderPattern.MatchString(absoluteURL) {
		list, err := e.driveHelper.ListSourcesInURL(ctx, absoluteURL)
		if err != nil {
			return
		}

		for _, item := range list {
			if filter(item.Name) {
				*sources = append(*sources, item)
			}
		}
		return
	}

	if googleSpreadsheetPattern.MatchString(absoluteURL) {
		src, err := e.driveHelper.GetSourceFromSpreadsheetLink(ctx, absoluteURL)
		if err == nil && src != nil && filter(src.Name) {
			*sources = append(*sources, src)
		}
	}
}

// buildDirectSource creates a source for a directly linked Excel file. Files
// whose name does not encode a date are ignored, as the date is what allows the
// sync logic to decide which version is the newest.
func buildDirectSource(uri, name string) *WebSource {
	date, err := extractDateFromFilename(name)
	if err != nil {
		return nil
	}

	return &WebSource{
		URL:        uri,
		Name:       name,
		UploadDate: date,
		Semester:   extractPeriodFromFilename(name),
	}
}

func (e *ScrapeEngine) makeAbsoluteURL(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	rel, err := url.Parse(href)
	if err != nil {
		return href
	}
	return e.baseURL.ResolveReference(rel).String()
}
