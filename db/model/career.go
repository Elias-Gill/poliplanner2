package model

import "database/sql"

type Career struct {
	CareerID       int64
	CareerCode     string
	SheetVersionID sql.NullInt64
}
