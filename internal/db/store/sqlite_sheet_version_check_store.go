package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/elias-gill/poliplanner2/internal/logger"
)

const last_check_entry_const_id = 1

type SqliteSheetVersionCheckStore struct {
}

func NewSqliteSheetVersionCheckStore() *SqliteSheetVersionCheckStore {
	return &SqliteSheetVersionCheckStore{}
}

func (s SqliteSheetVersionCheckStore) GetLastCheckedAt(
	ctx context.Context,
	exec Executor,
) (*time.Time, error) {

	var lastCheckedAt string

	err := exec.QueryRowContext(
		ctx,
		`SELECT last_checked_at FROM auto_sync_excel_check WHERE id = ?`,
		last_check_entry_const_id,
	).Scan(&lastCheckedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		logger.Error("Error querying auto sync table", "error", err)
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, lastCheckedAt)
	if err != nil {
		logger.Error("Error parsing last_checked_at", "value", lastCheckedAt, "error", err)
		return nil, err
	}

	return &t, nil
}

func (s SqliteSheetVersionCheckStore) SetLastCheckedAt(
	ctx context.Context,
	exec Executor,
	t time.Time,
) error {

	value := t.Format(time.RFC3339)

	result, err := exec.ExecContext(
		ctx,
		`UPDATE auto_sync_excel_check SET last_checked_at = ? WHERE id = ?`,
		value,
		last_check_entry_const_id,
	)
	if err != nil {
		logger.Error("Error updating last_checked_at", "error", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		_, err := exec.ExecContext(
			ctx,
			`INSERT INTO auto_sync_excel_check (id, last_checked_at) VALUES (?, ?)`,
			last_check_entry_const_id,
			value,
		)
		if err != nil {
			logger.Error("Error inserting auto sync record", "error", err)
			return err
		}
	}

	return nil
}
