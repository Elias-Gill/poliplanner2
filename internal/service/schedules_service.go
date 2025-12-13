package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
)

type ScheduleService struct {
	db                   *sql.DB
	scheduleStorer       store.ScheduleStorer
	scheduleDetailStorer store.ScheduleDetailStorer
}

func NewScheduleService(
	db *sql.DB,
	scheduleStorer store.ScheduleStorer,
	scheduleDetailStorer store.ScheduleDetailStorer,
) *ScheduleService {
	return &ScheduleService{
		db:                   db,
		scheduleStorer:       scheduleStorer,
		scheduleDetailStorer: scheduleDetailStorer,
	}
}

func (s *ScheduleService) FindUserSchedules(
	ctx context.Context,
	userID int64,
) ([]*model.Schedule, error) {

	schedules, err := s.scheduleStorer.GetByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return schedules, err
}

func (s *ScheduleService) FindScheduleDetail(
	ctx context.Context,
	scheduleID int64,
) ([]*model.Subject, error) {

	subjects, err := s.scheduleDetailStorer.GetSubjectsByScheduleID(ctx, s.db, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return subjects, nil
}

func (s *ScheduleService) CreateSchedule(
	ctx context.Context,
	userID int64,
	sheetVersionId int64,
	description string,
	subjects []int64,
) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction over schedule table: %w", err)
	}

	// Roll back if any error has been encountered
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	scheId, err := s.scheduleStorer.Insert(ctx, tx, &model.Schedule{
		UserID:       userID,
		SheetVersion: sheetVersionId,
		Description:  description,
	})
	if err != nil {
		return fmt.Errorf("error creating schedule: %w", err)
	}

	for _, id := range subjects {
		err := s.scheduleDetailStorer.Insert(ctx, tx, scheId, id)
		if err != nil {
			return fmt.Errorf("error inserting schedule detail: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot commit schedule creation: %w", err)
	}

	return nil
}

func (s *ScheduleService) DeleteSchedule(
	ctx context.Context,
	userID int64,
	scheduleID int64,
) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction over schedule table: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// User ownership validation
	schedule, err := s.scheduleStorer.GetByID(ctx, tx, scheduleID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}
	if schedule.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	err = s.scheduleStorer.Delete(ctx, tx, scheduleID)
	if err != nil {
		return fmt.Errorf("error deleting schedule details: %w", err)
	}

	err = s.scheduleStorer.Delete(ctx, tx, scheduleID)
	if err != nil {
		return fmt.Errorf("error deleting schedule: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot commit schedule deletion: %w", err)
	}

	return nil
}
