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

var (
	CheckLastSyncError = errors.New("Failed to retrieve last sync date")
)

type SyncService struct {
	// External services
	importService *DiscoveryService
	excelService  *ExcelService

	// Repositories
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

func (e SyncService) AutoSync(ctx context.Context) error {
	logger.Info("Auto sync check started")

	// Buscar la ultima sincronizacion realizada
	lastCheck, err := e.syncRepository.GetLastSyncAttempt(ctx)
	if err != nil {
		logger.Warn("Failed to retrieve last checked time", "error", err)
		return CheckLastSyncError
	}

	if lastCheck == nil {
		logger.Info("No previous sync check found")
		return e.Sync(ctx)
	}

	elapsed := time.Since(*lastCheck)
	logger.Info("Time since last check", "elapsed_hours", math.Round(elapsed.Hours()))

	if elapsed >= autoSyncInterval {
		return e.Sync(ctx)
	}

	logger.Info("Sync not required")

	return nil
}

func (e SyncService) Sync(ctx context.Context) error {
	logger.Info("Starting Excel sync")

	// Fetch the most recent sources from the web
	webSources, err := e.importService.FindLatestSources(ctx)
	if err != nil {
		logger.Error("Error retrieving latest source from web", "error", err)
		return fmt.Errorf("error retrieving latest source from web: %w", err)
	}

	if webSources == nil {
		logger.Error("No sources found in the web")
		return fmt.Errorf("No sources found in the web")
	}

	// Get the newest version already stored in the database
	serverVersion, err := e.excelService.GetLatestValidVersion(ctx)
	if err != nil && err != ErrNoSheetVersion {
		logger.Error("Failed to get newest version from database", "error", err)
		return fmt.Errorf("error retrieving latest version from db: %w", err)
	}

	// Record the sync attempt timestamp
	err = e.syncRepository.SetLastSyncAttempt(ctx, time.Now().In(timezone.ParaguayTZ))
	if err != nil {
		logger.Error("Failed to set sync date on database", "error", err)
		return fmt.Errorf("error setting sync date on database: %w", err)
	}

	// If no version exists in DB, perform initial import of all sources
	if serverVersion == nil {
		logger.Info("No previous version found in database, starting initial import")
		return e.persistAllSources(ctx, webSources.Sources)
	}

	// Compare the overall latest date from web with the DB's latest parsed date
	if !webSources.Date.After(serverVersion.ParsedAt) {
		logger.Info(
			"Current excel source is up to date",
			"source_date", webSources.Date,
			"db_parsed_at", serverVersion.ParsedAt,
		)
		return nil
	}

	logger.Info(
		"Newer excel sources found, starting import",
		"date", webSources.Date,
		"db_parsed_at", serverVersion.ParsedAt,
		"count", len(webSources.Sources),
	)

	// Persist all sources from the web (loop over each source)
	return e.persistAllSources(ctx, webSources.Sources)
}

// persistAllSources iterates over the given sources and persists each one.
// It aggregates any errors and returns a combined error if at least one fails.
func (e SyncService) persistAllSources(ctx context.Context, sources []source.ExcelSource) error {
	var errs []error
	for i, src := range sources {
		logger.Info("Persisting source", "index", i, "uri", src.Metadata().URI)
		if err := e.excelService.PersistSource(ctx, src); err != nil {
			logger.Error("Failed to persist source", "index", i, "error", err)
			errs = append(errs, fmt.Errorf("source %d: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors persisting sources: %w", errors.Join(errs...))
	}
	return nil
}
