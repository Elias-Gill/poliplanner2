package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

// ListByCurriculum returns every laboratory offering of a curriculum within a period,
// grouped with their weekly schedules.
func (r *LaboratoryRepository) ListByCurriculum(
	ctx context.Context,
	curriculum academic.CurriculumID,
	period academic.PeriodID,
) ([]academic.Laboratory, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	query := `
		SELECT
			l.id,
			l.seccion,
			COALESCE(h.dia, 0),
			COALESCE(h.desde, ''),
			COALESCE(h.hasta, ''),
			COALESCE(h.aula, '')
		FROM laboratorios l
		LEFT JOIN laboratorio_horarios h ON h.laboratorio = l.id
		WHERE l.malla = ? AND l.periodo = ?
		ORDER BY l.id, h.dia, h.desde
	`

	rows, err := exec.QueryContext(ctx, query, curriculum, period)
	if err != nil {
		return nil, fmt.Errorf("list laboratories by curriculum: %w", err)
	}
	defer rows.Close()

	const timeLayout = "15:04"

	var labs []academic.Laboratory
	var current *academic.Laboratory

	for rows.Next() {
		var (
			labID            int64
			section          string
			day              int
			startStr, endStr string
			room             string
		)

		if err := rows.Scan(&labID, &section, &day, &startStr, &endStr, &room); err != nil {
			return nil, fmt.Errorf("scan laboratory: %w", err)
		}

		if current == nil || current.ID != academic.LaboratoryID(labID) {
			labs = append(labs, academic.Laboratory{
				ID:      academic.LaboratoryID(labID),
				Section: section,
			})
			current = &labs[len(labs)-1]
		}

		session, err := buildClassSession(day, startStr, endStr, room, timeLayout)
		if err != nil {
			return nil, err
		}
		if session == nil {
			continue
		}

		current.Schedule = append(current.Schedule, *session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate laboratory rows: %w", err)
	}

	return labs, nil
}

// buildClassSession converts a persisted schedule row into a ClassSession. It returns nil
// when the row has no meaningful day/time data (laboratory without schedule).
func buildClassSession(day int, startStr, endStr, room, layout string) (*academic.ClassSession, error) {
	if day == 0 || startStr == "" || endStr == "" {
		return nil, nil
	}

	start, err := time.Parse(layout, startStr)
	if err != nil {
		return nil, fmt.Errorf("parse start time %q: %w", startStr, err)
	}

	end, err := time.Parse(layout, endStr)
	if err != nil {
		return nil, fmt.Errorf("parse end time %q: %w", endStr, err)
	}

	return &academic.ClassSession{
		Day:  academic.WeekDay(day),
		Room: room,
		Time: academic.TimeSlot{Start: &start, End: &end},
	}, nil
}
