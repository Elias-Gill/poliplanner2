package schedule

import (
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type ScheduleID int64

type Schedule struct {
	ID          ScheduleID
	Owner       user.UserID
	Description string
	PeriodID    period.PeriodID
	CreatedAt   *time.Time
	Courses     []courseOffering.CourseOfferingID
}
