package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

func FindUserSchedules(ctx context.Context, userID int64) ([]*model.Schedule, error) {
	schedules, err := scheduleStorer.GetByUserID(ctx, db, userID)

	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return schedules, err
}

func FindScheduleDetail(ctx context.Context, scheduleID int64) ([]*model.Subject, error) {
	subjects, err := scheduleDetailStorer.GetSubjectsByScheduleID(ctx, db, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return subjects, nil
}

func CreateSchedule(ctx context.Context, userID int64, sheetVersionId int64, description string, subjects []int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction over schedule table: %w", err)
	}

	// Roll back if any error has been encountered
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	scheId, err := scheduleStorer.Insert(ctx, tx, &model.Schedule{
		UserID:       userID,
		SheetVersion: sheetVersionId,
		Description:  description,
	})
	if err != nil {
		return fmt.Errorf("error creating schedule: %w", err)
	}

	for _, id := range subjects {
		err := scheduleDetailStorer.Insert(ctx, tx, scheId, id)
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
