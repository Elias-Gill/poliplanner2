package schedule

import "context"

type ScheduleStorer interface {
	Insert(ctx context.Context, s *ScheduleBasicData) (int64, error)
	Delete(ctx context.Context, scheduleID int64) error

	ListByUserID(ctx context.Context, userID int64) ([]*Schedule, error)
	GetByID(ctx context.Context, scheduleID int64) (*ScheduleDetails, error)
}
