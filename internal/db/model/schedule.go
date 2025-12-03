package model

import (
	"time"
)

type Schedule struct {
	ID           int64
	CreatedAt    time.Time
	UserID       int64
	Description  string
	SheetVersion int64
}
