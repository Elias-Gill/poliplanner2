package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteScheduleStore struct {
	db *sql.DB
}

func NewSqliteScheduleStore(db *sql.DB) *SqliteScheduleStore {
	return &SqliteScheduleStore{db: db}
}

func (s *SqliteScheduleStore) Insert(ctx context.Context, userID int64, sched *model.Schedule) (int64, error) {
	query := `
	INSERT INTO schedules (user_id, schedule_description, schedule_sheet_version)
	VALUES (?, ?, ?)
	`
	res, err := s.db.ExecContext(ctx, query,
		userID,
		sched.ScheduleDescription,
		sched.ScheduleSheetVersion,
		)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *SqliteScheduleStore) Delete(ctx context.Context, scheduleID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM schedules WHERE schedule_id = ?`, scheduleID)
	return err
}

func (s *SqliteScheduleStore) GetByUserID(ctx context.Context, userID int64) ([]*model.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT schedule_id, created_at, schedule_description, schedule_sheet_version
		FROM schedules
		WHERE user_id = ?
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Schedule
	for rows.Next() {
		sched := &model.Schedule{}
		err := rows.Scan(
			&sched.ScheduleID,
			&sched.CreatedAt,
			&sched.ScheduleDescription,
			&sched.ScheduleSheetVersion,
			)
		if err != nil {
			return nil, err
		}
		list = append(list, sched)
	}
	return list, rows.Err()
}

func (s *SqliteScheduleStore) GetByID(ctx context.Context, scheduleID int64) (*model.Schedule, error) {
	sched := &model.Schedule{}
	err := s.db.QueryRowContext(ctx, `
		SELECT schedule_id, created_at, user_id, schedule_description, schedule_sheet_version
		FROM schedules WHERE schedule_id = ?`, scheduleID).
		Scan(&sched.ScheduleID, &sched.CreatedAt, &sched.UserID, &sched.ScheduleDescription, &sched.ScheduleSheetVersion)
	if err != nil {
		return nil, err
	}
	return sched, nil
}
