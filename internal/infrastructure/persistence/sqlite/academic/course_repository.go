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

// FIX: IMPORTANTE, TODAS LAS OPERACIONES DE LOS REPOSITORIOS DEBERIA DE USAR GetExecutor

type CourseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) Upsert(ctx context.Context, course *academicRepo.CourseSaveParams) (academic.CourseID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	var courseID int64
	err := exec.QueryRowContext(ctx, `
		INSERT INTO cursos (
		malla, periodo, nombre, seccion, turno, tipo,
		comite_presidente, comite_miembro1, comite_miembro2, fechas_sabados
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(nombre, malla, seccion, periodo, turno) DO UPDATE SET
		tipo = excluded.tipo,
		comite_presidente = excluded.comite_presidente,
		comite_miembro1 = excluded.comite_miembro1,
		comite_miembro2 = excluded.comite_miembro2,
		fechas_sabados = excluded.fechas_sabados
		RETURNING id
		`,
		course.Curriculum,
		course.Period,
		course.Name,
		course.Section,
		course.Shift,
		int(course.Type),
		course.Comitee.President,
		course.Comitee.Member1,
		course.Comitee.Member2,
		course.SaturdayDates,
	).Scan(&courseID)

	if err != nil {
		return 0, err
	}

	return academic.CourseID(courseID), nil
}

func (r *CourseRepository) AssignTeachers(ctx context.Context, courseID academic.CourseID, teachers []academic.TeacherID) error {
	exec := txManager.GetExecutor(ctx, r.db)

	_, err := exec.ExecContext(ctx, `DELETE FROM docentes_curso WHERE id_curso = ?`, courseID)
	if err != nil {
		return err
	}

	stmt, err := exec.PrepareContext(ctx, `
		INSERT INTO docentes_curso (id_docente, id_curso) VALUES (?, ?)
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tid := range teachers {
		if _, err := stmt.ExecContext(ctx, tid, courseID); err != nil {
			return err
		}
	}

	return nil
}

func (r *CourseRepository) AssignSchedule(ctx context.Context, courseID academic.CourseID, schedule []academic.ClassSession) error {
	exec := txManager.GetExecutor(ctx, r.db)

	_, err := exec.ExecContext(ctx, `DELETE FROM curso_horarios WHERE curso_id = ?`, courseID)
	if err != nil {
		return err
	}

	stmt, err := exec.PrepareContext(ctx, `
		INSERT INTO curso_horarios (curso_id, dia, desde, hasta, aula)
		VALUES (?, ?, ?, ?, ?)
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range schedule {
		if s.Time.Start == nil || s.Time.End == nil {
			continue
		}

		_, err := stmt.ExecContext(ctx,
			courseID,
			int(s.Day),
			s.Time.Start.Format("15:04"),
			s.Time.End.Format("15:04"),
			s.Room,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CourseRepository) AssignExams(ctx context.Context, courseID academic.CourseID, exams []academic.Exam) error {
	exec := txManager.GetExecutor(ctx, r.db)

	_, err := exec.ExecContext(ctx, `DELETE FROM examenes WHERE curso_id = ?`, courseID)
	if err != nil {
		return err
	}

	stmt, err := exec.PrepareContext(ctx, `
		INSERT INTO examenes (
		curso_id, tipo, instancia,
		fecha, hora, aula,
		revision_fecha, revision_hora
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range exams {
		examDate := e.Date()
		if examDate == nil {
			continue
		}

		var typeStr string
		switch e.Type {
		case academic.ExamPartial:
			typeStr = "partial"
		case academic.ExamFinal:
			typeStr = "final"
		default:
			typeStr = "unknown"
		}

		var revDate any
		var revTime any
		if revision := e.Revision(); revision != nil {
			revDate = revision.Format("2006-01-02")
			revTime = revision.Format("15:04")
		}

		_, err := stmt.ExecContext(ctx,
			courseID,
			typeStr,
			e.Instance,
			examDate.Format("2006-01-02"),
			examDate.Format("15:04"),
			e.Room,
			revDate,
			revTime,
		)
		if err != nil {
			return fmt.Errorf("failed to execute exam insert stmt: %w", err)
		}
	}

	return nil
}

func (r *CourseRepository) ListByCurriculumID(ctx context.Context, curriculum academic.CurriculumID, period academic.PeriodID) ([]academic.CourseSummaryView, error) {
	query := `
		SELECT id, seccion, turno, tipo, nombre
		FROM cursos
		WHERE malla = $1 AND periodo = $2
	`

	rows, err := r.db.QueryContext(ctx, query, curriculum, period)
	if err != nil {
		return nil, fmt.Errorf("list courses by curriculum: %w", err)
	}
	defer rows.Close()

	var courses []academic.CourseSummaryView
	for rows.Next() {
		var c academic.CourseSummaryView
		if err := rows.Scan(&c.ID, &c.Section, &c.Shift, &c.Type, &c.Name); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *CourseRepository) GetCourseTeachers(ctx context.Context, courseID academic.CourseID) ([]academic.Teacher, error) {
	query := `
		SELECT d.titulo, d.nombre, d.apellido
		FROM docentes_curso dc join docentes d on d.id = dc.id_docente
		WHERE dc.id_curso = $1
	`

	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("get course teachers: %w", err)
	}
	defer rows.Close()

	var teachers []academic.Teacher
	for rows.Next() {
		var t academic.Teacher
		if err := rows.Scan(&t.Title, &t.FirstName, &t.LastName); err != nil {
			return nil, fmt.Errorf("scan teacher: %w", err)
		}
		teachers = append(teachers, t)
	}

	return teachers, nil
}

func (r *CourseRepository) GetCourseSchedules(ctx context.Context, courseID academic.CourseID) ([]academic.ClassSession, error) {
	query := `
		SELECT dia, desde, hasta, aula
		FROM curso_horarios
		WHERE curso_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("get course schedules: %w", err)
	}
	defer rows.Close()

	const timeLayout = "15:04"

	var schedules []academic.ClassSession
	for rows.Next() {
		var (
			s        academic.ClassSession
			startStr string
			endStr   string
		)

		if err := rows.Scan(&s.Day, &startStr, &endStr, &s.Room); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}

		if startStr != "" {
			t, err := time.Parse(timeLayout, startStr)
			if err != nil {
				return nil, fmt.Errorf("parse start time %q: %w", startStr, err)
			}
			s.Time.Start = &t
		}

		if endStr != "" {
			t, err := time.Parse(timeLayout, endStr)
			if err != nil {
				return nil, fmt.Errorf("parse end time %q: %w", endStr, err)
			}
			s.Time.End = &t
		}

		schedules = append(schedules, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schedule rows: %w", err)
	}

	return schedules, nil
}
