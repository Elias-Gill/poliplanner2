package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sheetversion "github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
	"github.com/elias-gill/poliplanner2/logger"
)

const last_check_entry_const_id = 1

type SqliteSheetVersionStore struct {
	db *sql.DB
}

func NewSqliteSheetVersionStore(db *sql.DB) *SqliteSheetVersionStore {
	return &SqliteSheetVersionStore{
		db: db,
	}
}

func (s SqliteSheetVersionStore) Insert(ctx context.Context, sv *sheetversion.SheetVersion) error {
	query := `
	INSERT INTO sheet_version (file_name, url)
	VALUES (?, ?)
	`
	res, err := s.db.ExecContext(ctx, query, sv.FileName, sv.URL)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	sv.ID = sheetversion.SheetVersionID(id)
	return nil
}

func (s SqliteSheetVersionStore) GetNewest(ctx context.Context) (*sheetversion.SheetVersion, error) {
	sv := &sheetversion.SheetVersion{}
	err := s.db.QueryRowContext(ctx, `
		SELECT version_id, file_name, url, parsed_at
		FROM sheet_version
		ORDER BY parsed_at DESC
		LIMIT 1
	`).Scan(&sv.ID, &sv.FileName, &sv.URL, &sv.ParsedAt)

	if err == sql.ErrNoRows {
		return nil, sheetversion.ErrNoSheetVersion
	}
	if err != nil {
		return nil, fmt.Errorf("database error fetching newest sheet version: %w", err)
	}

	return sv, nil
}

func (s SqliteSheetVersionStore) GetLastCheckedAt(
	ctx context.Context,
) (*time.Time, error) {

	var lastCheckedAt string

	err := s.db.QueryRowContext(
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

func (s SqliteSheetVersionStore) SetLastCheckedAt(
	ctx context.Context,
	t time.Time,
) error {

	value := t.Format(time.RFC3339)

	result, err := s.db.ExecContext(
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
		_, err := s.db.ExecContext(
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

func (s *SqliteSheetVersionStore) Save(
	ctx context.Context,
	fileName string,
	URI string,
	processedSheets int,
	succeededSheets int,
	errors []error,
) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	var versionID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO sheet_version (
			file_name,
			url,
			parsed_at,
			processed_sheets,
			succeeded_sheets,
			error_count
		) VALUES (?, ?, ?, datetime('now'), ?, ?)
		RETURNING version_id
	`,
		fileName,
		URI,
		processedSheets,
		succeededSheets,
		len(errors),
	).Scan(&versionID)
	if err != nil {
		return 0, fmt.Errorf("error inserting sheet version: %w", err)
	}

	for _, e := range errors {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO sheet_errors (version_id, error_message)
			VALUES (?, ?)
		`, versionID, e.Error())
		if err != nil {
			return 0, fmt.Errorf("error inserting sheet error: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return versionID, nil
}
