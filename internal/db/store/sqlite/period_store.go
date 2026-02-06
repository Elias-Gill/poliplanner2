package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqlitePeriodStore struct {
	db *sql.DB
}

func NewSqlitePeriodStore(db *sql.DB) *SqlitePeriodStore {
	return &SqlitePeriodStore{
		db: db,
	}
}

func (s *SqlitePeriodStore) FindByYearPeriod(ctx context.Context, year int, period int) (*model.Period, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT
		id,
		year,
		periodo
		FROM periodos
		WHERE year = ? AND periodo = ?
		`, year, period)

	var p model.Period
	err := row.Scan(
		&p.ID,
		&p.Year,
		&p.Period,
	)
	if err == sql.ErrNoRows {
		return nil, nil // o un error custom si preferís "not found"
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}
