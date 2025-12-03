package model

import "time"

type SheetVersion struct {
	ID       int64
	FileName string
	URL      string
	ParsedAt time.Time
}
