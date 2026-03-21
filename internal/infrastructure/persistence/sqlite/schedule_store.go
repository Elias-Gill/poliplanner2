package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type SqliteScheduleStore struct {
	db *sql.DB
}

func NewSqliteScheduleStore(db *sql.DB) *SqliteScheduleStore {
	return &SqliteScheduleStore{
		db: db,
	}
}

// ============================================================
// ScheduleStorer
// ============================================================

func (s *SqliteScheduleStore) Save(ctx context.Context, sched schedule.Schedule) (schedule.ScheduleID, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO horarios (usuario_id, descripcion, periodo_id, creado_en)
		VALUES (?, ?, ?, ?)
		RETURNING id
	`, sched.Owner, sched.Description, sched.PeriodID, time.Now().In(timezone.ParaguayTZ)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert horario: %w", err)
	}

	// Insert schedule details
	for _, courseID := range sched.Courses {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO horarios_detalle (horario_id, curso_id)
			VALUES (?, ?)
		`, id, courseID)
		if err != nil {
			return 0, fmt.Errorf("insert horario_detalle: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return schedule.ScheduleID(id), nil
}

func (s *SqliteScheduleStore) Delete(ctx context.Context, scheduleID schedule.ScheduleID) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM horarios
		WHERE id = ?
	`, scheduleID)
	if err != nil {
		return fmt.Errorf("delete horario: %w", err)
	}
	return nil
}

// ============================================================
// ScheduleReadStorer
// ============================================================

func (s *SqliteScheduleStore) ListByUser(ctx context.Context, userID user.UserID) ([]schedule.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, descripcion, periodo_id, creado_en
		FROM horarios
		WHERE usuario_id = ?
		ORDER BY creado_en DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query horarios: %w", err)
	}
	defer rows.Close()

	var schedules []schedule.Schedule
	for rows.Next() {
		var sched schedule.Schedule
		sched.Owner = userID
		err := rows.Scan(
			&sched.ID,
			&sched.Description,
			&sched.PeriodID,
			&sched.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan horario: %w", err)
		}
		// Load course IDs
		courseRows, err := s.db.QueryContext(ctx, `
			SELECT curso_id
			FROM horarios_detalle
			WHERE horario_id = ?
		`, sched.ID)
		if err != nil {
			return nil, fmt.Errorf("query horario_detalle: %w", err)
		}

		var courseIDs []courseOffering.CourseOfferingID
		for courseRows.Next() {
			var cid courseOffering.CourseOfferingID
			if err := courseRows.Scan(&cid); err != nil {
				courseRows.Close()
				return nil, fmt.Errorf("scan curso_id: %w", err)
			}
			courseIDs = append(courseIDs, cid)
		}
		courseRows.Close()
		sched.Courses = courseIDs

		schedules = append(schedules, sched)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}
