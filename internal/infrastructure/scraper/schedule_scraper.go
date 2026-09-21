package scraper

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper/engine"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
)

// ScheduleScraper discovers the class/exam schedule Excel sources published by
// the university. It is a thin, domain-specific wrapper around the shared
// ScrapeEngine: it only knows which files are schedules and which source type to
// return.
type ScheduleScraper struct {
	engine *engine.ScrapeEngine
}

func NewScheduleScraper(drive engine.DriveHelper, targetURL string) *ScheduleScraper {
	return &ScheduleScraper{
		engine: engine.NewScrapeEngine(drive, targetURL),
	}
}

func (s *ScheduleScraper) Discover(ctx context.Context) ([]source.ScheduleSource, error) {
	raw, err := s.engine.Discover(ctx, scheduleFilter)
	if err != nil {
		return nil, err
	}

	return toScheduleSources(raw), nil
}

// FindSourcesFromHTML parses an in-memory page without hitting the network. It
// is intended for tests.
func (s *ScheduleScraper) FindSourcesFromHTML(ctx context.Context, htmlContent string) ([]source.ScheduleSource, error) {
	raw, err := s.engine.FindSourcesFromHTML(ctx, htmlContent, scheduleFilter)
	if err != nil {
		return nil, err
	}

	return toScheduleSources(raw), nil
}

func toScheduleSources(raw []*engine.WebSource) []source.ScheduleSource {
	out := make([]source.ScheduleSource, len(raw))
	for i, r := range raw {
		out[i] = r
	}
	return out
}
