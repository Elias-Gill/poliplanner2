package scraper

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

	"github.com/gocolly/colly/v2"

	log "github.com/elias-gill/poliplanner2/logger"
)

const defaultTargetURL = "https://www.pol.una.py/academico/horarios-de-clases-y-examenes/"

var (
	ErrorNoSourceFound      = errors.New("no sources found")
	ErrorCannotParseURI     = errors.New("cannot parse target uri")
	ErrorGoogleDriveNotInit = errors.New("google drive helper not configured")
)

var (
	// Permite espacios y cualquier caracter entre la palabra clave y la extension .xlsx
	schedulePattern = regexp.MustCompile(
		`(?i).*(horario|clases|examen(?:es)?|exame|exam).*\.xlsx$`)
	laboratoryPattern = regexp.MustCompile(
		`(?i).*(laboratorio(?:s)?|lab|asignacior|asignacion).*\.xlsx$`)

	googleDriveFolderPattern = regexp.MustCompile(
		`^https://drive\.google\.com/(?:drive/(?:u/\d+/)?folders|folders)/[\w-]+`)
	googleSpreadsheetPattern = regexp.MustCompile(
		`^https://docs\.google\.com/spreadsheets/d/[\w-]+`)
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression: false,
	},
}

type WebScrapper struct {
	targetURL    string
	baseURL      *url.URL
	googleHelper *GoogleDriveHelper
}

func NewWebScraper(googleHelper *GoogleDriveHelper) *WebScrapper {
	base, err := url.Parse(defaultTargetURL)
	if err != nil {
		panic(fmt.Errorf("%w: %s %+v", ErrorCannotParseURI, defaultTargetURL, err))
	}

	if googleHelper == nil {
		log.Warn(ErrorGoogleDriveNotInit.Error())
	}

	return &WebScrapper{
		targetURL:    defaultTargetURL,
		baseURL:      base,
		googleHelper: googleHelper,
	}
}

func (ws *WebScrapper) DiscoverSchedules(ctx context.Context) ([]*WebScheduleSource, error) {
	log.Info("Finding web schedule sources", "target_url", ws.targetURL)
	rawSources, err := ws.discover(ctx, ws.targetURL, schedulePattern)
	if err != nil {
		return nil, err
	}

	sources := make([]*WebScheduleSource, len(rawSources))
	for i, s := range rawSources {
		sources[i] = (*WebScheduleSource)(s)
	}

	return sources, nil
}

func (ws *WebScrapper) DiscoverLabs(ctx context.Context) ([]*WebLaboratorySource, error) {
	log.Info("Finding web laboratory sources", "target_url", ws.targetURL)
	rawSources, err := ws.discover(ctx, ws.targetURL, laboratoryPattern)
	if err != nil {
		return nil, err
	}

	sources := make([]*WebLaboratorySource, len(rawSources))
	for i, s := range rawSources {
		sources[i] = (*WebLaboratorySource)(s)
	}

	return sources, nil
}

func (ws *WebScrapper) discover(ctx context.Context, targetURL string, pattern *regexp.Regexp) ([]*webSource, error) {
	sources := make([]*webSource, 0, 16)

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

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		href := e.Attr("href")
		absolute := e.Request.AbsoluteURL(href)
		ws.processURL(ctx, absolute, pattern, &sources)
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

	if len(sources) == 0 {
		return nil, ErrorNoSourceFound
	}

	return sources, nil
}

func (ws *WebScrapper) extractFromHTML(ctx context.Context, htmlContent string, pattern *regexp.Regexp) ([]*webSource, error) {
	sources := make([]*webSource, 0, 4)

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
		ws.processURL(ctx, absolute, pattern, &sources)
	})

	if err := c.PostRaw(ws.targetURL, []byte(htmlContent)); err != nil {
		return nil, err
	}

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

func (ws *WebScrapper) processURL(ctx context.Context, absoluteURL string, pattern *regexp.Regexp, sources *[]*webSource) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	// 1. Direct match with target Excel file pattern
	if pattern.MatchString(absoluteURL) {
		if s := ws.extractDirectSource(absoluteURL); s != nil {
			*sources = append(*sources, s)
		}
		return
	}

	if !strings.Contains(absoluteURL, "google.com") || ws.googleHelper == nil {
		return
	}

	// 2. Match with Google Drive Folder
	if googleDriveFolderPattern.MatchString(absoluteURL) {
		list, err := ws.googleHelper.ListSourcesInURL(ctx, absoluteURL)
		if err == nil {
			for _, item := range list {
				if pattern.MatchString(item.Name) {
					*sources = append(*sources, item)
				}
			}
		}
		return
	}

	// 3. Match with Google Spreadsheet link
	if googleSpreadsheetPattern.MatchString(absoluteURL) {
		src, err := ws.googleHelper.GetSourceFromSpreadsheetLink(ctx, absoluteURL)
		if err == nil && src != nil && pattern.MatchString(src.Name) {
			*sources = append(*sources, src)
		}
	}
}

func (ws *WebScrapper) extractDirectSource(uri string) *webSource {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil
	}

	name := filepath.Base(parsed.Path)
	date, err := extractDateFromFilename(name)
	if err != nil {
		return nil
	}

	return &webSource{
		URL:        uri,
		Name:       name,
		UploadDate: date,
		Semester:   extractPeriodFromFilename(name),
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
