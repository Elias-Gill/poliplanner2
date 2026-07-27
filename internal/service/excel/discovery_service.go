package excel

import (
	"context"
	"fmt"

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

func (i DiscoveryService) FindLatestSource(ctx context.Context) (source.ExcelSource, error) {
	if i.scraper == nil {
		return nil, fmt.Errorf("error searching for Excel versions: web scraper not initialized")
	}

	sources, err := i.scraper.Discover(ctx)
	if err != nil {
		logger.Error("Web scraper failed to find latest source", "error", err)
		return nil, fmt.Errorf("error searching for Excel versions: %w", err)
	}

	var newest source.ExcelSource = nil
	for _, s := range sources {
		if newest == nil || s.Metadata().Date.After(newest.Metadata().Date) {
			newest = s
		}
	}

	return newest, nil
}
