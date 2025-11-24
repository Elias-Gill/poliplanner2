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

func FindScheduleDetail(scheduleID int64) []model.Subject {
	return nil
}
