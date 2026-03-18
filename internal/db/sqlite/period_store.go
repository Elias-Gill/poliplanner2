package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/logger"
)

type SqlitePeriodStore struct {
	db *sql.DB
}

func NewSqlitePeriodStore(db *sql.DB) *SqlitePeriodStore {
	return &SqlitePeriodStore{
		db: db,
	}
}

func (s *SqlitePeriodStore) FindByYearPeriod(ctx context.Context, year int, p int) (*period.Period, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT
		id,
		year,
		periodo
		FROM periodos
		WHERE year = ? AND periodo = ?
		`, year, p)

	var auxP period.Period
	err := row.Scan(
		&auxP.ID,
		&auxP.Year,
		&auxP.Period,
	)
	if err == sql.ErrNoRows {
		logger.Debug("No period found", "year", year, "period", period)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}
