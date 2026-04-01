package sheetVersion

import (
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/period"
)

type SheetVersionID int64

type SheetVersion struct {
	ID     SheetVersionID
	Period period.PeriodID

	FileName     string
	URL          string
	ParsedAt     time.Time
	ParsedSheets int

	Succeeded bool
	Error     string
}

func NewSheetVersion(
	periodID period.PeriodID,
	fileName string,
	url string,
	parsedAt time.Time,
	parsedSheets int,
	errMsg error,
) (*SheetVersion, error) {

	if periodID == 0 {
		return nil, fmt.Errorf("period id is required")
	}

	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}

	if url == "" {
		return nil, fmt.Errorf("url is required")
	}

	if parsedAt.IsZero() {
		return nil, fmt.Errorf("parsedAt is required")
	}

	if parsedSheets < 0 {
		return nil, fmt.Errorf("parsedSheets cannot be negative")
	}

	v := &SheetVersion{
		Period:       periodID,
		FileName:     fileName,
		URL:          url,
		ParsedAt:     parsedAt,
		ParsedSheets: parsedSheets,
		Succeeded:    true,
	}

	if errMsg != nil {
		v.Error = errMsg.Error()
		v.Succeeded = false
	}

	return v, nil
}
