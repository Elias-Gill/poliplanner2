package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/db/model"
	parser "github.com/elias-gill/poliplanner2/internal/excelparser"
	mapper "github.com/elias-gill/poliplanner2/internal/excelparser/dto"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/scraper"
)

func SearchNewestExcel(ctx context.Context) error {
	key := config.Get().GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource()
	if err != nil {
		logger.Info("scraper failed to retrieve latest source", "error", err)
		return fmt.Errorf("error searching for excel versions: %w", err)
	}

	latestVersion, err := FindLatestSheetVersion(ctx)
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

	file, err := newestSource.DownloadThisSource()
	if err != nil {
		logger.Info("failed to download source", "source", newestSource.URL, "error", err)
		return fmt.Errorf("error downloading latest excel: %w", err)
	}

	parserExcel, err := parser.NewExcelParser(config.Get().LayoutsDir, file)
	if err != nil {
		return fmt.Errorf("error creating excel parser: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
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
		FileName: newestSource.FileName,
		URL:      newestSource.URL,
	}

	if err := sheetVersionStorer.Insert(ctx, tx, version); err != nil {
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

		if err := careerStorer.Insert(ctx, tx, career); err != nil {
			logger.Error("error persisting career", "error", err)
			return rollback(err)
		}

		for _, sub := range result.Subjects {
			subject := mapper.MapToSubject(sub)

			if subject.Semester == 0 {
				meta, merr := metadata.FindSubjectByName(result.Career, sub.SubjectName)
				if merr == nil {
					subject.Semester = meta.Semester
				}
			}

			if err := subjectStorer.Insert(ctx, tx, career.ID, subject); err != nil {
				logger.Error("error persisting subject", "error", err)
				return rollback(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("error committing transaction", "error", err)
		return rollback(err)
	}

	return nil
}
