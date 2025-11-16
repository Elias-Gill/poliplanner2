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
)

// ================================
// ======== Data Structures =======
// ================================

type GoogleDriveHelper struct {
	apiKey                   string
	folderIDPattern          *regexp.Regexp
	spreadsheetIDPattern     *regexp.Regexp
	httpClient               *http.Client
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
		fmt.Println("Warning: GOOGLE_API_KEY environment variable not set")
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
	fmt.Printf("Searching in Google Drive: %s\n", url)

	folderID := g.extractFolderID(url)
	if folderID == "" {
		return nil, fmt.Errorf("could not extract folder ID from URL: %s", url)
	}

	files, err := g.listFilesInFolder(folderID)
	if err != nil {
		return nil, fmt.Errorf("error listing files in folder: %v", err)
	}

	var sources []*ExcelDownloadSource
	for _, file := range files {
		if g.isExcelFile(file.Name) {
			downloadURL := fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", file.ID)
			fileDate := extractDateFromFilename(file.Name)

			if fileDate.IsZero() {
				continue
			}

			sources = append(sources, &ExcelDownloadSource{
				URL:        downloadURL,
				FileName:   file.Name,
				UploadDate: fileDate,
			})
		}
	}

	fmt.Printf("Successfully extracted %d sources\n", len(sources))
	return sources, nil
}

func (g *GoogleDriveHelper) GetSourceFromSpreadsheetLink(url string) (*ExcelDownloadSource, error) {
	spreadsheetID := g.extractSpreadsheetID(url)
	if spreadsheetID == "" {
		return nil, fmt.Errorf("could not extract spreadsheet ID from URL: %s", url)
	}

	metadata, err := g.fetchSpreadsheetMetadata(spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("error fetching spreadsheet metadata: %v", err)
	}

	if !g.containsExamKeyword(metadata.Name) {
		return nil, fmt.Errorf("spreadsheet name does not contain exam keywords: %s", metadata.Name)
	}

	downloadURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=xlsx", spreadsheetID)
	fileDate := extractDateFromFilename(metadata.Name)

	if fileDate.IsZero() {
		return nil, fmt.Errorf("could not extract date from filename: %s", metadata.Name)
	}

	return &ExcelDownloadSource{
		URL:        downloadURL,
		FileName:   metadata.Name,
		UploadDate: fileDate,
	}, nil
}

// =====================================
// ======== Private methods ============
// =====================================

func (g *GoogleDriveHelper) extractFolderID(url string) string {
	matches := g.folderIDPattern.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (g *GoogleDriveHelper) extractSpreadsheetID(url string) string {
	matches := g.spreadsheetIDPattern.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (g *GoogleDriveHelper) listFilesInFolder(folderID string) ([]GoogleFile, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files?q='%s'+in+parents&key=%s", 
		folderID, g.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var filesResponse GoogleFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&filesResponse); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return filesResponse.Files, nil
}

func (g *GoogleDriveHelper) fetchSpreadsheetMetadata(spreadsheetID string) (*GoogleFile, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?fields=name&key=%s", 
		spreadsheetID, g.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var file GoogleFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	return &file, nil
}

func (g *GoogleDriveHelper) isExcelFile(filename string) bool {
	return len(filename) > 5 && filename[len(filename)-5:] == ".xlsx"
}

func (g *GoogleDriveHelper) containsExamKeyword(filename string) bool {
	lowerName := strings.ToLower(filename)
	keywords := []string{"examen", "exame", "exam", "horario", "clases"}

	for _, keyword := range keywords {
		if strings.Contains(lowerName, keyword) {
			return true
		}
	}
	return false
}
