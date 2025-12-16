package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	log "github.com/elias-gill/poliplanner2/internal/logger"
)

// ================================
// ======== Data Structures =======
// ================================

type GoogleDriveHelper struct {
	apiKey               string
	folderIDPattern      *regexp.Regexp
	spreadsheetIDPattern *regexp.Regexp
	httpClient           *http.Client
}

type GoogleFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GoogleFilesResponse struct {
	Files []GoogleFile `json:"files"`
}

// ================================
// ======== Public API ============
// ================================

func NewGoogleDriveHelper(apiKey string) *GoogleDriveHelper {
	if apiKey == "" {
		log.Warn("GOOGLE_API_KEY not set, Google Drive integration disabled")
		return nil
	}

	return &GoogleDriveHelper{
		apiKey: apiKey,
		folderIDPattern: regexp.MustCompile(
			`folders/([a-zA-Z0-9_-]+)`),
		spreadsheetIDPattern: regexp.MustCompile(
			`spreadsheets/d/([a-zA-Z0-9_-]+)`),
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       5,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
	}
}

func (g *GoogleDriveHelper) ListSourcesInURL(
	ctx context.Context,
	url string,
) ([]*ExcelDownloadSource, error) {
	log.Info("Listing Google Drive folder sources", "url", url)

	folderID := g.extractFolderID(url)
	if folderID == "" {
		return nil, fmt.Errorf("could not extract folder ID from url")
	}

	files, err := g.listFilesInFolder(ctx, folderID)
	if err != nil {
		return nil, err
	}

	sources := make([]*ExcelDownloadSource, 0, len(files))
	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !g.isExcelFile(file.Name) {
			continue
		}

		fileDate, err := extractDateFromFilename(file.Name)
		if err != nil {
			log.Warn("Skipping file, date not found", "file", file.Name)
			continue
		}

		source := &ExcelDownloadSource{
			URL:        "https://drive.google.com/uc?export=download&id=" + file.ID,
			FileName:   file.Name,
			UploadDate: fileDate,
		}
		sources = append(sources, source)

		log.Info("Google Drive source found", "url", source.URL, "date", source.UploadDate.String())
	}

	return sources, nil
}

func (g *GoogleDriveHelper) GetSourceFromSpreadsheetLink(
	ctx context.Context,
	url string,
) (*ExcelDownloadSource, error) {
	log.Info("Processing Google Spreadsheet link", "url", url)

	spreadsheetID := g.extractSpreadsheetID(url)
	if spreadsheetID == "" {
		return nil, fmt.Errorf("could not extract spreadsheet ID")
	}

	metadata, err := g.fetchSpreadsheetMetadata(ctx, spreadsheetID)
	if err != nil {
		return nil, err
	}

	if !g.containsExamKeyword(metadata.Name) {
		return nil, fmt.Errorf("spreadsheet name does not match expected keywords")
	}

	date, err := extractDateFromFilename(metadata.Name)
	if err != nil {
		return nil, err
	}

	return &ExcelDownloadSource{
		URL:        "https://docs.google.com/spreadsheets/d/" + spreadsheetID + "/export?format=xlsx",
		FileName:   metadata.Name,
		UploadDate: date,
	}, nil
}

// =====================================
// ======== Private methods ============
// =====================================

func (g *GoogleDriveHelper) listFilesInFolder(
	ctx context.Context,
	folderID string,
) ([]GoogleFile, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	reqURL := fmt.Sprintf(
		"https://www.googleapis.com/drive/v3/files?q='%s'+in+parents&key=%s",
		folderID,
		g.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google drive api error %d: %s", resp.StatusCode, body)
	}

	var result GoogleFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Files, nil
}

func (g *GoogleDriveHelper) fetchSpreadsheetMetadata(
	ctx context.Context,
	spreadsheetID string,
) (*GoogleFile, error) {
	reqURL := fmt.Sprintf(
		"https://www.googleapis.com/drive/v3/files/%s?fields=name&key=%s",
		spreadsheetID,
		g.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("metadata api error %d: %s", resp.StatusCode, body)
	}

	var file GoogleFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, err
	}

	return &file, nil
}

func (g *GoogleDriveHelper) extractFolderID(url string) string {
	m := g.folderIDPattern.FindStringSubmatch(url)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func (g *GoogleDriveHelper) extractSpreadsheetID(url string) string {
	m := g.spreadsheetIDPattern.FindStringSubmatch(url)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func (g *GoogleDriveHelper) isExcelFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".xlsx")
}

func (g *GoogleDriveHelper) containsExamKeyword(name string) bool {
	n := strings.ToLower(name)
	for _, k := range []string{"examen", "exame", "exam", "horario", "clases"} {
		if strings.Contains(n, k) {
			return true
		}
	}
	return false
}
