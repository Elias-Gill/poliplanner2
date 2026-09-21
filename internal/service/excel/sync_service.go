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

func (s *SyncService) GetLastSyncAttempt(ctx context.Context) (*time.Time, error) {
	return s.syncRepository.GetLastSyncAttempt(ctx)
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

// Sync synchronizes every published Excel source type. Schedule and laboratory
// sources share the same last sync attempt marker: there is no separate check
// per type.
func (s *SyncService) Sync(ctx context.Context) error {
	logger.Info("Starting sources sync")

	var errs []error

	if err := s.syncSchedules(ctx); err != nil {
		logger.Error("Schedule sources sync failed", "error", err)
		errs = append(errs, err)
	}

	if err := s.syncLabs(ctx); err != nil {
		logger.Error("Laboratory sources sync failed", "error", err)
		errs = append(errs, err)
	}

	if err := s.syncRepository.SetLastSyncAttempt(ctx, time.Now().In(timezone.ParaguayTZ)); err != nil {
		logger.Error("Failed to set sync date on database", "error", err)
		errs = append(errs, fmt.Errorf("error setting sync date on database: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *SyncService) syncSchedules(ctx context.Context) error {
	logger.Info("Starting schedule sources sync")

	webSources, err := s.importService.FindLatestScheduleSources(ctx)
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

	if serverVersion == nil {
		logger.Info("No previous version found in database, persisting all latest web sources", "count", len(webSources.Sources))
		return s.persistAllScheduleSources(ctx, webSources.Sources)
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

	return s.persistAllScheduleSources(ctx, webSources.Sources)
}

func (s *SyncService) syncLabs(ctx context.Context) error {
	logger.Info("Starting laboratory sources sync")

	labSources, err := s.importService.FindLatestLabSources(ctx)
	if err != nil {
		logger.Error("Error retrieving latest laboratory sources from web", "error", err)
		return fmt.Errorf("error retrieving latest laboratory sources from web: %w", err)
	}

	if labSources == nil || len(labSources.Sources) == 0 {
		logger.Warn("No laboratory sources found on web")
		return nil
	}

	return s.persistAllLabSources(ctx, labSources.Sources)
}

func (s *SyncService) persistAllScheduleSources(ctx context.Context, sources []source.ScheduleSource) error {
	var errs []error
	for i, src := range sources {
		logger.Info("Persisting source", "index", i, "uri", src.Metadata().URI, "name", src.Metadata().Name)
		if err := s.excelService.PersistScheduleSource(ctx, src); err != nil {
			logger.Error("Failed to persist source", "index", i, "uri", src.Metadata().URI, "error", err)
			errs = append(errs, fmt.Errorf("source %d (%s): %w", i, src.Metadata().URI, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors persisting sources: %w", errors.Join(errs...))
	}

	return nil
}

func (s *SyncService) persistAllLabSources(ctx context.Context, sources []source.LabSource) error {
	var errs []error
	for i, src := range sources {
		logger.Info("Persisting laboratory source", "index", i, "uri", src.Metadata().URI, "name", src.Metadata().Name)
		if err := s.excelService.PersistLabSource(ctx, src); err != nil {
			logger.Error("Failed to persist laboratory source", "index", i, "uri", src.Metadata().URI, "error", err)
			errs = append(errs, fmt.Errorf("laboratory source %d (%s): %w", i, src.Metadata().URI, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors persisting laboratory sources: %w", errors.Join(errs...))
	}

	return nil
}
