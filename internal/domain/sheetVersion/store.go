package sheetversion

import (
	"context"
	"errors"
	"time"
)

var ErrNoSheetVersion = errors.New("no sheet version found")

type SheetVersionStorer interface {
	GetNewest(ctx context.Context) (*SheetVersion, error)
	GetLastCheckedAt(ctx context.Context) (*time.Time, error)

	// BUG: date depends if the server is correctly set
	SetLastCheckedAt(ctx context.Context, t time.Time) error

	// FUTURE: refactor to use a struct
	Save(
		ctx context.Context,
		fileName string,
		URI string,
		processedSheets int,
		succeededSheets int,
		errors []error,
	) (int64, error)
}
