package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const autoSyncRowID = 1

type SQLiteSyncRepository struct {
	db *sql.DB
}

func NewSyncRepository(db *sql.DB) *SQLiteSyncRepository {
	return &SQLiteSyncRepository{db: db}
}

func (s *SQLiteSyncRepository) GetLastSyncAttempt(
	ctx context.Context,
) (*time.Time, error) {

	var lastCheckedAt string

	err := s.db.QueryRowContext(
		ctx,
		`SELECT last_checked_at FROM auto_sync_excel_check WHERE id = ?`,
		autoSyncRowID,
	).Scan(&lastCheckedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("query last sync attempt: %w", err)
	}

	t, err := time.Parse(time.RFC3339, lastCheckedAt)
	if err != nil {
		return nil, fmt.Errorf("parse last sync attempt %q: %w", lastCheckedAt, err)
	}

	return &t, nil
}

func (s *SQLiteSyncRepository) SetLastSyncAttempt(
	ctx context.Context,
	t time.Time,
) error {

	value := t.Format(time.RFC3339)

	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO auto_sync_excel_check (id, last_checked_at)
		VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET
			last_checked_at = excluded.last_checked_at
		`,
		autoSyncRowID,
		value,
	)
	if err != nil {
		return fmt.Errorf("upsert last sync attempt: %w", err)
	}

	return nil
}
