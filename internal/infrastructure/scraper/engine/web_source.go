package engine

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

// WebSource is a Source backed by a remote Excel file. It knows how to download
// its own content and exposes the metadata gathered during discovery.
//
// A single WebSource satisfies both source.ScheduleSource and source.LabSource,
// which is why there is no need for per-domain wrappers: the fetching is
// domain-agnostic and only the parser differs.
type WebSource struct {
	URL        string
	Name       string
	UploadDate time.Time
	Semester   academic.YearSemester
}

func (s *WebSource) Content(ctx context.Context) (io.ReadCloser, error) {
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

func (s *WebSource) Metadata() source.SourceMetadata {
	return source.SourceMetadata{
		Name:     s.Name,
		URI:      s.URL,
		Semester: s.Semester,
		Date:     s.UploadDate,
	}
}
