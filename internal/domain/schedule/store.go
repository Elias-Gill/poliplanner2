package schedule

type ScheduleStorer interface {
	Insert(ctx context.Context, s *model.ScheduleBasicData) (int64, error)
	Delete(ctx context.Context, scheduleID int64) error

	ListByUserID(ctx context.Context, userID int64) ([]*model.Schedule, error)
	GetByID(ctx context.Context, scheduleID int64) (*model.ScheduleDetails, error)
}

