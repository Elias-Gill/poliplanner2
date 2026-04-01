package sheetVersion

import (
	"context"
	"errors"
	"time"
)

var ErrNoSheetVersion = errors.New("no sheet version found")

type SheetVersionStorer interface {
	GetNewest(ctx context.Context) (*SheetVersion, error)

	GetLastCheckedDate(ctx context.Context) (*time.Time, error)
	SetLastCheckedDate(ctx context.Context, t time.Time) error
}
