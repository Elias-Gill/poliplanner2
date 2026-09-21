package excel

import (
	"context"
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper/engine"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	"github.com/elias-gill/poliplanner2/logger"
)

// ======================================
// =            Public API              =
// ======================================

type DiscoveryService struct {
	scheduleScraper *scraper.ScheduleScraper
	labScraper      *scraper.LabScraper
}

func NewDiscoveryService(googleApikey string) *DiscoveryService {
	driveHelper := engine.NewGoogleDriveHelper(googleApikey)
	return &DiscoveryService{
		scheduleScraper: scraper.NewScheduleScraper(driveHelper, ""),
		labScraper:      scraper.NewLabScraper(driveHelper, ""),
	}
}

func (i DiscoveryService) FindLatestScheduleSources(ctx context.Context) (*scheduleSourcesResult, error) {
	if i.scheduleScraper == nil {
		return nil, fmt.Errorf("error searching for Excel versions: web scraper not initialized")
	}

	sources, err := i.scheduleScraper.Discover(ctx)
	if err != nil {
		logger.Error("Web scraper failed to find latest source", "error", err)
		return nil, fmt.Errorf("error searching for Excel versions: %w", err)
	}

	if len(sources) == 0 {
		logger.Warn("No sources found by scraper")
		return nil, nil
	}

	var latestSources []source.ScheduleSource
	var tracker latestDayTracker

	for _, s := range sources {
		meta := s.Metadata()

		newer, same := tracker.Add(meta.Date)
		switch {
		case newer:
			latestSources = []source.ScheduleSource{s}
			logger.Info("Source from a newer day found", "name", meta.Name, "uri", meta.URI, "date", meta.Date)
		case same:
			latestSources = append(latestSources, s)
			logger.Info("Source matching latest day found", "name", meta.Name, "uri", meta.URI, "date", meta.Date)
		}
	}

	logger.Info("Finished filtering latest sources",
		"total_scraped", len(sources),
		"latest_count", len(latestSources),
		"newest_date", tracker.newestDate,
	)

	return &scheduleSourcesResult{
		Sources: latestSources,
		Date:    tracker.newestDate,
	}, nil
}

func (i DiscoveryService) FindLatestLabSources(ctx context.Context) (*labSourcesResult, error) {
	if i.labScraper == nil {
		return nil, fmt.Errorf("error searching for lab Excel versions: web scraper not initialized")
	}

	sources, err := i.labScraper.Discover(ctx)
	if err != nil {
		logger.Error("Web scraper failed to find latest laboratory source", "error", err)
		return nil, fmt.Errorf("error searching for lab Excel versions: %w", err)
	}

	if len(sources) == 0 {
		logger.Warn("No laboratory sources found by scraper")
		return nil, nil
	}

	var latestSources []source.LabSource
	var tracker latestDayTracker

	for _, s := range sources {
		meta := s.Metadata()

		newer, same := tracker.Add(meta.Date)
		switch {
		case newer:
			latestSources = []source.LabSource{s}
			logger.Info("Laboratory source from a newer day found", "name", meta.Name, "uri", meta.URI, "date", meta.Date)
		case same:
			latestSources = append(latestSources, s)
			logger.Info("Laboratory source matching latest day found", "name", meta.Name, "uri", meta.URI, "date", meta.Date)
		}
	}

	logger.Info("Finished filtering latest laboratory sources",
		"total_scraped", len(sources),
		"latest_count", len(latestSources),
		"newest_date", tracker.newestDate,
	)

	return &labSourcesResult{
		Sources: latestSources,
		Date:    tracker.newestDate,
	}, nil
}

// ======================================
// =               Utils                =
// ======================================

type scheduleSourcesResult struct {
	Sources []source.ScheduleSource
	Date    time.Time
}

type labSourcesResult struct {
	Sources []source.LabSource
	Date    time.Time
}

// ======================================
// =         Private methods            =
// ======================================

// latestDayTracker keeps track of the most recent calendar day while iterating
// over discovered sources. It only looks at the source dates, so it is shared by
// every source type.
//
// Add classifies a source date against the newest one seen so far:
//   - newer: the source belongs to a more recent day, so the caller must start a
//     fresh batch with it.
//   - same: the source shares the current newest day, so the caller must append
//     it to the current batch.
type latestDayTracker struct {
	newestDate time.Time
	seen       bool
}

func (t *latestDayTracker) Add(currentDate time.Time) (newer, same bool) {
	// Strip the time component to compare calendar days only.
	currentDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
	newestDay := time.Date(t.newestDate.Year(), t.newestDate.Month(), t.newestDate.Day(), 0, 0, 0, 0, t.newestDate.Location())

	if !t.seen || currentDay.After(newestDay) {
		t.seen = true
		t.newestDate = currentDate
		return true, false
	}

	if currentDay.Equal(newestDay) {
		// Keep the exact latest timestamp of the day.
		if currentDate.After(t.newestDate) {
			t.newestDate = currentDate
		}
		return false, true
	}

	return false, false
}
