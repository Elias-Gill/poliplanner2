package excel

import (
	excelimport "github.com/elias-gill/poliplanner2/internal/app/excelImport"
	"github.com/go-chi/chi/v5"
)

func NewExcelRouter(excelService *excelimport.ExcelImporter) func(r chi.Router) {
	handlers := newExcelHandlers(excelService)

	return func(r chi.Router) {
		r.Get("/", handlers.SyncForm)
		r.Post("/sync", handlers.Sync)
	}
}
