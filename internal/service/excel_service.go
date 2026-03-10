package service

import (
	"context"
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	sheetversion "github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
	"github.com/elias-gill/poliplanner2/internal/domain/teacher"
	"github.com/elias-gill/poliplanner2/internal/parser"
	"github.com/elias-gill/poliplanner2/internal/parser/metadata"
	"github.com/elias-gill/poliplanner2/internal/scraper"
	"github.com/elias-gill/poliplanner2/internal/source"
	"github.com/elias-gill/poliplanner2/logger"
)

type ExcelService struct {
	sheetVersionStorer sheetversion.SheetVersionStorer
	coursesStorer      courseOffering.CourseStorer
	teachersStorer     teacher.TeacherStorer
	planStorer         academicPlan.AcademicPlanStorer
}

func NewExcelService(
	sheetVersionStorer sheetversion.SheetVersionStorer,
	coursesStorer courseOffering.CourseStorer,
) *ExcelService {
	return &ExcelService{
		sheetVersionStorer: sheetVersionStorer,
		coursesStorer:      coursesStorer,
	}
}

// ================================
// =         Public API           =
// ================================

// SearchOnStartup triggers automatic Excel sync on startup if last check was more than 48h ago
func (s *ExcelService) SearchOnStartup(ctx context.Context) {
	lastCheck, _ := s.sheetVersionStorer.GetLastCheckedAt(ctx)
	if lastCheck != nil && time.Since(*lastCheck) < 48*time.Hour {
		logger.Info("Excel auto-sync skipped: last check < 48h ago")
		return
	}

	logger.Info("Automatic Excel sync started")
	s.sheetVersionStorer.SetLastCheckedAt(ctx, time.Now())

	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	if err := s.SearchNewestExcel(ctx); err != nil {
		logger.Error("Error on automatic version sync", "error", err)
		return
	}

	logger.Info("Successful auto Excel sync")
}

// SearchNewestExcel checks for the newest Excel version and triggers parsing/persisting if needed
func (s *ExcelService) SearchNewestExcel(ctx context.Context) error {
	key := config.Get().Excel.GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource(ctx)
	if err != nil {
		logger.Info("Scraper failed", "error", err)
		return fmt.Errorf("error searching for Excel versions: %w", err)
	}

	latestVersion, err := s.sheetVersionStorer.GetNewest(ctx)
	// If the error is ErrNoSheetVersion, we continue processing because the
	// absence of a previous version can mean that this is the first Excel import
	// (unlikely on production tho).
	if err != nil && err != sheetversion.ErrNoSheetVersion {
		logger.Info("Cannot retrieve latest version from DB", "error", err)
		return fmt.Errorf("error searching for Excel versions: %w", err)
	}

	// If no versions or the new source is newer, parse and persist
	if latestVersion == nil || newestSource.GetMetadata().Date.After(latestVersion.ParsedAt) {
		return s.ParseAndPersistNewExcel(ctx, newestSource)
	}

	logger.Info("Excel already at latest version",
		"source_date", newestSource.GetMetadata().Date,
		"db_date", latestVersion.ParsedAt,
	)
	return nil
}

// ParseAndPersistNewExcel handles reading, parsing and persisting the Excel source
func (s *ExcelService) ParseAndPersistNewExcel(ctx context.Context, source source.ExcelSource) error {
	// Get content reader from the Excel source
	content, err := source.GetContent(ctx)
	if err != nil {
		return fmt.Errorf("cannot open Excel source: %w", err)
	}
	defer content.Close()

	// Initialize parser
	parserExcel, err := parser.NewExcelParser(content)
	if err != nil {
		return fmt.Errorf("error creating Excel parser: %w", err)
	}
	defer parserExcel.Close()

	processedSheets, succeded := 0, 0
	var sheetErrors []error

	for parserExcel.NextSheet() {
		processedSheets++

		sheetResult, pErr := parserExcel.ParseCurrentSheet()
		if pErr != nil {
			sheetErrors = append(sheetErrors, pErr)
			continue
		}

		insertedCount, err := s.persistSheetSubjects(ctx, sheetResult)
		if err != nil {
			sheetErrors = append(sheetErrors, err)
			continue
		}

		logger.Info("Processed sheet", "career", sheetResult.Name, "inserted_subjects", insertedCount)
		succeded++
	}

	// Save brief audit summary of Excel parsing
	excelMeta := source.GetMetadata()
	_, err = s.sheetVersionStorer.Save(ctx, excelMeta.Name, excelMeta.URI, processedSheets, succeded, sheetErrors)
	if err != nil {
		return fmt.Errorf("failed to save Excel audit summary: %w", err)
	}

	logger.Info("Excel import completed", "file", excelMeta.Name, "url", excelMeta.URI)
	return nil
}

// ================================
// =       Private methods        =
// ================================

// persistSheetSubjects inserts all subjects of a sheet and returns number of inserted subjects
func (s *ExcelService) persistSheetSubjects(
	ctx context.Context,
	sheet *parser.ParsedSheet,
) (int, error) {
	// Load subject metadata for the current career
	planLoader, err := metadata.NewAcademicPlanLoader(sheet.Name)
	if err != nil {
		logger.Info("Cannot load academic plan", "career", sheet.Name, "error", err)
	}

	inserted := 0

	// NOTE: for now we get the career from the sheet name. This is because it may
	// collide with "Oviedo" and "Villarrica".
	plan, _ := s.planStorer.GetOrCreateByCareer(ctx, sheet.Name)
	// FIX: error handling

	for _, sub := range sheet.Subjects {
		course, teachers := BuildAggregatesFromDTO(sub, planLoader)

		// FIX: error handling
		ids, _ := s.teachersStorer.Save(ctx, teachers)

		s.planStorer.AddSubject(ctx, plan)

		s.coursesStorer.Upsert(ctx, course)

		inserted++
	}

	return inserted, err
}

// ================================
// =           Helpers            =
// ================================
