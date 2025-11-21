package scrapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func NewGoogleDriveHelper() *GoogleDriveHelper {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.GetLogger().Warn("GOOGLE_API_KEY environment variable not set")
	} else {
		log.GetLogger().Debug("Google Drive Helper initialized with API key", "key_length", len(apiKey))
	}

	return &GoogleDriveHelper{
		apiKey: apiKey,
		folderIDPattern: regexp.MustCompile(
			`folders/([a-zA-Z0-9_-]+)`),
		spreadsheetIDPattern: regexp.MustCompile(
			`spreadsheets/d/([a-zA-Z0-9_-]+)`),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *GoogleDriveHelper) ListSourcesInURL(url string) ([]*ExcelDownloadSource, error) {
	log.GetLogger().Info("Searching for sources in Google Drive URL", "url", url)

	folderID := g.extractFolderID(url)
	if folderID == "" {
		return nil, fmt.Errorf("could not extract folder ID from URL: %s", url)
	}

	log.GetLogger().Debug("Extracted folder ID", "folder_id", folderID)
	files, err := g.listFilesInFolder(folderID)
	if err != nil {
		return nil, fmt.Errorf("error listing files in folder: %v", err)
	}

	log.GetLogger().Debug("Retrieved files from folder", "total_files", len(files))
	var sources []*ExcelDownloadSource
	excelCount := 0

	for _, file := range files {
		if g.isExcelFile(file.Name) {
			downloadURL := fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", file.ID)
			fileDate, err := extractDateFromFilename(file.Name)

			if err != nil {
				log.GetLogger().Debug("Skipping file - cannot extract date", "file", file.Name, "error", err)
				continue
			}

			sources = append(sources, &ExcelDownloadSource{
				URL:        downloadURL,
				FileName:   file.Name,
				UploadDate: fileDate,
			})
			excelCount++
			log.GetLogger().Debug("Added Excel source", "file", file.Name, "date", fileDate.Format("2006-01-02"))
		}
	}

	log.GetLogger().Info("Successfully extracted Google Drive sources",
		"excel_files", excelCount,
		"total_files", len(files),
		"folder_id", folderID)
	return sources, nil
}

func (g *GoogleDriveHelper) GetSourceFromSpreadsheetLink(url string) (*ExcelDownloadSource, error) {
	log.GetLogger().Info("Processing Google Spreadsheet link", "url", url)

	spreadsheetID := g.extractSpreadsheetID(url)
	if spreadsheetID == "" {
		return nil, fmt.Errorf("could not extract spreadsheet ID from URL: %s", url)
	}

	log.GetLogger().Debug("Extracted spreadsheet ID", "spreadsheet_id", spreadsheetID)
	metadata, err := g.fetchSpreadsheetMetadata(spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("error fetching spreadsheet metadata: %v", err)
	}

	log.GetLogger().Debug("Retrieved spreadsheet metadata", "name", metadata.Name)
	if !g.containsExamKeyword(metadata.Name) {
		return nil, fmt.Errorf("spreadsheet name does not contain exam keywords: %s", metadata.Name)
	}

	downloadURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=xlsx", spreadsheetID)
	fileDate, err := extractDateFromFilename(metadata.Name)

	if err != nil {
		return nil, fmt.Errorf("could not extract date from filename: %s", metadata.Name)
	}

	source := &ExcelDownloadSource{
		URL:        downloadURL,
		FileName:   metadata.Name,
		UploadDate: fileDate,
	}

	log.GetLogger().Info("Successfully created spreadsheet source",
		"file", source.FileName,
		"date", source.UploadDate.Format("2006-01-02"),
		"spreadsheet_id", spreadsheetID)
	return source, nil
}

// =====================================
// ======== Private methods ============
// =====================================

func (g *GoogleDriveHelper) extractFolderID(url string) string {
	matches := g.folderIDPattern.FindStringSubmatch(url)
	if len(matches) > 1 {
		folderID := matches[1]
		log.GetLogger().Debug("Extracted folder ID from URL", "url", url, "folder_id", folderID)
		return folderID
	}
	log.GetLogger().Debug("Could not extract folder ID from URL", "url", url)
	return ""
}

func (g *GoogleDriveHelper) extractSpreadsheetID(url string) string {
	matches := g.spreadsheetIDPattern.FindStringSubmatch(url)
	if len(matches) > 1 {
		spreadsheetID := matches[1]
		log.GetLogger().Debug("Extracted spreadsheet ID from URL", "url", url, "spreadsheet_id", spreadsheetID)
		return spreadsheetID
	}
	log.GetLogger().Debug("Could not extract spreadsheet ID from URL", "url", url)
	return ""
}

func (g *GoogleDriveHelper) listFilesInFolder(folderID string) ([]GoogleFile, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files?q='%s'+in+parents&key=%s",
		folderID, g.apiKey)

	log.GetLogger().Debug("Making Google Drive API request", "folder_id", folderID, "url", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.GetLogger().Error("Error creating Google Drive API request", "error", err)
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		log.GetLogger().Error("Error making Google Drive API request", "error", err)
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.GetLogger().Error("Google Drive API request failed",
			"status_code", resp.StatusCode,
			"response", string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var filesResponse GoogleFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&filesResponse); err != nil {
		log.GetLogger().Error("Error decoding Google Drive API response", "error", err)
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	log.GetLogger().Debug("Successfully decoded Google Drive API response", "files_count", len(filesResponse.Files))
	return filesResponse.Files, nil
}

func (g *GoogleDriveHelper) fetchSpreadsheetMetadata(spreadsheetID string) (*GoogleFile, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?fields=name&key=%s",
		spreadsheetID, g.apiKey)

	log.GetLogger().Debug("Fetching spreadsheet metadata", "spreadsheet_id", spreadsheetID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.GetLogger().Error("Error creating spreadsheet metadata request", "error", err)
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		log.GetLogger().Error("Error making spreadsheet metadata request", "error", err)
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.GetLogger().Error("Spreadsheet metadata request failed",
			"status_code", resp.StatusCode,
			"response", string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var file GoogleFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		log.GetLogger().Error("Error decoding spreadsheet metadata response", "error", err)
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	log.GetLogger().Debug("Successfully retrieved spreadsheet metadata", "name", file.Name)
	return &file, nil
}

func (g *GoogleDriveHelper) isExcelFile(filename string) bool {
	isExcel := len(filename) > 5 && filename[len(filename)-5:] == ".xlsx"
	log.GetLogger().Debug("Checking if file is Excel", "file", filename, "is_excel", isExcel)
	return isExcel
}

func (g *GoogleDriveHelper) containsExamKeyword(filename string) bool {
	lowerName := strings.ToLower(filename)
	keywords := []string{"examen", "exame", "exam", "horario", "clases"}

	for _, keyword := range keywords {
		if strings.Contains(lowerName, keyword) {
			log.GetLogger().Debug("Filename contains exam keyword", "file", filename, "keyword", keyword)
			return true
		}
	}
	log.GetLogger().Debug("Filename does not contain exam keywords", "file", filename)
	return false
}
