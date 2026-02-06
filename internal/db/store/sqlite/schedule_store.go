package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteScheduleStore struct {
	db *sql.DB
}

func NewSqliteScheduleStore(db *sql.DB) *SqliteScheduleStore {
	return &SqliteScheduleStore{
		db: db,
	}
}

func (s *SqliteScheduleStore) ListByUserID(ctx context.Context, userID int64) ([]*model.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, nombre, descripcion, periodo_id, creado_en
		FROM horarios
		WHERE usuario_id = ?
		ORDER BY creado_en DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.Schedule
	for rows.Next() {
		sched := &model.Schedule{}
		err := rows.Scan(
			&sched.ID,
			&sched.Name,
			&sched.Description,
			&sched.PeriodID,
			&sched.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sched.OwnerID = userID // ya filtramos por userID
		schedules = append(schedules, sched)
	}

	return schedules, rows.Err()
}

func (s *SqliteScheduleStore) Insert(ctx context.Context, data *model.ScheduleBasicData) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var scheduleID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO horarios (usuario_id, nombre, descripcion, periodo_id)
		VALUES (?, ?, ?, ?)
		RETURNING id
		`, data.Owner, data.Name, data.Description, 1).
		Scan(&scheduleID)
	if err != nil {
		return 0, err
	}

	// Vincular cursos
	for _, courseID := range data.CourseIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO horarios_detalle (horario_id, curso_id)
			VALUES (?, ?)
		`, scheduleID, courseID)
		if err != nil {
			return 0, err
		}
	}

	return scheduleID, tx.Commit()
}

func (s *SqliteScheduleStore) Delete(ctx context.Context, scheduleID int64) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM horarios WHERE id = ?
	`, scheduleID)
	return err
}

func (s *SqliteScheduleStore) GetByUserID(ctx context.Context, userID int64) ([]*model.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, creado_en, nombre, descripcion, periodo_id
		FROM horarios
		WHERE usuario_id = ?
		ORDER BY creado_en DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query horarios: %w", err)
	}
	defer rows.Close()

	var schedules []*model.Schedule
	for rows.Next() {
		var sched model.Schedule
		err := rows.Scan(
			&sched.ID,
			&sched.CreatedAt,
			&sched.Name,
			&sched.Description,
			&sched.PeriodID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan horario: %w", err)
		}
		schedules = append(schedules, &sched)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}

func (s *SqliteScheduleStore) GetByID(ctx context.Context, scheduleID int64) (*model.ScheduleDetails, error) {
	var sched model.Schedule
	var ownerID int64

	err := s.db.QueryRowContext(ctx, `
        SELECT id, creado_en, usuario_id, nombre, descripcion, periodo_id
        FROM horarios
        WHERE id = ?`,
		scheduleID,
	).Scan(
		&sched.ID,
		&sched.CreatedAt,
		&ownerID,
		&sched.Name,
		&sched.Description,
		&sched.PeriodID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("schedule not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error finding schedule: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
        SELECT 
            c.nombre, c.seccion,
            c.lunes_desde, c.lunes_hasta, c.lunes_aula,
            c.martes_desde, c.martes_hasta, c.martes_aula,
            c.miercoles_desde, c.miercoles_hasta, c.miercoles_aula,
            c.jueves_desde, c.jueves_hasta, c.jueves_aula,
            c.viernes_desde, c.viernes_hasta, c.viernes_aula,
            c.sabado_desde, c.sabado_hasta, c.sabado_aula, c.sabado_night_fechas,
            c.partial1_fecha, c.partial1_hora, c.partial1_aula,
            c.partial2_fecha, c.partial2_hora, c.partial2_aula,
            c.final1_fecha, c.final1_hora, c.final1_aula, c.final1_fecha_revision, c.final1_hora_revision,
            c.final2_fecha, c.final2_hora, c.final2_aula, c.final2_fecha_revision, c.final2_hora_revision,
            c.comite_presidente, c.comite_miembro1, c.comite_miembro2
        FROM horarios_detalle hd
        JOIN cursos c ON hd.curso_id = c.id
        WHERE hd.horario_id = ?`,
		scheduleID,
	)
	if err != nil {
		return nil, fmt.Errorf("obtener cursos: %w", err)
	}
	defer rows.Close()

	var courses []model.CourseModel
	for rows.Next() {
		var gm model.CourseModel
		err := rows.Scan(
			&gm.Name,
			&gm.Section,
			&gm.Monday.Start, &gm.Monday.End, &gm.MondayRoom,
			&gm.Tuesday.Start, &gm.Tuesday.End, &gm.TuesdayRoom,
			&gm.Wednesday.Start, &gm.Wednesday.End, &gm.WednesdayRoom,
			&gm.Thursday.Start, &gm.Thursday.End, &gm.ThursdayRoom,
			&gm.Friday.Start, &gm.Friday.End, &gm.FridayRoom,
			&gm.Saturday.Start, &gm.Saturday.End, &gm.SaturdayRoom, &gm.SaturdayDates,
			&gm.Partial1Date, &gm.Partial1Time, &gm.Partial1Room,
			&gm.Partial2Date, &gm.Partial2Time, &gm.Partial2Room,
			&gm.Final1Date, &gm.Final1Time, &gm.Final1Room, &gm.Final1RevDate, &gm.Final1RevTime,
			&gm.Final2Date, &gm.Final2Time, &gm.Final2Room, &gm.Final2RevDate, &gm.Final2RevTime,
			&gm.CommitteePresident, &gm.CommitteeMember1, &gm.CommitteeMember2,
		)
		if err != nil {
			return nil, fmt.Errorf("scan curso: %w", err)
		}
		courses = append(courses, gm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar cursos: %w", err)
	}

	return &model.ScheduleDetails{
		Schedule: sched,
		Courses:  courses,
	}, nil
}
