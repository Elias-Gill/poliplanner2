package excel

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	"github.com/elias-gill/poliplanner2/internal/model/excel"
	excelRepo "github.com/elias-gill/poliplanner2/internal/repository/excel"
	"github.com/elias-gill/poliplanner2/logger"
)

// autoSyncInterval is how long a source type waits before searching the web
// again. The last search timestamp is stored per type in the database.
const autoSyncInterval = 6 * time.Hour

var ErrCheckLastSync = errors.New("failed to retrieve last sync date")

type SyncService struct {
	importService  *DiscoveryService
	excelService   *ExcelService
	syncRepository excelRepo.SyncRepository
}

func NewSyncService(
	discvSrv *DiscoveryService,
	excelService *ExcelService,
	syncRepo excelRepo.SyncRepository,
) *SyncService {
	return &SyncService{
		importService:  discvSrv,
		excelService:   excelService,
		syncRepository: syncRepo,
	}
}

func (s *SyncService) GetSyncState(ctx context.Context, kind excel.SourceType) (*excel.SyncState, error) {
	return s.syncRepository.GetSyncState(ctx, kind)
}

// AutoSync searches and syncs each source type whose last search is older than
// autoSyncInterval.
func (s *SyncService) AutoSync(ctx context.Context) error {
	logger.Info("Auto sync check started")

	var errs []error

	// Schedules
	searchSchedules, err := s.shouldSearch(ctx, excel.SourceTypeSchedule)
	if err != nil {
		logger.Warn("Failed to retrieve schedule sync state", "error", err)
		errs = append(errs, ErrCheckLastSync)
	} else if searchSchedules {
		// Sync schedules if a new version is available
		err := s.syncSchedules(ctx)
		if err != nil {
			logger.Error("Schedule sources sync failed", "error", err)
			errs = append(errs, err)
		}
	}

	// Laboratories
	searchLabs, err := s.shouldSearch(ctx, excel.SourceTypeLab)
	if err != nil {
		logger.Warn("Failed to retrieve laboratory sync state", "error", err)
		errs = append(errs, ErrCheckLastSync)
	} else if searchLabs {
		// Sync labs if a new version is available
		err := s.syncLabs(ctx)
		if err != nil {
			logger.Error("Laboratory sources sync failed", "error", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// shouldSearch reports whether enough time passed since the last search of a
// source type to search the web again.
func (s *SyncService) shouldSearch(ctx context.Context, kind excel.SourceType) (bool, error) {
	state, err := s.syncRepository.GetSyncState(ctx, kind)
	if err != nil {
		return false, err
	}

	if state.LastSearchAt == nil {
		return true, nil
	}

	elapsed := time.Since(*state.LastSearchAt)
	if elapsed >= autoSyncInterval {
		return true, nil
	}

	logger.Info("Source search not required", "source_type", kind, "elapsed_hours", math.Round(elapsed.Hours()))

	return false, nil
}

// Sync forces a search and sync of both source types, ignoring the interval.
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

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *SyncService) syncSchedules(ctx context.Context) error {
	logger.Info("Starting schedule sources sync")

	webSources, err := s.importService.FindLatestScheduleSources(ctx)
	if err != nil {
		logger.Error("Error retrieving latest schedule sources from web", "error", err)
		return fmt.Errorf("error retrieving latest schedule sources from web: %w", err)
	}

	s.recordSearchAttempt(ctx, excel.SourceTypeSchedule)

	if webSources == nil || len(webSources.Sources) == 0 {
		logger.Warn("No schedule sources found on web")
		return nil
	}

	pending, err := s.pendingScheduleSources(ctx, webSources.Sources)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		logger.Info("All schedule sources are up to date")
		return nil
	}

	logger.Info("New schedule sources found, starting import", "count", len(pending))

	if err := s.persistAllScheduleSources(ctx, pending); err != nil {
		return err
	}

	return s.recordSyncAttempt(ctx, excel.SourceTypeSchedule)
}

func (s *SyncService) syncLabs(ctx context.Context) error {
	logger.Info("Starting laboratory sources sync")

	labSources, err := s.importService.FindLatestLabSources(ctx)
	if err != nil {
		logger.Error("Error retrieving latest laboratory sources from web", "error", err)
		return fmt.Errorf("error retrieving latest laboratory sources from web: %w", err)
	}

	s.recordSearchAttempt(ctx, excel.SourceTypeLab)

	if labSources == nil || len(labSources.Sources) == 0 {
		logger.Warn("No laboratory sources found on web")
		return nil
	}

	pending, err := s.pendingLabSources(ctx, labSources.Sources)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		logger.Info("All laboratory sources are up to date")
		return nil
	}

	logger.Info("New laboratory sources found, starting import", "count", len(pending))

	if err := s.persistAllLabSources(ctx, pending); err != nil {
		return err
	}

	return s.recordSyncAttempt(ctx, excel.SourceTypeLab)
}

func (s *SyncService) pendingScheduleSources(ctx context.Context, sources []source.ScheduleSource) ([]source.ScheduleSource, error) {
	pending := make([]source.ScheduleSource, 0, len(sources))

	for _, src := range sources {
		meta := src.Metadata()

		upToDate, err := s.excelService.IsSourceUpToDate(ctx, excel.SourceTypeSchedule, meta.Name, meta.URI, meta.Date)
		if err != nil {
			return nil, fmt.Errorf("error checking schedule source '%s': %w", meta.Name, err)
		}

		if !upToDate {
			pending = append(pending, src)
		}
	}

	return pending, nil
}

func (s *SyncService) pendingLabSources(ctx context.Context, sources []source.LabSource) ([]source.LabSource, error) {
	pending := make([]source.LabSource, 0, len(sources))

	for _, src := range sources {
		meta := src.Metadata()

		upToDate, err := s.excelService.IsSourceUpToDate(ctx, excel.SourceTypeLab, meta.Name, meta.URI, meta.Date)
		if err != nil {
			return nil, fmt.Errorf("error checking laboratory source '%s': %w", meta.Name, err)
		}

		if !upToDate {
			pending = append(pending, src)
		}
	}

	return pending, nil
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

func (s *SyncService) recordSearchAttempt(ctx context.Context, kind excel.SourceType) {
	if err := s.syncRepository.SetLastSearchAttempt(ctx, kind, time.Now().In(timezone.ParaguayTZ)); err != nil {
		logger.Error("Failed to set last search attempt", "source_type", kind, "error", err)
	}
}

func (s *SyncService) recordSyncAttempt(ctx context.Context, kind excel.SourceType) error {
	if err := s.syncRepository.SetLastSyncAttempt(ctx, kind, time.Now().In(timezone.ParaguayTZ)); err != nil {
		logger.Error("Failed to set last sync attempt", "source_type", kind, "error", err)
		return fmt.Errorf("error setting last sync attempt: %w", err)
	}

	return nil
}
