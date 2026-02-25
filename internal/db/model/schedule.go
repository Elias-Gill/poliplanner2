package model

import (
	"time"
)

type Schedule struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
	PeriodID    int64
	CreatedAt   time.Time
}

type ScheduleDetails struct {
	Schedule Schedule
	Courses  []CourseAggregate
}

type ScheduleBasicData struct {
	Owner       int64
	Name        string
	Description string
	CourseIDs    []int64
}
