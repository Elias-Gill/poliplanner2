package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/domain/career"
)

type SqliteCareerStore struct {
	db *sql.DB
}

func NewSqliteCareerStore(db *sql.DB) *SqliteCareerStore {
	return &SqliteCareerStore{
		db: db,
	}
}

func (s SqliteCareerStore) List(ctx context.Context) ([]*career.Career, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			siglas
		FROM carreras
		ORDER BY siglas
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	careers := []*career.Career{}
	for rows.Next() {
		c := &career.Career{}
		if err := rows.Scan(&c.ID, &c.Code); err != nil {
			return nil, err
		}
		careers = append(careers, c)
	}

	return careers, rows.Err()
}
