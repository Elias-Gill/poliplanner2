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

func (s *ExcelService) SearchNewestExcel(ctx context.Context) error {
	key := config.Get().GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource()
	if err != nil {
		logger.Info("scraper failed to retrieve latest source", "error", err)
		return fmt.Errorf("error searching for excel versions: %w", err)
	}

	latestVersion, err := s.sheetVersionStorer.GetNewest(ctx, s.db)
	if err != nil {
		logger.Info("cannot retrieve latest version from database", "error", err)
		return fmt.Errorf("error searching for excel versions: %w", err)
	}

	if newestSource.UploadDate.Before(latestVersion.ParsedAt) {
		logger.Info("excel is already the latest version",
			"fpuna", newestSource.UploadDate,
			"database", latestVersion.ParsedAt)
		return nil
	}

	path, err := newestSource.DownloadThisSource()
	if err != nil {
		logger.Info("failed to download source", "source", newestSource.URL, "error", err)
		return fmt.Errorf("error downloading latest excel: %w", err)
	}

	return s.ParseExcelFile(ctx, path, newestSource.FileName, newestSource.URL)
}

func (s *ExcelService) ParseExcelFile(ctx context.Context, path string, name string, url string) error {
	parserExcel, err := parser.NewExcelParser(config.Get().LayoutsDir, path)
	if err != nil {
		return fmt.Errorf("error creating excel parser: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	rollback := func(e error) error {
		rbErr := tx.Rollback()
		if rbErr != nil {
			logger.Error("rollback failed", "error", rbErr)
		}
		return e
	}

	version := &model.SheetVersion{
		FileName: name,
		URL:      url,
	}

	if err := s.sheetVersionStorer.Insert(ctx, tx, version); err != nil {
		logger.Error("error persisting excel version", "error", err)
		return rollback(err)
	}

	metadata := parser.NewSubjectMetadataLoader(config.Get().MetadataDir)

	for parserExcel.NextSheet() {
		result, perr := parserExcel.ParseCurrentSheet()
		if perr != nil {
			logger.Error("error parsing sheet", "error", perr)
			return rollback(perr)
		}

		career := &model.Career{
			CareerCode:     result.Career,
			SheetVersionID: version.ID,
		}

		if err := s.careerStorer.Insert(ctx, tx, career); err != nil {
			logger.Error("error persisting career", "error", err)
			return rollback(err)
		}

		logger.Info("persisting subjects from career", "career", career.CareerCode, "num_subjects", len(result.Subjects), "cache_hits", metadata.CacheHits)

		insertedCount := 0
		for _, sub := range result.Subjects {
			subject := mapper.MapToSubject(sub)

			if subject.Semester == 0 {
				meta, merr := metadata.FindSubjectByName(result.Career, sub.SubjectName)
				if merr == nil {
					subject.Semester = meta.Semester
				}
			}

			if err := s.subjectStorer.Insert(ctx, tx, career.ID, subject); err != nil {
				logger.Error("error persisting subject", "error", err)
				return rollback(err)
			}
			insertedCount++
		}

		logger.Info("persisted subjects from career", "career", career.CareerCode, "inserted_subjects", insertedCount)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("error committing transaction", "error", err)
		return rollback(err)
	}

	logger.Info("excel import completed successfully", "file", name, "url", url)
	return nil
}
