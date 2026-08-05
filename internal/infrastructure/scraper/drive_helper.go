package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	log "github.com/elias-gill/poliplanner2/logger"
)

const (
	driveFilesAPI = "https://www.googleapis.com/drive/v3/files"

	excelExtension = ".xlsx"

	spreadsheetExportURL = "https://docs.google.com/spreadsheets/d/%s/export?format=xlsx"
	driveDownloadURL     = "https://drive.google.com/uc?export=download&id=%s"
)

type GoogleDriveHelper struct {
	apiKey               string
	folderIDPattern      *regexp.Regexp
	spreadsheetIDPattern *regexp.Regexp
	httpClient           *http.Client
}

type GoogleFile struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ModifiedDate time.Time `json:"modifiedTime"`
}

type GoogleFilesResponse struct {
	Files []GoogleFile `json:"files"`
}

func NewGoogleDriveHelper(apiKey string) *GoogleDriveHelper {
	if apiKey == "" {
		log.Warn("Google Drive integration disabled: missing API key")
		return nil
	}

	return &GoogleDriveHelper{
		apiKey: apiKey,

		folderIDPattern: regexp.MustCompile(
			`folders/([a-zA-Z0-9_-]+)`,
		),

		spreadsheetIDPattern: regexp.MustCompile(
			`spreadsheets/d/([a-zA-Z0-9_-]+)`,
		),

		httpClient: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       5,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
	}
}

func (g *GoogleDriveHelper) ListSourcesInURL(ctx context.Context, url string) ([]*webSource, error) {
	folderID := g.extractFolderID(url)

	if folderID == "" {
		return nil, fmt.Errorf("invalid Google Drive folder URL")
	}

	log.Info("Scanning Google Drive folder", "folderID", folderID)

	files, err := g.listFilesInFolder(ctx, folderID)
	if err != nil {
		return nil, err
	}

	sources := make([]*webSource, 0, len(files))

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if !g.isExcelFile(file.Name) {
			continue
		}

		source := &webSource{
			URL:        fmt.Sprintf(driveDownloadURL, file.ID),
			Name:       file.Name,
			UploadDate: file.ModifiedDate,
			Semester:   extractPeriodFromFilename(file.Name),
		}

		sources = append(sources, source)
	}

	log.Info(
		"Google Drive scan completed",
		"filesFound", len(sources),
	)

	return sources, nil
}

func (g *GoogleDriveHelper) GetSourceFromSpreadsheetLink(ctx context.Context, link string) (*webSource, error) {
	spreadsheetID := g.extractSpreadsheetID(link)

	if spreadsheetID == "" {
		return nil, fmt.Errorf("invalid Google Spreadsheet URL")
	}

	log.Info(
		"Fetching spreadsheet metadata",
		"id", spreadsheetID,
	)

	metadata, err := g.fetchSpreadsheetMetadata(ctx, spreadsheetID)
	if err != nil {
		return nil, err
	}

	return &webSource{
		URL:        fmt.Sprintf(spreadsheetExportURL, spreadsheetID),
		Name:       metadata.Name,
		UploadDate: metadata.ModifiedDate,
		Semester:   extractPeriodFromFilename(metadata.Name),
	}, nil
}

func (g *GoogleDriveHelper) listFilesInFolder(ctx context.Context, folderID string) ([]GoogleFile, error) {
	query := fmt.Sprintf(
		"'%s' in parents",
		folderID,
	)

	fields := "files(id,name,modifiedTime)"

	reqURL := fmt.Sprintf(
		"%s?q=%s&fields=%s&key=%s",
		driveFilesAPI,
		url.QueryEscape(query),
		fields,
		g.apiKey,
	)

	var result GoogleFilesResponse

	if err := g.doRequest(ctx, reqURL, &result); err != nil {
		return nil, fmt.Errorf("listing Google Drive files: %w", err)
	}

	return result.Files, nil
}

func (g *GoogleDriveHelper) fetchSpreadsheetMetadata(ctx context.Context, spreadsheetID string) (*GoogleFile, error) {
	reqURL := fmt.Sprintf(
		"%s/%s?fields=name,modifiedTime&key=%s",
		driveFilesAPI,
		spreadsheetID,
		g.apiKey,
	)

	var file GoogleFile

	if err := g.doRequest(ctx, reqURL, &file); err != nil {
		return nil, fmt.Errorf("fetching spreadsheet metadata: %w", err)
	}

	return &file, nil
}

func (g *GoogleDriveHelper) doRequest(ctx context.Context, requestURL string, target any) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestURL,
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return fmt.Errorf(
			"Google Drive API returned %d: %s",
			resp.StatusCode,
			body,
		)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (g *GoogleDriveHelper) extractFolderID(url string) string {
	match := g.folderIDPattern.FindStringSubmatch(url)

	if len(match) > 1 {
		return match[1]
	}

	return ""
}

func (g *GoogleDriveHelper) extractSpreadsheetID(url string) string {
	match := g.spreadsheetIDPattern.FindStringSubmatch(url)

	if len(match) > 1 {
		return match[1]
	}

	return ""
}

func (g *GoogleDriveHelper) isExcelFile(name string) bool {
	return strings.HasSuffix(
		strings.ToLower(name),
		excelExtension,
	)
}
