package excel

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/repository/excel"
	"github.com/elias-gill/poliplanner2/logger"
)

const autoSyncInterval = 8 * time.Hour

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

	// Buscar la version mas nueva en la web
	webSource, err := e.importService.FindLatestSource(ctx)
	if err != nil {
		logger.Error("Error retrieving latest source from web", "error", err)
		return fmt.Errorf("error retrieving latest source from web: %w", err)
	}

	// Obtener la version mas nueva del server
	serverVersion, err := e.excelService.GetLatestValidVersion(ctx)
	if err != nil && err != ErrNoSheetVersion {
		logger.Error("Failed to get newest version from database", "error", err)
		return fmt.Errorf("error retrieving latest version from db: %w", err)
	}

	// Actualizar la fecha de sincronizacion del sistema
	err = e.syncRepository.SetLastSyncAttempt(ctx, time.Now().In(timezone.ParaguayTZ))
	if err != nil {
		logger.Error("Failed to set sync date on database", "error", err)
		return fmt.Errorf("error setting sync date on database: %w", err)
	}

	if serverVersion == nil {
		logger.Info("No previous version found in database, starting initial import")
		return e.excelService.PersistSource(ctx, webSource)
	}

	// Comparar la version de la web con la version del server mas nueva
	if !webSource.Metadata().Date.After(serverVersion.ParsedAt) {
		logger.Info(
			"Current excel source is up to date",
			"source_date", webSource.Metadata().Date,
			"db_parsed_at", serverVersion.ParsedAt,
		)
		return nil
	}

	logger.Info(
		"Newer excel source found, starting import",
		"source_date", webSource.Metadata().Date,
		"db_parsed_at", serverVersion.ParsedAt,
	)

	return e.excelService.PersistSource(ctx, webSource)
}
