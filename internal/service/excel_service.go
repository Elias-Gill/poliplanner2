package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	"github.com/elias-gill/poliplanner2/internal/excel"
	parser "github.com/elias-gill/poliplanner2/internal/excel/parser"
	"github.com/elias-gill/poliplanner2/internal/excel/scraper"
	"github.com/elias-gill/poliplanner2/internal/logger"
)

type ExcelService struct {
	sheetVersionStorer store.SheetVersionStorer
	CoursesStorer      store.CourseStorer
}

func NewExcelService(sheetVersionStorer store.SheetVersionStorer, CoursesStorer store.CourseStorer) *ExcelService {
	return &ExcelService{
		sheetVersionStorer: sheetVersionStorer,
		CoursesStorer:      CoursesStorer,
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

	if err := s.SearchNewestExcel(ctx); err != nil {
		logger.Error("Error on automatic version sync", "error", err)
		return
	}

	logger.Info("Successful auto Excel sync")
}

// SearchNewestExcel checks for the newest Excel version and triggers parsing/persisting if needed
func (s *ExcelService) SearchNewestExcel(ctx context.Context) error {
	key := config.Get().Excel.GoogleAPIKey
	scrapper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scrapper.FindLatestDownloadSource(ctx)
	if err != nil {
		logger.Info("Scraper failed", "error", err)
		return fmt.Errorf("error searching for Excel versions: %w", err)
	}

	latestVersion, err := s.sheetVersionStorer.GetNewest(ctx)
	// If the error is ErrNoSheetVersion, we continue processing because the
	// absence of a previous version can mean that this is the first Excel import
	// (unlikely on production tho).
	if err != nil && err != store.ErrNoSheetVersion {
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
func (s *ExcelService) ParseAndPersistNewExcel(ctx context.Context, source excel.ExcelSource) error {
	// Get content reader from the Excel source
	content, err := source.GetContent(ctx)
	if err != nil {
		return fmt.Errorf("cannot open Excel source: %w", err)
	}
	defer content.Close()

	excelMeta := source.GetMetadata()

	// Create Excel parser
	// REFACTOR: think about keeping the API passing the layout path or encapsulating more
	parserExcel, err := parser.NewExcelParser(config.Get().Paths.ExcelParsingLayoutsDir, content)
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

		// Load subject metadata for known careers
		subjectMeta, err := parser.NewAcademicPlanLoader(config.Get().Paths.SubjectsMetadataDir, sheetResult.Career)
		if err != nil {
			logger.Info("Cannot load academic plan", "career", sheetResult.Career, "error", err)
		}

		insertedCount, err := s.persistSheetSubjects(ctx, sheetResult, subjectMeta, excelMeta)
		if err != nil {
			sheetErrors = append(sheetErrors, err)
			continue
		}

		logger.Info("Processed sheet", "career", sheetResult.Career, "inserted_subjects", insertedCount)
		succeded++
	}

	// Save brief audit summary of Excel parsing
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
	sheet *parser.ParsingResult,
	planLoader *parser.AcademicPlanLoader,
	excelMeta excel.ExcelSourceMetadata,
) (int, error) {
	inserted := 0
	err := s.CoursesStorer.Upsert(ctx, func(persist func(model.CourseModel) error) error {
		for _, sub := range sheet.Subjects {
			if sub.RawSubjectName == "" || sub.TentativeRealSubjectName == "" {
				// Ignore if any of them are empty
				continue
			}

			// Fill semester from academic plan if possible
			// FUTURE: raw subject name can be replaced for tentative real subject name
			if sub.Semester == 0 && planLoader != nil {
				if m, err := planLoader.FindSubject(sub.RawSubjectName); err == nil {
					sub.Semester = m.Semester
				}
			}

			// Build our final aggregate from SubjectDTO
			course := buildCourseModel(sub, sheet, excelMeta)
			if err := persist(course); err != nil {
				return err
			}
			inserted++
		}
		return nil
	})
	return inserted, err
}

// ================================
// =           Helpers            =
// ================================

// buildCourseModel constructs the final CourseModel to persist in DB
func buildCourseModel(
	sub parser.SubjectDTO,
	sheet *parser.ParsingResult,
	excelMeta excel.ExcelSourceMetadata,
) model.CourseModel {
	course := model.CourseModel{
		Name:       sub.RawSubjectName,
		Section:    sub.Section,
		CourseType: sub.CourseType,
		Period: model.Period{
			Year:   excelMeta.Date.Year(),
			Period: excelMeta.Period,
		},
		Curriculum: model.Curriculum{
			Career:   sheet.Career,
			Semester: sub.Semester,
			Subject: model.Subject{
				Name:       sub.TentativeRealSubjectName,
				Department: sub.Department,
			},
		},
		Teachers: make([]model.Teacher, len(sub.Teachers)),

		// Weekly schedules
		Monday:    model.TimeSlot{Start: sub.Monday.Start, End: sub.Monday.End},
		Tuesday:   model.TimeSlot{Start: sub.Tuesday.Start, End: sub.Tuesday.End},
		Wednesday: model.TimeSlot{Start: sub.Wednesday.Start, End: sub.Wednesday.End},
		Thursday:  model.TimeSlot{Start: sub.Thursday.Start, End: sub.Thursday.End},
		Friday:    model.TimeSlot{Start: sub.Friday.Start, End: sub.Friday.End},
		Saturday:  model.TimeSlot{Start: sub.Saturday.Start, End: sub.Saturday.End},

		// Rooms
		MondayRoom:    sub.MondayRoom,
		TuesdayRoom:   sub.TuesdayRoom,
		WednesdayRoom: sub.WednesdayRoom,
		ThursdayRoom:  sub.ThursdayRoom,
		FridayRoom:    sub.FridayRoom,
		SaturdayRoom:  sub.SaturdayRoom,
		SaturdayDates: sub.SaturdayDates,

		// Partials
		Partial1Date: sub.Partial1Date, Partial1Time: sub.Partial1Time, Partial1Room: sub.Partial1Room,
		Partial2Date: sub.Partial2Date, Partial2Time: sub.Partial2Time, Partial2Room: sub.Partial2Room,

		// Finals
		Final1Date: sub.Final1Date, Final1Time: sub.Final1Time, Final1Room: sub.Final1Room,
		Final1RevDate: sub.Final1RevDate, Final1RevTime: sub.Final1RevTime,
		Final2Date: sub.Final2Date, Final2Time: sub.Final2Time, Final2Room: sub.Final2Room,
		Final2RevDate: sub.Final2RevDate, Final2RevTime: sub.Final2RevTime,

		// Revision committee
		CommitteeMember1:   sub.CommitteeMember1,
		CommitteeMember2:   sub.CommitteeMember2,
		CommitteePresident: sub.CommitteePresident,
	}

	for i, t := range sub.Teachers {
		teacher := model.Teacher{
			Name:  strings.TrimSpace(t.FirstName + " " + t.LastName),
			Email: t.Email,
		}
		if err := teacher.GenerateSearchKey(t.FirstName, t.LastName); err == nil {
			course.Teachers[i] = teacher
		}
	}

	return course
}
