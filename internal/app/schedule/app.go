package schedule

import (
	"context"
	"errors"

	"github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"github.com/elias-gill/poliplanner2/logger"
)

var (
	ErrPermissionDenied = errors.New("User has no permission")
)

type Schedule struct {
	storer schedule.ScheduleRepository
}

func New(storer schedule.ScheduleRepository) *Schedule {
	return &Schedule{
		storer: storer,
	}
}

// ListUserSchedules returns all schedules for a user
func (s Schedule) ListUserSchedules(ctx context.Context, userID user.UserID) ([]schedule.ScheduleSummary, error) {
	logger.Debug("ListUserSchedules called", "userID", userID)
	sched, err := s.storer.ListByUserID(ctx, userID)
	if err != nil {
		logger.Debug("cannot list user schedules", "userID", userID, "error", err)
		return nil, err
	}
	logger.Debug("ListUserSchedules successful", "userID", userID, "count", len(sched))
	return sched, nil
}

// Get returns schedule details if the user owns it
func (s Schedule) Get(ctx context.Context, userID user.UserID, scheduleID schedule.ScheduleID) (*schedule.Schedule, error) {
	logger.Debug("GetSchedule called", "userID", userID, "scheduleID", scheduleID)
	sche, err := s.storer.GetDetailsByID(ctx, scheduleID)
	if err != nil {
		logger.Debug("cannot get schedule details", "scheduleID", scheduleID, "error", err)
		return nil, err
	}

	if sche.Owner != userID {
		logger.Debug("permission denied for schedule", "scheduleID", scheduleID, "userID", userID)
		return nil, ErrPermissionDenied
	}

	logger.Debug("GetSchedule successful", "scheduleID", scheduleID, "userID", userID)
	return sche, nil
}

// Save persists a schedule and returns its ID
func (s Schedule) Save(ctx context.Context, sche schedule.Schedule) (schedule.ScheduleID, error) {
	logger.Debug("Save called", "title", sche.Title, "owner", sche.Owner)

	id, err := s.storer.Save(ctx, sche)
	if err != nil {
		logger.Debug("Save failed", "title", sche.Title, "owner", sche.Owner, "error", err)
		return 0, err
	}

	logger.Debug("Save successful", "scheduleID", id, "title", sche.Title)
	return id, nil
}

func (s Schedule) Delete(ctx context.Context, userID user.UserID, scheduleID schedule.ScheduleID) error {
	logger.Debug("Delete schedule called", "userID", userID, "scheduleID", scheduleID)
	sche, err := s.storer.GetDetailsByID(ctx, scheduleID)
	if err != nil {
		logger.Debug("cannot get schedule details for deletion", "scheduleID", scheduleID, "error", err)
		return err
	}

	if sche.Owner != userID {
		logger.Debug("permission denied for schedule", "scheduleID", scheduleID, "userID", userID)
		return ErrPermissionDenied
	}

	err = s.storer.Delete(ctx, scheduleID)
	if err != nil {
		return err
	}

	logger.Debug("Schedule deletion successful", "scheduleID", scheduleID, "userID", userID)
	return nil
}

// TitleIsAvailable checks if the user has a schedule with the same title
func (s Schedule) TitleIsAvailable(ctx context.Context, userID user.UserID, title string) (bool, error) {
	logger.Debug("TitleIsAvailable called", "userID", userID, "title", title)
	list, err := s.ListUserSchedules(ctx, userID)
	if err != nil {
		logger.Error("cannot check title existence", "userID", userID, "title", title, "error", err)
		return false, err
	}

	for _, entry := range list {
		if entry.Title == title {
			logger.Debug("title already exists", "userID", userID, "title", title)
			return false, nil
		}
	}

	logger.Debug("title available", "userID", userID, "title", title)
	return true, nil
}
