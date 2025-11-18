package models

import "time"

type SheetVersion struct {
	VersionID int64
	FileName  string
	URL       string
	ParsedAt  time.Time
}
