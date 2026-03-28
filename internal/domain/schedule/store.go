package schedule

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type ScheduleStorer interface {
	Save(ctx context.Context, s Schedule) (ScheduleID, error)
	Delete(ctx context.Context, scheduleID ScheduleID) error

	ListByUserID(ctx context.Context, userID user.UserID) ([]ScheduleSummary, error)
	GetDetailsByID(ctx context.Context, ID ScheduleID) (*Schedule, error)
}
