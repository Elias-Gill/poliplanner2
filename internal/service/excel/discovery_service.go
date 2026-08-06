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

		// strip time component to compare calendar days only
		currDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
		newestDay := time.Date(newestDate.Year(), newestDate.Month(), newestDate.Day(), 0, 0, 0, 0, newestDate.Location())

		if len(latestSources) == 0 || currDay.After(newestDay) {
			// Found a source from a newer day
			newestDate = currentDate
			latestSources = []source.ScheduleSource{s}
			logger.Info("Source from a newer day found", "name", meta.Name, "uri", meta.URI, "date", currentDate)

		} else if currDay.Equal(newestDay) {
			// Source modified on the same day are added to the batch
			latestSources = append(latestSources, s)

			// Keep the exact latest timestamp of the day
			if currentDate.After(newestDate) {
				newestDate = currentDate
			}
			logger.Info("Source matching latest day found", "name", meta.Name, "uri", meta.URI, "date", currentDate)
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
