package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	txManager "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type LaboratoryRepository struct {
	db *sql.DB
}

func NewLaboratoryRepository(db *sql.DB) *LaboratoryRepository {
	return &LaboratoryRepository{db: db}
}

func (r *LaboratoryRepository) Upsert(
	ctx context.Context,
	params academicRepo.LaboratorySaveParams,
) (academic.LaboratoryID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	var labID int64
	err := exec.QueryRowContext(ctx, `
		INSERT INTO laboratorios (malla, periodo, seccion)
		VALUES (?, ?, ?)
		ON CONFLICT(malla, seccion, periodo) DO UPDATE SET seccion = excluded.seccion
		RETURNING id
	`, params.Curriculum, params.Period, params.Section).Scan(&labID)
	if err != nil {
		return 0, fmt.Errorf("upsert laboratory: %w", err)
	}

	if _, err := exec.ExecContext(ctx, `DELETE FROM laboratorio_horarios WHERE laboratorio = ?`, labID); err != nil {
		return 0, fmt.Errorf("clear laboratory schedule: %w", err)
	}

	stmt, err := exec.PrepareContext(ctx, `
		INSERT INTO laboratorio_horarios (laboratorio, dia, desde, hasta, aula)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare laboratory schedule insert: %w", err)
	}
	defer stmt.Close()

	for _, s := range params.Schedule {
		if s.Time.Start == nil || s.Time.End == nil {
			continue
		}

		_, err := stmt.ExecContext(
			ctx,
			labID,
			int(s.Day),
			s.Time.Start.Format("15:04"),
			s.Time.End.Format("15:04"),
			s.Room,
		)
		if err != nil {
			return 0, fmt.Errorf("insert laboratory schedule: %w", err)
		}
	}

	return academic.LaboratoryID(labID), nil
}
