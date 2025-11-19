package model

import (
	"time"
)

type Schedule struct {
	ScheduleID           int64
	CreatedAt            time.Time
	UserID               int64
	ScheduleDescription  string
	ScheduleSheetVersion int64
}
