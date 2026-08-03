package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/logger"
)

type webSource struct {
	URL        string
	Name       string
	UploadDate time.Time
	Semester   academic.YearSemester
}

type WebScheduleSource webSource

type WebLaboratorySource webSource

func (s *webSource) Content(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "poliplanner-bot/1.0")

	resp, err := httpClient.Do(req)
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

func (s *webSource) Metadata() source.SourceMetadata {
	return source.SourceMetadata{
		Name:     s.Name,
		URI:      s.URL,
		Semester: s.Semester,
		Date:     s.UploadDate,
	}
}

func (s *WebScheduleSource) Content(ctx context.Context) (io.ReadCloser, error) {
	return ((*webSource)(s)).Content(ctx)
}

func (s *WebScheduleSource) Metadata() source.SourceMetadata {
	return ((*webSource)(s)).Metadata()
}

func (s *WebLaboratorySource) Content(ctx context.Context) (io.ReadCloser, error) {
	return ((*webSource)(s)).Content(ctx)
}

func (s *WebLaboratorySource) Metadata() source.SourceMetadata {
	return ((*webSource)(s)).Metadata()
}
