package store

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/db/model"
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
