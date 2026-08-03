package excel

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	"github.com/elias-gill/poliplanner2/internal/repository/excel"
	"github.com/elias-gill/poliplanner2/logger"
)

const autoSyncInterval = 6 * time.Hour

var ErrCheckLastSync = errors.New("failed to retrieve last sync date")

type SyncService struct {
	importService  *DiscoveryService
	excelService   *ExcelService
	syncRepository excel.SyncRepository
}

func NewSyncService(
	discvSrv *DiscoveryService,
	excelService *ExcelService,
	syncRepo excel.SyncRepository,
) *SyncService {
	return &SyncService{
		importService:  discvSrv,
		excelService:   excelService,
		syncRepository: syncRepo,
	}
}

func (s *SyncService) AutoSync(ctx context.Context) error {
	logger.Info("Auto sync check started")

	lastCheck, err := s.syncRepository.GetLastSyncAttempt(ctx)
	if err != nil {
		logger.Warn("Failed to retrieve last checked time", "error", err)
		return ErrCheckLastSync
	}

	if lastCheck == nil {
		logger.Info("No previous sync check found, executing sync")
		return s.Sync(ctx)
	}

	elapsed := time.Since(*lastCheck)
	logger.Info("Time since last check", "elapsed_hours", math.Round(elapsed.Hours()))

	if elapsed >= autoSyncInterval {
		return s.Sync(ctx)
	}

	logger.Info("Auto sync not required")
	return nil
}

func (s *SyncService) Sync(ctx context.Context) error {
	logger.Info("Starting schedule sources sync")

	webSources, err := s.importService.FindLatestSources(ctx)
	if err != nil {
		logger.Error("Error retrieving latest sources from web", "error", err)
		return fmt.Errorf("error retrieving latest sources from web: %w", err)
	}

	if webSources == nil || len(webSources.Sources) == 0 {
		logger.Error("No schedule sources found on web")
		return fmt.Errorf("no schedule sources found on web")
	}

	serverVersion, err := s.excelService.GetLatestValidVersion(ctx)
	if err != nil && !errors.Is(err, ErrNoSheetVersion) {
		logger.Error("Failed to get newest version from database", "error", err)
		return fmt.Errorf("error retrieving latest version from db: %w", err)
	}

	err = s.syncRepository.SetLastSyncAttempt(ctx, time.Now().In(timezone.ParaguayTZ))
	if err != nil {
		logger.Error("Failed to set sync date on database", "error", err)
		return fmt.Errorf("error setting sync date on database: %w", err)
	}

	if serverVersion == nil {
		logger.Info("No previous version found in database, persisting all latest web sources", "count", len(webSources.Sources))
		return s.persistAllSources(ctx, webSources.Sources)
	}

	if !webSources.Date.After(serverVersion.ParsedAt) {
		logger.Info(
			"Current excel source is up to date",
			"web_source_date", webSources.Date,
			"db_source_date", serverVersion.ParsedAt,
		)
		return nil
	}

	logger.Info(
		"Newer excel sources found, starting import",
		"web_source_date", webSources.Date,
		"db_source_date", serverVersion.ParsedAt,
		"count", len(webSources.Sources),
	)

	return s.persistAllSources(ctx, webSources.Sources)
}

func (s *SyncService) persistAllSources(ctx context.Context, sources []source.ScheduleSource) error {
	var errs []error
	for i, src := range sources {
		logger.Info("Persisting source", "index", i, "uri", src.Metadata().URI, "name", src.Metadata().Name)
		if err := s.excelService.PersistSource(ctx, src); err != nil {
			logger.Error("Failed to persist source", "index", i, "uri", src.Metadata().URI, "error", err)
			errs = append(errs, fmt.Errorf("source %d (%s): %w", i, src.Metadata().URI, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors persisting sources: %w", errors.Join(errs...))
	}

	return nil
}
