package service

import (
	"github.com/elias-gill/poliplanner2/internal/db/model"
	parser "github.com/elias-gill/poliplanner2/internal/excelparser"
	mapper "github.com/elias-gill/poliplanner2/internal/excelparser/dto"
	"github.com/elias-gill/poliplanner2/internal/scraper"
)

func newVersion() {
	drive := scraper.NewGoogleDriveHelper()
	scraper := scraper.NewWebScrapper(drive)

	source, err := scraper.FindLatestDownloadSource()
	if err != nil {
		return
	}

	var latestVersion model.SheetVersion

	if source.UploadDate.Before(latestVersion.ParsedAt) {
		return
	}

	// FIX: error handling
	file, _ := source.DownloadThisSource()

	// FIX: layouts dir and error handling
	p, _ := parser.NewExcelParser("layouts", file)

	// TODO: aca deberia de comenzar una transaccion

	for p.NextValidSheet() {
		// FIX: error handling
		result, _ := p.ParseCurrentSheet()

		// TODO: crear la carrera con la info de result

		for _, sub := range result.Subjects {
			mapper.MapToSubject(sub)
			// TODO: crear cada subject con la info de la carrera creada
		}
	}

	// TODO: finalizar transaccion y retornar algo que diga que si hay version nueva o algo asi
}
