package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (s *SqliteScheduleStore) Save(ctx context.Context, sche schedule.Schedule) (schedule.ScheduleID, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO horarios(usuario_id, descripcion, creado_en)
		VALUES (?, ?, ?)`,
		sche.Owner, sche.Description, time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, c := range sche.Courses {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO horarios_detalle(horario_id, curso_id)
			VALUES (?, ?)`,
			id, c,
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return schedule.ScheduleID(id), nil
}

func (s *SqliteScheduleStore) ListByUserID(ctx context.Context, ownerID user.UserID) ([]schedule.ScheduleBasicData, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, descripcion, creado_en
		FROM horarios
		WHERE usuario_id = ?`, ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []schedule.ScheduleBasicData
	for rows.Next() {
		var sbd schedule.ScheduleBasicData
		if err := rows.Scan(&sbd.ID, &sbd.Description); err != nil {
			return nil, err
		}
		list = append(list, sbd)
	}

	return list, nil
}

func (s *SqliteScheduleStore) GetDetailsByID(ctx context.Context, ID schedule.ScheduleID) (*schedule.Schedule, error) {
	var sch schedule.Schedule
	var created string
	row := s.db.QueryRowContext(ctx, `
		SELECT usuario_id, descripcion, creado_en
		FROM horarios
		WHERE id = ?`, ID)
	if err := row.Scan(&sch.Owner, &sch.Description, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schedule not found")
		}
		return nil, err
	}
	sch.ID = ID
	sch.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)

	rows, err := s.db.QueryContext(ctx, `
		SELECT curso_id
		FROM horarios_detalle
		WHERE horario_id = ?`, ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int64
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		sch.Courses = append(sch.Courses, courseOffering.CourseOfferingID(cid))
	}

	return &sch, nil
}

func (s *SqliteScheduleStore) Delete(ctx context.Context, scheduleID schedule.ScheduleID) error {
	res, err := s.db.ExecContext(ctx, `
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
