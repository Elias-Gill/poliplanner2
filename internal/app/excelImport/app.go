package excelimport

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	sheetversion "github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser/metadata"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	"github.com/elias-gill/poliplanner2/logger"
)

type ExcelImporter struct {
	sheetVersionStorer sheetversion.SheetVersionRepository
	importStorer       ImportRepository
}

func New(
	importStorer ImportRepository,
	sheetVersionStorer sheetversion.SheetVersionRepository,
) *ExcelImporter {
	return &ExcelImporter{
		importStorer:       importStorer,
		sheetVersionStorer: sheetVersionStorer,
	}
}

// ================================
// =         Public API           =
// ================================

// PeriodicSync checks if the last sync was 48hs ago to perform a search and parsing
func (s *ExcelImporter) AutoSync(ctx context.Context) {
	if !s.shouldAutoSync(ctx) {
		logger.Info("Excel auto-sync skipped: last check was less than 48 hours ago")
		return
	}

	logger.Info("Starting automatic Excel sync")
	s.sheetVersionStorer.SetLastCheckedDate(ctx, time.Now().In(timezone.ParaguayTZ))

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := s.syncNewestVersion(ctx); err != nil {
		logger.Error("Automatic Excel sync failed", "error", err)
		return
	}

	logger.Info("Automatic Excel sync completed successfully")
}

// Sync ignores the 48hs threshold but still checks if the web source is newer
// before parsing it.
func (s *ExcelImporter) Sync(ctx context.Context) error {
	logger.Info("Manual Excel sync requested")

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	err := s.syncNewestVersion(ctx)
	if err != nil {
		logger.Error("Manual Excel sync failed", "error", err)
	} else {
		logger.Info("Manual Excel sync completed")
	}

	return err
}

// PersistSource parses and persists an ExcelSource into persistent storage
func (s *ExcelImporter) PersistSource(ctx context.Context, src source.ExcelSource) error {
	logger.Info("Starting persistence of Excel source")

	sourceMeta := src.GetMetadata()
	logger.Debug("Source metadata", "name", sourceMeta.Name, "uri", sourceMeta.URI, "date", sourceMeta.Date)

	content, err := src.GetContent(ctx)
	if err != nil {
		logger.Error("Failed to open Excel source content", "uri", sourceMeta.URI, "error", err)
		return fmt.Errorf("cannot open Excel source: %w", err)
	}
	defer content.Close()

	parserExcel, err := parser.NewExcelParser(content)
	if err != nil {
		logger.Error("Failed to create Excel parser", "error", err)
		return fmt.Errorf("error creating Excel parser: %w", err)
	}
	defer parserExcel.Close()

	logger.Debug("Excel parser initialized successfully")

	var pID period.PeriodID = 0
	sheetCount := 0

	pError := s.importStorer.RunImport(ctx, func(writer ImportWriter) error {
		// Search or create a new period based on the source data
		periodID, err := writer.EnsurePeriod(period.NewPeriodFromTime(sourceMeta.Date))
		if err != nil {
			logger.Error("Failed to ensure period exists", "date", sourceMeta.Date, "error", err)
			return fmt.Errorf("cannot save current period: %w", err)
		}
		logger.Debug("Period ensured", "period_id", periodID)
		pID = periodID

		for parserExcel.NextSheet() {
			sheetCount++
			logger.Debug("Processing sheet", "sheet_number", sheetCount)

			sheetResult, err := parserExcel.ParseCurrentSheet()
			if err != nil {
				logger.Error("Failed to parse current sheet", "career", sourceMeta.Name, "error", err)
				return err
			}

			logger.Info("Sheet parsed successfully", "career", sheetResult.Name, "subjects_count", len(sheetResult.Subjects))

			err = s.persistSheetSubjects(sheetResult, periodID, writer)
			if err != nil {
				logger.Error("Failed to persist sheet subjects", "career", sheetResult.Name, "error", err)
				return err
			}

			logger.Info("Sheet subjects persisted successfully", "career", sheetResult.Name)
		}

		logger.Info("Excel import completed successfully",
			"file", sourceMeta.Name,
			"url", sourceMeta.URI,
			"sheets_processed", sheetCount,
		)
		return nil
	})

	success := pError == nil
	now := time.Now().In(timezone.ParaguayTZ)
	version, err := sheetversion.NewSheetVersion(
		pID,
		sourceMeta.Name,
		sourceMeta.URI,
		now,
		sheetCount,
		pError,
	)
	if err != nil {
		logger.Error("Failed to create audit struct", "uri", sourceMeta.URI, "error", err)
		return fmt.Errorf("failed to create Excel audit summary: %w", err)
	}

	_, err = s.importStorer.SaveAudit(ctx, version)
	if err != nil {
		logger.Error("Failed to save import audit", "uri", sourceMeta.URI, "error", err)
		return fmt.Errorf("failed to save Excel audit summary: %w", err)
	}

	if success {
		logger.Info("Import audit saved successfully")
	}

	return pError
}

// ================================
// =       Private helpers        =
// ================================

func (s *ExcelImporter) shouldAutoSync(ctx context.Context) bool {
	lastCheck, err := s.sheetVersionStorer.GetLastCheckedDate(ctx)
	if err != nil {
		logger.Warn("Failed to retrieve last checked time", "error", err)
		return true // si hay error, mejor intentar sincronizar
	}

	if lastCheck == nil {
		logger.Debug("No previous sync timestamp found - auto-sync will run")
		return true
	}

	elapsed := time.Since(*lastCheck)
	logger.Info("Time since last check", "elapsed_hours", math.Round(elapsed.Hours()))

	return elapsed >= 48*time.Hour
}

func (s *ExcelImporter) syncNewestVersion(ctx context.Context) error {
	logger.Info("Searching for newest Excel version")

	key := config.Get().Excel.GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource(ctx)
	if err != nil {
		logger.Error("Web scraper failed to find latest source", "error", err)
		return fmt.Errorf("error searching for Excel versions: %w", err)
	}

	logger.Debug("Found potential newest source", "date", newestSource.GetMetadata().Date)

	latestVersion, err := s.sheetVersionStorer.GetNewest(ctx)
	if err != nil && err != sheetversion.ErrNoSheetVersion {
		logger.Error("Failed to get newest version from database", "error", err)
		return fmt.Errorf("error retrieving latest version: %w", err)
	}

	if latestVersion == nil {
		logger.Info("No previous version found in database - proceeding with import")
		return s.PersistSource(ctx, newestSource)
	}

	sourceDate := newestSource.GetMetadata().Date
	if sourceDate.After(latestVersion.ParsedAt) {
		logger.Info("Newer version found - starting import",
			"source_date", sourceDate,
			"db_parsed_at", latestVersion.ParsedAt,
		)
		return s.PersistSource(ctx, newestSource)
	}

	logger.Info("Current excel source version is up to date",
		"source_date", sourceDate,
		"db_parsed_at", latestVersion.ParsedAt,
	)

	return nil
}

func (s *ExcelImporter) persistSheetSubjects(
	sheet *parser.ParsedSheet,
	periodID period.PeriodID,
	writer ImportWriter,
) error {
	logger.Debug("Starting persistence of sheet subjects", "career", sheet.Name, "subjects", len(sheet.Subjects))

	planLoader, err := metadata.NewAcademicPlanLoader(sheet.Name)
	if err != nil {
		logger.Warn("Could not load academic plan metadata - proceeding without it", "career", sheet.Name, "error", err)
	} else {
		logger.Debug("Academic plan metadata loaded successfully", "career", sheet.Name)
	}

	for i, sub := range sheet.Subjects {
		offering := buildOfferingFromDTO(sheet.Name, periodID, sub, planLoader)

		if err := writer.SaveCourseOffering(offering); err != nil {
			logger.Error("Failed to save course offering",
				"career", sheet.Name,
				"subject", sub.RawSubjectName,
				"index", i,
				"error", err,
			)
			return fmt.Errorf("cannot save course offering for subject %q: %w", sub.RawSubjectName, err)
		}

		if (i+1)%50 == 0 {
			logger.Debug("Processed subjects batch", "career", sheet.Name, "count", i+1)
		}
	}

	logger.Info("All subjects persisted successfully", "career", sheet.Name, "total", len(sheet.Subjects))
	return nil
}
