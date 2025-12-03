package model

import "database/sql"

type Career struct {
	ID             int64
	CareerCode     string
	SheetVersionID sql.NullInt64
}
