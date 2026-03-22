package sheetVersion

import "time"

type SheetVersionID int64

type SheetVersion struct {
	ID SheetVersionID

	FileName string
	URL      string
	ParsedAt time.Time

	sheets int
	errors []string

	// hacer el calculo manual de la cantidad de errores al crear el agregado
	// En la db simplemente se guardan los errores y ya

	// TODO: enlazar con el periodo para poder hacer filtrado
	// Period period.PeriodID
}
