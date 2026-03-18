package schedule

import (
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type ScheduleBasicData struct {
	Owner       user.UserID
	Description string
	CoursesIDs  []courseOffering.CourseOfferingID
}

type ScheduleDetails struct {
	Owner       user.UserID
	Description string
	Courses     []courseOffering.CourseOffering
}
