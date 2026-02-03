package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type ScheduleService struct {
	scheduleStorer store.ScheduleStorer
}

func NewScheduleService(
	scheduleStorer store.ScheduleStorer,
) *ScheduleService {
	return &ScheduleService{
		scheduleStorer: scheduleStorer,
	}
}

func (s *ScheduleService) ListByUser(
	ctx context.Context,
	userID int64,
) ([]*model.Schedule, error) {

	schedules, err := s.scheduleStorer.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return schedules, err
}

func (s *ScheduleService) FindDetails(
	ctx context.Context,
	userID int64,
	scheduleID int64,
) (*model.ScheduleDetails, error) {

	details, err := s.scheduleStorer.GetByID(ctx, scheduleID)
	if details.Schedule.OwnerID != userID {
		return nil, fmt.Errorf("Operation not authorized")
	}

	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return details, nil
}

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	userID int64,
	name string,
	description string,
	gradeIDs []int64,
) (int64, error) {

	scheId, err := s.scheduleStorer.Insert(
		ctx,
		&model.ScheduleBasicData{
			Owner:       userID,
			Name:        name,
			Description: description,
			GradeIDs:    gradeIDs,
		})
	if err != nil {
		return -1, fmt.Errorf("error creating schedule: %w", err)
	}

	return scheId, nil
}

func (s *ScheduleService) DeleteSchedule(
	ctx context.Context,
	userID int64,
	scheduleID int64,
) error {
	sche, err := s.scheduleStorer.GetByID(ctx, scheduleID)
	if err != nil {
		// FIX: continue
		return err
	}

	if sche.Schedule.OwnerID != userID {
		// TODO: return new error
		// FIX: continue
		return nil
	}

	return s.scheduleStorer.Delete(ctx, scheduleID)
}
