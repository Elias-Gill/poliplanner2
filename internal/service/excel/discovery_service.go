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
	Sources []source.ExcelSource
	Date    time.Time
}

func (i DiscoveryService) FindLatestSources(ctx context.Context) (*sourcesResult, error) {
	if i.scraper == nil {
		return nil, fmt.Errorf("error searching for Excel versions: web scraper not initialized")
	}

	sources, err := i.scraper.Discover(ctx)
	if err != nil {
		logger.Error("Web scraper failed to find latest source", "error", err)
		return nil, fmt.Errorf("error searching for Excel versions: %w", err)
	}

	if len(sources) == 0 {
		return nil, nil
	}

	var latestSources []source.ExcelSource
	var newestDate time.Time

	for _, s := range sources {
		currentDate := s.Metadata().Date

		if len(latestSources) == 0 || currentDate.After(newestDate) {
			// A newer date is found
			newestDate = currentDate
			latestSources = []source.ExcelSource{s}
		} else if currentDate.Equal(newestDate) {
			// Same date as the newest date, then append the source
			latestSources = append(latestSources, s)
		}
	}

	return &sourcesResult{
		Sources: latestSources,
		Date:    newestDate,
	}, nil
}
