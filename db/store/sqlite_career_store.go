package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/db/models"
)

type SqliteCareerStore struct {
	db *sql.DB
}

func NewSqliteCareerStore(db *sql.DB) *SqliteCareerStore {
	return &SqliteCareerStore{db: db}
}

func (s *SqliteCareerStore) Insert(ctx context.Context, c *models.Career) error {
	query := `INSERT INTO careers (career_code, sheet_version_id) VALUES (?, ?)`
	res, err := s.db.ExecContext(ctx, query, c.CareerCode, c.SheetVersionID)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	c.CareerID = id
	return nil
}

func (s *SqliteCareerStore) Delete(ctx context.Context, careerID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM careers WHERE career_id = ?`, careerID)
	return err
}

func (s *SqliteCareerStore) GetByID(ctx context.Context, careerID int64) (*models.Career, error) {
	c := &models.Career{}
	var sheetVersionID sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT career_id, career_code, sheet_version_id
		FROM careers WHERE career_id = ?`, careerID).
		Scan(&c.CareerID, &c.CareerCode, &sheetVersionID)

	if err != nil {
		return nil, err
	}
	c.SheetVersionID = sheetVersionID
	return c, nil
}

func (s *SqliteCareerStore) GetBySheetVersion(ctx context.Context, versionID int64) ([]*models.Career, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT career_id, career_code, sheet_version_id
		FROM careers
		WHERE sheet_version_id = ?
		ORDER BY career_code`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	careers := []*models.Career{}
	for rows.Next() {
		c := &models.Career{}
		var sheetVersionID sql.NullInt64
		if err := rows.Scan(&c.CareerID, &c.CareerCode, &sheetVersionID); err != nil {
			return nil, err
		}
		c.SheetVersionID = sheetVersionID
		careers = append(careers, c)
	}
	return careers, rows.Err()
}
