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
	sheetVersionStorer   store.SheetVersionStorer
	subjecStorer         store.SubjectStorer
}

func NewScheduleService(
	db *sql.DB,
	scheduleStorer store.ScheduleStorer,
	scheduleDetailStorer store.ScheduleDetailStorer,
	sheetVersionStorer store.SheetVersionStorer,
	subjecStorer store.SubjectStorer,
) *ScheduleService {
	return &ScheduleService{
		db:                   db,
		scheduleStorer:       scheduleStorer,
		scheduleDetailStorer: scheduleDetailStorer,
		sheetVersionStorer:   sheetVersionStorer,
		subjecStorer:         subjecStorer,
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

func (s *ScheduleService) MigrateSchedule(ctx context.Context, userID int64, scheduleID int64) error {
	schedule, err := s.scheduleStorer.GetByID(ctx, s.db, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// Check if the schedule belongs to the user
	if schedule.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	latestExcel, err := s.sheetVersionStorer.GetNewest(ctx, s.db)
	if err != nil {
		return fmt.Errorf("failed to get newest sheet version: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Make sure to rollback if anything fails
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get current subjects linked to the schedule
	details, err := s.scheduleDetailStorer.GetSubjectsByScheduleID(ctx, tx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule subjects: %w", err)
	}

	// Collect updated subject IDs for the new sheet version
	var updatedSubjectIDs []int64
	for _, subject := range details {
		updatedVersion, err := s.subjecStorer.GetBySheetVersion(ctx, tx, subject, latestExcel.ID)
		if err != nil {
			return fmt.Errorf("failed to find updated subject version for %s: %w", subject.SubjectName, err)
		}
		updatedSubjectIDs = append(updatedSubjectIDs, updatedVersion)
	}

	// Update schedule subjects
	err = s.scheduleStorer.UpdateScheduleSubjects(ctx, tx, scheduleID, updatedSubjectIDs)
	if err != nil {
		return fmt.Errorf("failed to update schedule subjects: %w", err)
	}

	// Update schedule sheet version
	err = s.scheduleStorer.UpdateScheduleExcelVersion(ctx, tx, scheduleID, latestExcel.ID)
	if err != nil {
		return fmt.Errorf("failed to update schedule sheet version: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
