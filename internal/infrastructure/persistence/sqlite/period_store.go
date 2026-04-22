package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/domain/period"
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
	// TODO: implementar
	return nil, nil
}
