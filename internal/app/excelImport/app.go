package excelimport

import (
	"context"
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	sheetversion "github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
	"github.com/elias-gill/poliplanner2/internal/parser"
	"github.com/elias-gill/poliplanner2/internal/parser/metadata"
	"github.com/elias-gill/poliplanner2/internal/scraper"
	"github.com/elias-gill/poliplanner2/internal/source"
	"github.com/elias-gill/poliplanner2/logger"
)

type ImportService struct {
	sheetVersionStorer sheetversion.SheetVersionStorer
	importStorer       ImportStorer
}

func NewExcelImportService(
	importStorer ImportStorer,
	sheetVersionStorer sheetversion.SheetVersionStorer,
) *ImportService {
	return &ImportService{
		importStorer:       importStorer,
		sheetVersionStorer: sheetVersionStorer,
	}
}

// ================================
// =         Public API           =
// ================================

// PeriodicSync checks if the last sync was 48hs ago to perform a search and parsing
func (s *ImportService) AutoSync(ctx context.Context) {
	if !s.shouldAutoSync(ctx) {
		logger.Info("Excel auto-sync skipped: last check < 48h ago")
		return
	}

	logger.Info("Automatic Excel sync started")
	s.sheetVersionStorer.SetLastCheckedAt(ctx, time.Now().In(timezone.ParaguayTZ))

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := s.syncNewestVersion(ctx); err != nil {
		logger.Error("Error on automatic Excel sync", "error", err)
		return
	}

	logger.Info("Automatic Excel sync completed")
}

// Sync ignores the 48hs time treshhold but still checks if the web source is newer
// before parsing it.
func (s *ImportService) Sync(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return s.syncNewestVersion(ctx)
}

// PersistSource parses and persists an ExcelSource into persistent storage
func (s *ImportService) PersistSource(ctx context.Context, src source.ExcelSource) error {
	content, err := src.GetContent(ctx)
	if err != nil {
		return fmt.Errorf("cannot open Excel source: %w", err)
	}
	defer content.Close()

	sourceMeta := src.GetMetadata()

	parserExcel, err := parser.NewExcelParser(content)
	if err != nil {
		return fmt.Errorf("error creating Excel parser: %w", err)
	}
	defer parserExcel.Close()

	pError := s.importStorer.RunImport(ctx, func(writter ImportWriter) error {
		// Search or create a new period based on the source data
		periodID, err := writter.EnsurePeriod(period.NewPeriodFromTime(sourceMeta.Date))
		if err != nil {
			return fmt.Errorf("cannot save current period: %w", err)
		}

		for parserExcel.NextSheet() {
			// Parse the source content
			sheetResult, err := parserExcel.ParseCurrentSheet()
			if err != nil {
				logger.Error("Excel parsing error", "career", sourceMeta.Name, "error", err)
				return err
			}

			// Persist the sheet data
			err = s.persistSheetSubjects(sheetResult, periodID, writter)
			if err != nil {
				logger.Error("Sheet persistence error", "career", sourceMeta.Name, "error", err)
				return err
			}

			logger.Info("Processed sheet", "career", sheetResult.Name)
		}

		logger.Info("Excel import completed", "file", sourceMeta.Name, "url", sourceMeta.URI)
		return nil
	})

	success := pError == nil
	err = s.importStorer.SaveAudit(ctx, sourceMeta, success, pError)
	if err != nil {
		return fmt.Errorf("failed to save Excel audit summary: %w", err)
	}

	return pError
}

// ================================
// =       Private helpers        =
// ================================

// Verifies if at least 48hs had passed
func (s *ImportService) shouldAutoSync(ctx context.Context) bool {
	lastCheck, _ := s.sheetVersionStorer.GetLastCheckedAt(ctx)
	return lastCheck == nil || time.Since(*lastCheck) >= 48*time.Hour
}

// Searches the web for a new excel version source
func (s *ImportService) syncNewestVersion(ctx context.Context) error {
	key := config.Get().Excel.GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource(ctx)
	if err != nil {
		logger.Info("Scraper failed", "error", err)
		return fmt.Errorf("error searching for Excel versions: %w", err)
	}

	latestVersion, err := s.sheetVersionStorer.GetNewest(ctx)
	if err != nil && err != sheetversion.ErrNoSheetVersion {
		logger.Info("Cannot retrieve latest version from DB", "error", err)
		return fmt.Errorf("error searching for Excel versions: %w", err)
	}

	if latestVersion == nil || newestSource.GetMetadata().Date.After(latestVersion.ParsedAt) {
		return s.PersistSource(ctx, newestSource)
	}

	logger.Info("Excel already at latest version",
		"source_date", newestSource.GetMetadata().Date,
		"db_date", latestVersion.ParsedAt,
	)
	return nil
}

// persistSheetSubjects inserts all subjects of a sheet and returns the number of
// successfully inserted subjects.
//
// The course offerings, must be treated atomically. If a failure occurs
// during the sheet import, partial updates to course offerings could leave some
// careers with outdated or incomplete offerings. For this reason, course offering
// writes are executed in a transaction so that the entire excel update either
// succeeds or is rolled back.
func (s *ImportService) persistSheetSubjects(
	sheet *parser.ParsedSheet,
	periodID period.PeriodID,
	writter ImportWriter,
) error {
	// Load subject metadata for the current career
	planLoader, err := metadata.NewAcademicPlanLoader(sheet.Name)
	if err != nil {
		logger.Info("Cannot load academic plan metadata", "career", sheet.Name, "error", err)
	}

	for _, sub := range sheet.Subjects {
		offering := buildOfferingFromDTO(sheet.Name, periodID, sub, planLoader)
		if err := writter.SaveCourseOffering(offering); err != nil {
			return fmt.Errorf("cannot save course offering for subject %q: %w", sub.SubjectName, err)
		}
	}

	return err
}
