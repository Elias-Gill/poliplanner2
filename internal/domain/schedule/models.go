package schedule

import (
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type ScheduleID int64

type Schedule struct {
	ID          ScheduleID
	Owner       user.UserID
	Description string
	CreatedAt   time.Time
	Courses     []courseOffering.CourseOfferingID
}

func NewSchedule(owner user.UserID, description string, courses []courseOffering.CourseOfferingID) (*Schedule, error) {
	if len(courses) == 0 {
		return nil, fmt.Errorf("cannot have and empty schedule")
	}

	return &Schedule{
		Owner:       owner,
		Description: description,
		CreatedAt:   time.Now().In(timezone.ParaguayTZ),
		Courses:     courses,
	}, nil
}
