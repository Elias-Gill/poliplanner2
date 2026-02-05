package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	parser "github.com/elias-gill/poliplanner2/internal/excelparser"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/scraper"
)

type ExcelService struct {
	sheetVersionStorer store.SheetVersionStorer
	GradeStorer        store.GradeStorer
}

func NewExcelService(
	sheetVersionStorer store.SheetVersionStorer,
	GradeStorer store.GradeStorer,
) *ExcelService {
	return &ExcelService{
		sheetVersionStorer: sheetVersionStorer,
		GradeStorer:        GradeStorer,
	}
}

func (s *ExcelService) SearchOnStartup(ctx context.Context) {
	checkDate, err := s.sheetVersionStorer.GetLastCheckedAt(ctx)
	if checkDate != nil && time.Since(*checkDate) < 48*time.Hour {
		logger.Info("Excel auto-sync skipped: last check was performed less than 48 hours ago")
		return
	}

	logger.Info("Automatic excel sync started")

	err = s.SearchNewestExcel(ctx)
	if err != nil {
		logger.Error("Error on automatic version sync", "error", err)
	}

	s.sheetVersionStorer.SetLastCheckedAt(ctx, time.Now())

	logger.Info("Successfull auto excel sync")
}

func (s *ExcelService) SearchNewestExcel(ctx context.Context) error {
	key := config.Get().Excel.GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource(ctx)
	if err != nil {
		logger.Info("Scraper failed to retrieve latest source", "error", err)
		return fmt.Errorf("error searching for excel versions: %w", err)
	}

	latestVersion, err := s.sheetVersionStorer.GetNewest(ctx)
	if err != nil && !errors.Is(err, store.ErrNoSheetVersion) {
		logger.Info("Cannot retrieve latest version from database", "error", err)
		return fmt.Errorf("error searching for excel versions: %w", err)
	}

	// Si no hay versiones o la nueva es más reciente, descargar y procesar
	if latestVersion == nil || newestSource.UploadDate.After(latestVersion.ParsedAt) {
		path, err := newestSource.DownloadThisSource(ctx)
		if err != nil {
			logger.Info("Failed to download source", "source", newestSource.URL, "error", err)
			return fmt.Errorf("error downloading latest excel: %w", err)
		}

		return s.ParseAndPersistExcelFile(ctx, path, newestSource)
	}

	logger.Info("Excel is already on the latest version",
		"fpuna", newestSource.UploadDate,
		"database", latestVersion.ParsedAt)
	return nil
}

func (s *ExcelService) ParseAndPersistExcelFile(
	ctx context.Context,
	filePath string,
	source *scraper.ExcelDownloadSource,
) error {
	parserExcel, err := parser.NewExcelParser(config.Get().Paths.ExcelParsingLayoutsDir, filePath)
	if err != nil {
		return fmt.Errorf("error creating excel parser: %w", err)
	}
	defer parserExcel.Close()

	// Start the sheets parsing and persisting process
	processedSheets := 0
	succeded := 0
	var errors []error
	for parserExcel.NextSheet() {
		processedSheets++

		// Parse sheet data, if error register the error and ignore this sheet
		result, pError := parserExcel.ParseCurrentSheet()
		if pError != nil {
			logger.Error("Error parsing sheet", "error", pError)
			errors = append(errors, pError)
			continue
		}

		// Load subjects metadata for the known careers
		metadata, err := parser.NewSubjectMetadataLoader(config.Get().Paths.SubjectsMetadataDir, result.Career)
		if err != nil {
			logger.Warn("Metadata loading error", "error", err)
		}

		// insert every subject, one at a time
		insertedCount := 0
		upsertError := s.GradeStorer.Upsert(
			ctx,
			func(persist func(model.GradeModel) error) error {
				// Agreggate and persist structs data
				for _, sub := range result.Subjects {
					// Resolve all empty metadata first
					if sub.Semester == 0 && metadata != nil {
						meta, merr := metadata.Find(sub.RawSubjectName)
						if merr == nil {
							sub.Semester = meta.Semester
						}
						// FUTURE: expand with more metadata
					}

					// Create the final aggregated data model
					var grade model.GradeModel

					grade.Name = sub.RawSubjectName

					grade.Section = sub.Section

					grade.Period = model.Period{
						Year:   source.UploadDate.Year(),
						Period: source.Period,
					}

					grade.Teachers = make([]model.Teacher, len(sub.Teachers))
					for i, t := range sub.Teachers {
						teacher := model.Teacher{
							Name:  strings.TrimSpace(t.FirstName + " " + t.LastName),
							Email: t.Email,
						}
						if err := teacher.GenerateSearchKey(t.FirstName, t.LastName); err != nil {
							logger.Debug("Cannot generate searching key for teacher", "name", teacher.Name)
							continue
						}
						grade.Teachers[i] = teacher
					}

					grade.Curriculum = model.Curriculum{
						Semester: sub.Semester,
						Career:   result.Career,
						Subject: model.Subject{
							Name:       sub.TentativeRealSubjectName,
							Department: sub.Department,
						},
					}

					// partials info
					grade.Partial1Date = sub.Partial1Date
					grade.Partial1Time = sub.Partial1Time
					grade.Partial1Room = sub.Partial1Room

					grade.Partial2Date = sub.Partial2Date
					grade.Partial2Time = sub.Partial2Time
					grade.Partial2Room = sub.Partial2Room

					// finals info
					grade.Final1Date = sub.Final1Date
					grade.Final1Time = sub.Final1Time
					grade.Final1Room = sub.Final1Room
					grade.Final1RevDate = sub.Final1RevDate
					grade.Final1RevTime = sub.Final1RevTime

					grade.Final2Date = sub.Final2Date
					grade.Final2Time = sub.Final2Time
					grade.Final2Room = sub.Final2Room
					grade.Final2RevDate = sub.Final2RevDate
					grade.Final2RevTime = sub.Final2RevTime

					// revision committee info
					grade.CommitteeMember1 = sub.CommitteeMember1
					grade.CommitteeMember2 = sub.CommitteeMember2
					grade.CommitteePresident = sub.CommitteePresident

					// weekly schedules
					grade.Monday = model.TimeSlot{Start: sub.Monday.Start, End: sub.Monday.End}
					grade.Tuesday = model.TimeSlot{Start: sub.Tuesday.Start, End: sub.Tuesday.End}
					grade.Wednesday = model.TimeSlot{Start: sub.Wednesday.Start, End: sub.Wednesday.End}
					grade.Thursday = model.TimeSlot{Start: sub.Thursday.Start, End: sub.Thursday.End}
					grade.Friday = model.TimeSlot{Start: sub.Friday.Start, End: sub.Friday.End}
					grade.Saturday = model.TimeSlot{Start: sub.Saturday.Start, End: sub.Saturday.End}

					// classrooms
					grade.MondayRoom = sub.MondayRoom
					grade.TuesdayRoom = sub.TuesdayRoom
					grade.WednesdayRoom = sub.WednesdayRoom
					grade.ThursdayRoom = sub.ThursdayRoom
					grade.FridayRoom = sub.FridayRoom
					grade.SaturdayRoom = sub.SaturdayRoom
					grade.SaturdayDates = sub.SaturdayDates

					if err := persist(grade); err != nil {
						logger.Error("Error persisting subject", "error", err)
						errors = append(errors, err)
						return err
					}

					insertedCount++
				}

				// no errors for this sheet
				return nil
			},
		)

		// Log parsing result summary
		if upsertError != nil {
			logger.Error("Cannot persist sheet", "career", result.Career, "error", upsertError)
			continue
		}

		cacheHits := 0
		if metadata != nil {
			cacheHits = metadata.CacheHits
		}
		logger.Info(
			"Persisted subjects from career",
			"career", result.Career,
			"inserted_subjects", insertedCount,
			"cache_hits", cacheHits,
		)

		succeded++
	}

	// Saves a brief audit summary of the Excel parsing process.
	// Includes the number of processed sheets and any errors encountered.
	_, err = s.sheetVersionStorer.Save(
		ctx,
		source.Name,
		filePath,
		source.URL,
		processedSheets,
		succeded,
		errors,
	)
	if err != nil {
		logger.Error("Error persisting excel audit summary data", "error", err)
		return err
	}

	logger.Info("Excel import completed successfully", "file", source.Name, "url", source.URL)
	return nil
}
