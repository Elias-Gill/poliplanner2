package schedule

import (
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/model/user"
)

type ScheduleID int64

type Schedule struct {
	ID        ScheduleID
	Owner     user.UserID
	Title     string
	CreatedAt time.Time
	Courses   []academic.CourseID
}

func NewSchedule(owner user.UserID, description string, courses []academic.CourseID) (*Schedule, error) {
	if len(courses) == 0 {
		return nil, fmt.Errorf("cannot have and empty schedule")
	}

	return &Schedule{
		Owner:     owner,
		Title:     description,
		CreatedAt: time.Now().In(timezone.ParaguayTZ),
		Courses:   courses,
	}, nil
}
