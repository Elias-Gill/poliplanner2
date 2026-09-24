package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	txManager "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/model/schedule"
	"github.com/elias-gill/poliplanner2/internal/model/user"
)

type SqliteScheduleStore struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) *SqliteScheduleStore {
	return &SqliteScheduleStore{
		db: db,
	}
}

// ============================================================
// ScheduleStorer
// ============================================================

func (s *SqliteScheduleStore) Save(ctx context.Context, sche schedule.Schedule) (schedule.ScheduleID, error) {
	exec := txManager.GetExecutor(ctx, s.db)

	res, err := exec.ExecContext(ctx, `
		INSERT INTO horarios(usuario_id, titulo, creado_en)
		VALUES (?, ?, ?)`,
		sche.Owner, sche.Title, time.Now().In(timezone.ParaguayTZ),
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, c := range sche.Courses {
		_, err := exec.ExecContext(ctx, `
			INSERT INTO horarios_detalle(horario_id, curso_id)
			VALUES (?, ?)`,
			id, c,
		)
		if err != nil {
			return 0, err
		}
	}

	return schedule.ScheduleID(id), nil
}

func (s *SqliteScheduleStore) ListByUserID(ctx context.Context, ownerID user.UserID) ([]schedule.ScheduleSummaryView, error) {
	exec := txManager.GetExecutor(ctx, s.db)

	rows, err := exec.QueryContext(ctx, `
		SELECT id, titulo
		FROM horarios
		WHERE usuario_id = ?`, ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []schedule.ScheduleSummaryView
	for rows.Next() {
		var sbd schedule.ScheduleSummaryView
		if err := rows.Scan(&sbd.ID, &sbd.Title); err != nil {
			return nil, err
		}
		list = append(list, sbd)
	}

	return list, nil
}

func (s *SqliteScheduleStore) GetByID(ctx context.Context, ID schedule.ScheduleID) (*schedule.ScheduleDetails, error) {
	exec := txManager.GetExecutor(ctx, s.db)

	var sch schedule.ScheduleDetails
	var created string

	row := exec.QueryRowContext(ctx, `
        SELECT usuario_id, titulo, creado_en
        FROM horarios
        WHERE id = ?`, ID)

	if err := row.Scan(&sch.Owner, &sch.Title, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("schedule not found")
		}
		return nil, fmt.Errorf("failed to scan schedule: %w", err)
	}

	sch.ID = ID
	t, _ := time.Parse("2006-01-02 15:04:05", created)
	sch.CreatedAt = t

	courseRows, err := exec.QueryContext(ctx, `
        SELECT curso_id 
        FROM horarios_detalle 
        WHERE horario_id = ?`, ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedule courses: %w", err)
	}
	defer courseRows.Close()

	var courses []academic.CourseID
	for courseRows.Next() {
		var courseID academic.CourseID
		if err := courseRows.Scan(&courseID); err != nil {
			return nil, fmt.Errorf("failed to scan course id: %w", err)
		}
		courses = append(courses, courseID)
	}

	if err := courseRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating course rows: %w", err)
	}

	sch.Courses = courses

	return &sch, nil
}

func (s *SqliteScheduleStore) Delete(ctx context.Context, scheduleID schedule.ScheduleID) error {
	exec := txManager.GetExecutor(ctx, s.db)

	res, err := exec.ExecContext(ctx, `
		DELETE FROM horarios
		WHERE id = ?`, scheduleID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("schedule not found")
	}

	return nil
}
