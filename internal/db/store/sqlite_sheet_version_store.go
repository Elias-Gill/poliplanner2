package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/logger"
)

type SqliteSheetVersionStore struct {
}

func NewSqliteSheetVersionStore() *SqliteSheetVersionStore {
	return &SqliteSheetVersionStore{}
}

func (s SqliteSheetVersionStore) Insert(ctx context.Context, exec Executor, sv *model.SheetVersion) error {
	query := `
	INSERT INTO sheet_version (file_name, url)
	VALUES (?, ?)
	`
	res, err := exec.ExecContext(ctx, query, sv.FileName, sv.URL)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	sv.ID = id
	return nil
}

func (s SqliteSheetVersionStore) GetNewest(ctx context.Context, exec Executor) (*model.SheetVersion, error) {
	sv := &model.SheetVersion{}
	err := exec.QueryRowContext(ctx, `
		SELECT version_id, file_name, url, parsed_at
		FROM sheet_version
		ORDER BY parsed_at DESC
		LIMIT 1
		`).Scan(&sv.ID, &sv.FileName, &sv.URL, &sv.ParsedAt)

	if err != nil {
		return nil, err
	}
	return sv, nil
}

func (s SqliteSheetVersionStore) HasToUpdate(ctx context.Context, exec Executor) bool {
	const id = 1
	var lastCheckedAt string

	err := exec.QueryRowContext(ctx, `SELECT last_checked_at FROM auto_sync_excel_check WHERE id = ?`, id).Scan(&lastCheckedAt)
	if err != nil {
		// If no row, initialize it and return true to trigger update
		if errors.Is(err, sql.ErrNoRows) {
			_, err := exec.ExecContext(ctx, `INSERT INTO auto_sync_excel_check (id, last_checked_at) VALUES (?, ?)`, id, time.Now().Format(time.RFC3339))
			if err != nil {
				logger.Error("Error on auto sync table insert", "error", err)
			}
			return true
		}

		logger.Error("Error querying auto sync table", "error", err)
		// Do NOT trigger update on error, just log and return false
		return false
	}

	// Parse time
	lastChecked, err := time.Parse(time.RFC3339, lastCheckedAt)
	if err != nil {
		logger.Error("Error parsing last update date", "error", err)
		// Do NOT trigger update on parsing error, just log and return false
		return false
	}

	// Check if last check was more than 2 days ago
	if time.Since(lastChecked) > 48*time.Hour {
		// Update the timestamp to now
		_, err := exec.ExecContext(ctx, `UPDATE auto_sync_excel_check SET last_checked_at = ? WHERE id = ?`, time.Now().Format(time.RFC3339), id)
		if err != nil {
			logger.Error("Error updating last_checked_at", "error", err)
		}
		return true
	}

	return false
}
