package schedule

import (
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type ScheduleBasicData struct {
	Owner       user.UserID
	Name        string
	Description string
	CourseIDs   []courseOffering.CourseOfferingID
}
