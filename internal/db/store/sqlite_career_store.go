package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteCareerStore struct {
}

func NewSqliteCareerStore() *SqliteCareerStore {
	return &SqliteCareerStore{}
}

func (s SqliteCareerStore) Insert(ctx context.Context, exec Executor, c *model.Career) error {
	// Try to find existing career by career_code
	var careerID int64
	err := exec.QueryRowContext(ctx, `SELECT career_id FROM career WHERE career_code = ?`, c.CareerCode).Scan(&careerID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Career not found, insert new career
			res, err := exec.ExecContext(ctx, `INSERT INTO career (career_code) VALUES (?)`, c.CareerCode)
			if err != nil {
				return err
			}
			careerID, err = res.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Insert new career_version linking career and sheet_version
	res, err := exec.ExecContext(ctx, `
        INSERT INTO career_version (career_id, sheet_version_id)
        VALUES (?, ?)
    `, careerID, c.SheetVersionID)
	if err != nil {
		return err
	}

	careerVersionID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Assign the careerVersionID as the ID of the Career model
	c.ID = careerVersionID

	return nil
}

func (s SqliteCareerStore) Delete(ctx context.Context, exec Executor, careerID int64) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM careers_version WHERE career_version_id = ?`, careerID)
	return err
}

func (s SqliteCareerStore) GetByID(ctx context.Context, exec Executor, careerID int64) (*model.Career, error) {
	c := &model.Career{}

	err := exec.QueryRowContext(ctx, `
		SELECT
			cv.career_id,
			cl.career_code,
			cv.sheet_version_id
		FROM careers cv
		JOIN career cl ON cl.career_id = cv.career_code_id
		WHERE cv.career_id = ?
	`, careerID).Scan(&c.ID, &c.CareerCode, &c.SheetVersionID)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s SqliteCareerStore) GetBySheetVersion(ctx context.Context, exec Executor, versionID int64) ([]*model.Career, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT
			cv.career_version_id,
			c.career_code,
			cv.sheet_version_id
		FROM career_version cv
		JOIN career c ON c.career_id = cv.career_id
		WHERE cv.sheet_version_id = ?
		ORDER BY c.career_code
	`, versionID)
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
