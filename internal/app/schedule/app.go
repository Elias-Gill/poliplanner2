package schedule

import (
	"context"
	"errors"
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

var (
	ErrPermissionDenied = errors.New("User has no permission")
)

type ScheduleService struct {
	storer schedule.ScheduleStorer
}

func NewScheduleService(storer schedule.ScheduleStorer) *ScheduleService {
	return &ScheduleService{
		storer: storer,
	}
}

func (s ScheduleService) ListUserSchedules(ctx context.Context, userID user.UserID) ([]schedule.ScheduleSummary, error) {
	return s.storer.ListByUserID(ctx, userID)
}

func (s ScheduleService) GetSchedule(ctx context.Context, userID user.UserID, scheduleID schedule.ScheduleID) (*schedule.Schedule, error) {
	// TODO: implementar
	return &schedule.Schedule{
		ID:          1,
		Owner:       1,
		Description: "Nada nuevo",
		CreatedAt:   time.Now(),
		Courses:     []courseOffering.CourseOfferingID{1, 2},
	}, nil
	// sche, err := s.storer.GetDetailsByID(ctx, scheduleID)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// if sche.Owner != userID {
	// 	return nil, ErrPermissionDenied
	// }
	//
	// return sche, nil
}

func (s ScheduleService) Save(ctx context.Context, sche schedule.Schedule) (schedule.ScheduleID, error) {
	return s.storer.Save(ctx, sche)
}
