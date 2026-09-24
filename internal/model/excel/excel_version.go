package excel

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type SheetVersionID int64

// SourceType identifies which kind of Excel source a version belongs to. Version
// matching and sync state are scoped by it, so schedules and laboratories never
// block each other.
type SourceType string

const (
	SourceTypeSchedule SourceType = "schedule"
	SourceTypeLab      SourceType = "lab"
)

// SheetVersion is a successfully parsed source. It is append-only: every
// successful parse adds a row, so the table doubles as version history.
type SheetVersion struct {
	ID       SheetVersionID
	PeriodID academic.PeriodID

	SourceType SourceType

	Name string
	URL  string

	// SourceDate is the date the source itself carries (filename date or Drive
	// modified time). It is the reference for deciding whether a source is newer.
	SourceDate time.Time

	// ParsedAt is when the server parsed the source.
	ParsedAt time.Time

	ParsedSheets int
}
