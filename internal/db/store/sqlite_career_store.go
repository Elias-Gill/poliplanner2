package store

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteCareerStore struct {
}

func NewSqliteCareerStore() *SqliteCareerStore {
	return &SqliteCareerStore{}
}

func (s SqliteCareerStore) Insert(ctx context.Context, exec Executor, c *model.Career) error {
	query := `INSERT INTO careers (career_code, sheet_version_id) VALUES (?, ?)`
	res, err := exec.ExecContext(ctx, query, c.CareerCode, c.SheetVersionID)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	c.ID = id
	return nil
}

func (s SqliteCareerStore) Delete(ctx context.Context, exec Executor, careerID int64) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM careers WHERE career_id = ?`, careerID)
	return err
}

func (s SqliteCareerStore) GetByID(ctx context.Context, exec Executor, careerID int64) (*model.Career, error) {
	c := &model.Career{}
	var sheetVersionID int64

	err := exec.QueryRowContext(ctx, `
		SELECT career_id, career_code, sheet_version_id
		FROM careers WHERE career_id = ?`, careerID).
		Scan(&c.ID, &c.CareerCode, &sheetVersionID)

	if err != nil {
		return nil, err
	}
	c.SheetVersionID = sheetVersionID
	return c, nil
}

func (s SqliteCareerStore) GetBySheetVersion(ctx context.Context, exec Executor, versionID int64) ([]*model.Career, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT career_id, career_code, sheet_version_id
		FROM careers
		WHERE sheet_version_id = ?
		ORDER BY career_code`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	careers := []*model.Career{}
	for rows.Next() {
		c := &model.Career{}
		var sheetVersionID int64
		if err := rows.Scan(&c.ID, &c.CareerCode, &sheetVersionID); err != nil {
			return nil, err
		}
		c.SheetVersionID = sheetVersionID
		careers = append(careers, c)
	}
	return careers, rows.Err()
}
