package model

import (
	"time"
)

type Schedule struct {
	ID          int64
	OwnerID     int64
	Name        int64
	Description string
	PeriodID    int64
	CreatedAt   time.Time
}

type ScheduleDetails struct {
	Schedule Schedule
	Courses  []GradeModel
}

type ScheduleBasicData struct {
	Owner       int64
	Name        string
	Description string
	GradeIDs    []int64
}
