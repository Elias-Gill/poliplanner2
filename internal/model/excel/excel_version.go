package excel

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type SheetVersionID int64

type SheetVersion struct {
	ID       SheetVersionID
	PeriodID academic.PeriodID

	Name         string
	URL          string
	ParsedAt     time.Time
	ParsedSheets int

	Succeeded bool
	Error     string
}
