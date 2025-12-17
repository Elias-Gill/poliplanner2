package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	parser "github.com/elias-gill/poliplanner2/internal/excelparser"
	mapper "github.com/elias-gill/poliplanner2/internal/excelparser/dto"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/scraper"
)

type ExcelService struct {
	db                 *sql.DB
	sheetVersionStorer store.SheetVersionStorer
	careerStorer       store.CareerStorer
	subjectStorer      store.SubjectStorer
}

func NewExcelService(
	db *sql.DB,
	sheetVersionStorer store.SheetVersionStorer,
	careerStorer store.CareerStorer,
	subjectStorer store.SubjectStorer,
) *ExcelService {
	return &ExcelService{
		db:                 db,
		sheetVersionStorer: sheetVersionStorer,
		careerStorer:       careerStorer,
		subjectStorer:      subjectStorer,
	}
}

func (s *ExcelService) SearchOnStartup(ctx context.Context) {
	if s.sheetVersionStorer.HasToUpdate(ctx, s.db) {
		logger.Info("Automatic search for new excel versions started")
		err := s.SearchNewestExcel(ctx)
		if err != nil {
			logger.Error("Error on automatic version sync", "error", err)
		}
	}
}

func (s *ExcelService) SearchNewestExcel(ctx context.Context) error {
	key := config.Get().Excel.GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource(ctx)
	if err != nil {
		logger.Info("Scraper failed to retrieve latest source", "error", err)
		return fmt.Errorf("error searching for excel versions: %w", err)
	}

	latestVersion, err := s.sheetVersionStorer.GetNewest(ctx, s.db)
	if err != nil {
		logger.Info("Cannot retrieve latest version from database", "error", err)
		return fmt.Errorf("error searching for excel versions: %w", err)
	}

	if newestSource.UploadDate.Before(latestVersion.ParsedAt) {
		logger.Info("Excel is already on the latest version",
			"fpuna", newestSource.UploadDate,
			"database", latestVersion.ParsedAt)
		return nil
	}

	path, err := newestSource.DownloadThisSource(ctx)
	if err != nil {
		logger.Info("Failed to download source", "source", newestSource.URL, "error", err)
		return fmt.Errorf("error downloading latest excel: %w", err)
	}

	return s.ParseExcelFile(ctx, path, newestSource.FileName, newestSource.URL)
}

func (s *ExcelService) ParseExcelFile(ctx context.Context, path string, name string, url string) error {
	parserExcel, err := parser.NewExcelParser(config.Get().Paths.ExcelParsingLayoutsDir, path)
	if err != nil {
		return fmt.Errorf("error creating excel parser: %w", err)
	}
	defer parserExcel.Close()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	rollback := func(e error) error {
		rbErr := tx.Rollback()
		if rbErr != nil {
			logger.Error("Rollback failed", "error", rbErr)
		}
		return e
	}

	version := &model.SheetVersion{
		FileName: name,
		URL:      url,
	}

	if err := s.sheetVersionStorer.Insert(ctx, tx, version); err != nil {
		logger.Error("Error persisting excel version", "error", err)
		return rollback(err)
	}

	for parserExcel.NextSheet() {
		result, perr := parserExcel.ParseCurrentSheet()
		if perr != nil {
			logger.Error("Error parsing sheet", "error", perr)
			return rollback(perr)
		}

		career := &model.Career{
			CareerCode:     result.Career,
			SheetVersionID: version.ID,
		}

		if err := s.careerStorer.Insert(ctx, tx, career); err != nil {
			logger.Error("Error persisting career", "error", err)
			return rollback(err)
		}

		metadata, err := parser.NewSubjectMetadataLoader(config.Get().Paths.SubjectsMetadataDir, result.Career)
		if err != nil {
			logger.Warn("Metadata loading error", "error", err)
		}

		insertedCount := 0
		for _, sub := range result.Subjects {
			subject := mapper.MapToSubject(sub)

			if subject.Semester == 0 && metadata != nil {
				meta, merr := metadata.FindSubjectByName(sub.SubjectName)
				if merr == nil {
					subject.Semester = meta.Semester
				}
			}

			if err := s.subjectStorer.Insert(ctx, tx, career.ID, subject); err != nil {
				logger.Error("Error persisting subject", "error", err)
				return rollback(err)
			}
			insertedCount++
		}

		// Log parsing summary
		cacheHits := 0
		if metadata != nil {
			cacheHits = metadata.CacheHits
		}
		logger.Info(
			"Persisted subjects from career",
			"career", career.CareerCode,
			"inserted_subjects", insertedCount,
			"cache_hits", cacheHits,
		)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Error committing transaction", "error", err)
		return rollback(err)
	}

	logger.Info("Excel import completed successfully", "file", name, "url", url)
	return nil
}
