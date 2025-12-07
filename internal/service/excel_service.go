package service

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/config"
	parser "github.com/elias-gill/poliplanner2/internal/excelparser"
	mapper "github.com/elias-gill/poliplanner2/internal/excelparser/dto"
	"github.com/elias-gill/poliplanner2/internal/scraper"
)

func SearchNewestExcel(ctx context.Context) {
	key := config.Get().GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	newestSource, err := scraper.FindLatestDownloadSource()
	if err != nil {
		// TODO: mostrar un mensaje de respuesta
		return
	}

	// FIX: error handling
	latestVersion, err := FindLatestSheetVersion(ctx)
	if newestSource.UploadDate.Before(latestVersion.ParsedAt) {
		// TODO: mensaje de que ya es la version mas nueva
		return
	}

	// FIX: error handling
	file, _ := newestSource.DownloadThisSource()

	// FIX: layouts dir and error handling
	p, _ := parser.NewExcelParser(config.Get().LayoutsDir, file)

	// TODO: aca deberia de comenzar una transaccion
	tx, err := db.BeginTx(ctx, nil)
	defer func() {
		// FIX: error handling
		if err != nil {
			tx.Rollback()
		}
	}()

	metadata := parser.NewSubjectMetadataLoader(config.Get().MetadataDir)
	for p.NextSheet() {
		// FIX: error handling
		result, _ := p.ParseCurrentSheet()

		// TODO: crear la carrera con la info de result

		for _, sub := range result.Subjects {
			mapper.MapToSubject(sub)
			// TODO: buscar la metadata del subject para completar semestre
			metadata.FindSubjectByName(result.Career, sub.SubjectName)
			// TODO: crear cada subject con la info de la carrera creada
		}
	}

	// TODO: finalizar transaccion y retornar algo que diga que si hay version nueva o algo asi
	// FIX: error hanlding
	tx.Commit()
}
