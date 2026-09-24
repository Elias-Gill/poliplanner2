package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/logger"
)

var urlHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       5,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	},
}

// URLSource is a source backed by a remote Excel file. It downloads its own
// content lazily and exposes the metadata gathered during discovery.
//
// A single URLSource satisfies both ScheduleSource and LabSource, so the fetching
// stays domain-agnostic and only the parser differs.
type URLSource struct {
	URL        string
	Name       string
	UploadDate time.Time
	Semester   academic.YearSemester
}

func (s *URLSource) Content(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "poliplanner-bot/1.0")

	resp, err := urlHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Info("Failed to download source", "source", s.URL, "status", resp.StatusCode)
		resp.Body.Close()
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (s *URLSource) Metadata() SourceMetadata {
	return SourceMetadata{
		Name:     s.Name,
		URI:      s.URL,
		Semester: s.Semester,
		Date:     s.UploadDate,
	}
}

// NewScheduleSourceFromURL builds a schedule source downloaded from a direct URL.
func NewScheduleSourceFromURL(url, name string, semester academic.YearSemester, date time.Time) ScheduleSource {
	return &URLSource{URL: url, Name: name, Semester: semester, UploadDate: date}
}

// NewLabSourceFromURL builds a laboratory source downloaded from a direct URL.
func NewLabSourceFromURL(url, name string, semester academic.YearSemester, date time.Time) LabSource {
	return &URLSource{URL: url, Name: name, Semester: semester, UploadDate: date}
}
