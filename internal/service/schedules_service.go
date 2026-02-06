package service

import (
	"context"
	"database/sql"
	"errors"
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
			CourseIDs:    gradeIDs,
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
	schedule, err := s.scheduleStorer.GetByID(ctx, scheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("schedule not found")
		}
		return fmt.Errorf("get schedule: %w", err)
	}

	if schedule.Schedule.OwnerID != userID {
		return errors.New("not authorized: not the owner of this schedule")
	}

	err = s.scheduleStorer.Delete(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}

	return nil
}
