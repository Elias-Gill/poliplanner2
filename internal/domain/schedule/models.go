package schedule

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type Schedule struct {
	ID          int64
	Owner       user.UserID
	Description string
	PeriodID    int64
	CreatedAt   time.Time
}

type ScheduleDetails struct {
	Schedule Schedule
	Courses  []courseOffering.CourseOfferingID
}
