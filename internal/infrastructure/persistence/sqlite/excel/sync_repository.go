package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/excel"
	txManager "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
)

type SQLiteSyncRepository struct {
	db *sql.DB
}

func NewSyncRepository(db *sql.DB) *SQLiteSyncRepository {
	return &SQLiteSyncRepository{db: db}
}

func (s *SQLiteSyncRepository) GetSyncState(ctx context.Context, kind excel.SourceType) (*excel.SyncState, error) {
	exec := txManager.GetExecutor(ctx, s.db)

	var lastSearch, lastSync sql.NullString

	err := exec.QueryRowContext(
		ctx,
		`SELECT last_search_at, last_sync_at FROM excel_sync_state WHERE source_type = ?`,
		string(kind),
	).Scan(&lastSearch, &lastSync)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &excel.SyncState{SourceType: kind}, nil
		}

		return nil, fmt.Errorf("query sync state: %w", err)
	}

	state := &excel.SyncState{SourceType: kind}

	if lastSearch.Valid {
		t, err := time.Parse(time.RFC3339, lastSearch.String)
		if err != nil {
			return nil, fmt.Errorf("parse last search attempt %q: %w", lastSearch.String, err)
		}

		state.LastSearchAt = &t
	}

	if lastSync.Valid {
		t, err := time.Parse(time.RFC3339, lastSync.String)
		if err != nil {
			return nil, fmt.Errorf("parse last sync attempt %q: %w", lastSync.String, err)
		}

		state.LastSyncAt = &t
	}

	return state, nil
}

func (s *SQLiteSyncRepository) SetLastSearchAttempt(ctx context.Context, kind excel.SourceType, t time.Time) error {
	return s.setAttempt(ctx, kind, "last_search_at", t)
}

func (s *SQLiteSyncRepository) SetLastSyncAttempt(ctx context.Context, kind excel.SourceType, t time.Time) error {
	return s.setAttempt(ctx, kind, "last_sync_at", t)
}

func (s *SQLiteSyncRepository) setAttempt(ctx context.Context, kind excel.SourceType, column string, t time.Time) error {
	exec := txManager.GetExecutor(ctx, s.db)

	_, err := exec.ExecContext(
		ctx,
		fmt.Sprintf(`
		INSERT INTO excel_sync_state (source_type, %[1]s)
		VALUES (?, ?)
		ON CONFLICT(source_type) DO UPDATE SET %[1]s = excluded.%[1]s
		`, column),
		string(kind),
		t.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("upsert %s: %w", column, err)
	}

	return nil
}
