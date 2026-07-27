package schedule

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/schedule"
	"github.com/elias-gill/poliplanner2/internal/model/user"
)

type ScheduleRepository interface {
	Save(ctx context.Context, s schedule.Schedule) (schedule.ScheduleID, error)
	Delete(ctx context.Context, scheduleID schedule.ScheduleID) error

	ListByUserID(ctx context.Context, userID user.UserID) ([]schedule.ScheduleSummaryView, error)

	GetDetailsByID(ctx context.Context, ID schedule.ScheduleID) (*schedule.ScheduleDetails, error)
}
