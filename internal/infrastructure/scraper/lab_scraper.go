package scraper

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper/engine"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
)

// LabScraper discovers the laboratory practice Excel sources published by the
// university. It is completely independent from ScheduleScraper: it has its own
// filter and returns its own source type, while sharing the generic
// ScrapeEngine.
type LabScraper struct {
	engine *engine.ScrapeEngine
}

func NewLabScraper(drive engine.DriveHelper, targetURL string) *LabScraper {
	return &LabScraper{
		engine: engine.NewScrapeEngine(drive, targetURL),
	}
}

func (l *LabScraper) Discover(ctx context.Context) ([]source.LabSource, error) {
	raw, err := l.engine.Discover(ctx, laboratoryFilter)
	if err != nil {
		return nil, err
	}

	return toLabSources(raw), nil
}

// FindSourcesFromHTML parses an in-memory page without hitting the network. It
// is intended for tests.
func (l *LabScraper) FindSourcesFromHTML(ctx context.Context, htmlContent string) ([]source.LabSource, error) {
	raw, err := l.engine.FindSourcesFromHTML(ctx, htmlContent, laboratoryFilter)
	if err != nil {
		return nil, err
	}

	return toLabSources(raw), nil
}

func toLabSources(raw []*engine.WebSource) []source.LabSource {
	out := make([]source.LabSource, len(raw))
	for i, r := range raw {
		out[i] = r
	}
	return out
}
