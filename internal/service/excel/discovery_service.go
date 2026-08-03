package excel

import (
	"context"
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	"github.com/elias-gill/poliplanner2/logger"
)

type DiscoveryService struct {
	scraper *scraper.WebScrapper
}

func NewDiscoveryService(googleApikey string) *DiscoveryService {
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(googleApikey))
	return &DiscoveryService{
		scraper: scraper,
	}
}

type sourcesResult struct {
	Sources []source.ScheduleSource
	Date    time.Time
}

func (i DiscoveryService) FindLatestSources(ctx context.Context) (*sourcesResult, error) {
	if i.scraper == nil {
		return nil, fmt.Errorf("error searching for Excel versions: web scraper not initialized")
	}

	sources, err := i.scraper.DiscoverSchedules(ctx)
	if err != nil {
		logger.Error("Web scraper failed to find latest source", "error", err)
		return nil, fmt.Errorf("error searching for Excel versions: %w", err)
	}

	if len(sources) == 0 {
		logger.Warn("No sources found by scraper")
		return nil, nil
	}
	var latestSources []source.ScheduleSource
	var newestDate time.Time

	for _, s := range sources {
		meta := s.Metadata()
		currentDate := meta.Date

		if len(latestSources) == 0 || currentDate.After(newestDate) {
			newestDate = currentDate
			latestSources = []source.ScheduleSource{s}
			logger.Info("Newer source date found", "name", meta.Name, "uri", meta.URI, "date", currentDate)
		} else if currentDate.Equal(newestDate) {
			latestSources = append(latestSources, s)
			logger.Info("Source with matching latest date found", "name", meta.Name, "uri", meta.URI, "date", currentDate)
		}
	}

	logger.Info("Finished filtering latest sources",
		"total_scraped", len(sources),
		"latest_count", len(latestSources),
		"newest_date", newestDate,
	)

	return &sourcesResult{
		Sources: latestSources,
		Date:    newestDate,
	}, nil
}
