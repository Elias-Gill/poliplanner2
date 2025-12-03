package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

func FindUserSchedules(ctx context.Context, userID int64) ([]*model.Schedule, error) {
	schedules, err := scheduleStorer.GetByUserID(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return schedules, err
}

func FindScheduleDetail(ctx context.Context, scheduleID int64) ([]*model.Subject, error) {
	subjects, err := scheduleDetailStorer.GetSubjectsByScheduleID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("error searching schedules: %w", err)
	}

	return subjects, nil
}

// FIX: make transactions and more secure inserts
func CreateSchedule(ctx context.Context, userID int64, sheetVersionId int64, description string, subjects []int64) error {
	scheId, err := scheduleStorer.Insert(ctx, &model.Schedule{
		UserID:               userID,
		SheetVersion: sheetVersionId,
		Description:  description,
	})
	if err != nil {
		return fmt.Errorf("error creating schedule: %w", err)
	}

	for _, id := range subjects {
		err := scheduleDetailStorer.Insert(ctx, scheId, id)
		if err != nil {
			return fmt.Errorf("error inserting schedule detail: %w", err)
		}
	}

	return nil
}
