package schedule

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type ScheduleStorer interface {
	Save(ctx context.Context, s Schedule) (ScheduleID, error)
	Delete(ctx context.Context, scheduleID ScheduleID) error

	ListByUser(ctx context.Context, userID user.UserID) ([]Schedule, error)
}
