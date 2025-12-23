package store

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteScheduleStore struct {
}

func NewSqliteScheduleStore() *SqliteScheduleStore {
	return &SqliteScheduleStore{}
}

func (s SqliteScheduleStore) Insert(ctx context.Context, exec Executor, sched *model.Schedule) (int64, error) {
	query := `
	INSERT INTO schedules (user_id, schedule_description, schedule_sheet_version)
	VALUES (?, ?, ?)
	`
	res, err := exec.ExecContext(ctx, query,
		sched.UserID,
		sched.Description,
		sched.SheetVersion,
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

func (s SqliteScheduleStore) Delete(ctx context.Context, exec Executor, scheduleID int64) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM schedules WHERE schedule_id = ?`, scheduleID)
	return err
}

func (s SqliteScheduleStore) GetByUserID(ctx context.Context, exec Executor, userID int64) ([]*model.Schedule, error) {
	rows, err := exec.QueryContext(ctx, `
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
			&sched.ID,
			&sched.CreatedAt,
			&sched.Description,
			&sched.SheetVersion,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, sched)
	}
	return list, rows.Err()
}

func (s SqliteScheduleStore) GetByID(ctx context.Context, exec Executor, scheduleID int64) (*model.Schedule, error) {
	sched := &model.Schedule{}
	err := exec.QueryRowContext(ctx, `
		SELECT schedule_id, created_at, user_id, schedule_description, schedule_sheet_version
		FROM schedules WHERE schedule_id = ?`, scheduleID).
		Scan(&sched.ID, &sched.CreatedAt, &sched.UserID, &sched.Description, &sched.SheetVersion)
	if err != nil {
		return nil, err
	}
	return sched, nil
}

func (s SqliteScheduleStore) UpdateScheduleSubjects(
	ctx context.Context,
	exec Executor,
	scheduleID int64,
	newSubjectIDs []int64,
) error {
	// Delete old entries
	_, err := exec.ExecContext(ctx, "DELETE FROM schedule_subjects WHERE schedule_id = ?", scheduleID)
	if err != nil {
		return err
	}

	// Prepare multiple insertion
	stmt, err := exec.PrepareContext(ctx, "INSERT INTO schedule_subjects(schedule_id, subject_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, subjectID := range newSubjectIDs {
		_, err := stmt.ExecContext(ctx, scheduleID, subjectID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s SqliteScheduleStore) UpdateScheduleExcelVersion(
	ctx context.Context,
	exec Executor,
	scheduleID int64,
	newSheetVersionID int64,
) error {
	_, err := exec.ExecContext(ctx,
		`UPDATE schedules SET schedule_sheet_version = ? WHERE schedule_id = ?`,
		newSheetVersionID,
		scheduleID,
	)
	if err != nil {
		return err
	}

	return nil
}
