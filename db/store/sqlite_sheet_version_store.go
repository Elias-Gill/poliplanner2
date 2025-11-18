package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/db/models"
)

type SqliteSheetVersionStore struct {
	db *sql.DB
}

func NewSqliteSheetVersionStore(db *sql.DB) *SqliteSheetVersionStore {
	return &SqliteSheetVersionStore{db: db}
}

func (s *SqliteSheetVersionStore) Insert(ctx context.Context, sv *models.SheetVersion) (int64, error) {
	query := `
	INSERT INTO sheet_version (file_name, url)
	VALUES (?, ?)
	`
	res, err := s.db.ExecContext(ctx, query, sv.FileName, sv.URL)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *SqliteSheetVersionStore) GetNewest(ctx context.Context) (*models.SheetVersion, error) {
	sv := &models.SheetVersion{}
	err := s.db.QueryRowContext(ctx, `
		SELECT version_id, file_name, url, parsed_at
		FROM sheet_version
		ORDER BY parsed_at DESC
		LIMIT 1
		`).Scan(&sv.VersionID, &sv.FileName, &sv.URL, &sv.ParsedAt)

	if err != nil {
		return nil, err
	}
	return sv, nil
}
