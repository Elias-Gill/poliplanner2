package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type PeriodRepository struct {
	db *sql.DB
}

func NewPeriodRepository(db *sql.DB) *PeriodRepository {
	return &PeriodRepository{db: db}
}

func (r *PeriodRepository) Upsert(ctx context.Context, p academic.Period) (academic.PeriodID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	_, err := exec.ExecContext(ctx, `
		INSERT INTO periodos (year, periodo)
		VALUES (?, ?)
		ON CONFLICT(year, periodo) DO NOTHING
		`, p.Year, p.Semester)
	if err != nil {
		return 0, err
	}

	var id int64
	err = exec.QueryRowContext(ctx, `
		SELECT id FROM periodos WHERE year = ? AND periodo = ?
		`, p.Year, p.Semester).Scan(&id)
	if err != nil {
		return 0, err
	}

	return academic.PeriodID(id), nil
}
