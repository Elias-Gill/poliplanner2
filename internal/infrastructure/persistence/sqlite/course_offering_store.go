package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
)

type SqliteCourseOfferingStore struct {
	db *sql.DB
}

func NewSqliteCourseOfferingStore(connection *sql.DB) *SqliteCourseOfferingStore {
	return &SqliteCourseOfferingStore{db: connection}
}

func (s SqliteCourseOfferingStore) FindOfferForSubject(
	ctx context.Context,
	subjectID academicPlan.SubjectID,
	p period.Period,
) ([]courseOffering.Section, error) {

	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT 
			c.id,
			c.nombre,
			c.seccion,
			c.tipo
		FROM cursos c
		JOIN mallas m ON m.id = c.malla
		JOIN periodos pe ON pe.id = c.periodo
		WHERE 
			m.id = ?
			AND pe.year = ?
			AND pe.periodo = ?
		ORDER BY c.nombre, c.seccion
		`,
		subjectID,
		p.Year,
		int(p.Semester),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query course offerings: %w", err)
	}
	defer rows.Close()

	var sections []courseOffering.Section

	for rows.Next() {
		var sec courseOffering.Section
		var courseID int64
		var tipo int

		err := rows.Scan(
			&courseID,
			&sec.CourseName,
			&sec.Section,
			&tipo,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan course offering: %w", err)
		}

		sec.ID = courseOffering.SectionID(courseID)
		sec.Type = courseOffering.CourseType(tipo)

		// Traer docentes para este curso
		teacherRows, err := s.db.QueryContext(
			ctx,
			`
			SELECT d.nombre || ' ' || d.apellido AS full_name, d.correo
			FROM docentes_curso dc
			JOIN docentes d ON d.id = dc.id_docente
			WHERE dc.id_curso = ?
			ORDER BY d.nombre, d.apellido
			`,
			courseID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query teachers for course %d: %w", courseID, err)
		}

		var teachers []courseOffering.TeacherInfo
		for teacherRows.Next() {
			var t courseOffering.TeacherInfo
			if err := teacherRows.Scan(&t.Name, &t.Email); err != nil {
				teacherRows.Close()
				return nil, fmt.Errorf("failed to scan teacher: %w", err)
			}
			teachers = append(teachers, t)
		}
		teacherRows.Close()

		sec.Teachers = teachers
		sections = append(sections, sec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating course offerings: %w", err)
	}

	return sections, nil
}

func (s SqliteCourseOfferingStore) GetCourseDetails(
	ctx context.Context,
	id courseOffering.CourseOfferingID,
) (*courseOffering.CourseSummary, error) {

	// 1. Traer datos principales del curso, incluyendo SaturdayDates
	row := s.db.QueryRowContext(
		ctx,
		`
		SELECT 
			c.nombre,
			c.seccion,
			c.tipo,
			c.comite_presidente,
			c.comite_miembro1,
			c.comite_miembro2,
			c.fechas_sabados
		FROM cursos c
		WHERE c.id = ?
		LIMIT 1
		`,
		id,
	)

	var cs courseOffering.CourseSummary
	var tipo int
	err := row.Scan(
		&cs.Name,
		&cs.Section,
		&tipo,
		&cs.CommitteePresident,
		&cs.CommitteeMember1,
		&cs.CommitteeMember2,
		&cs.SaturdayDates,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	cs.CourseType = courseOffering.CourseType(tipo)

	// 2. Traer docentes
	teacherRows, err := s.db.QueryContext(
		ctx,
		`
		SELECT d.nombre || ' ' || d.apellido AS full_name, d.correo
		FROM docentes_curso dc
		JOIN docentes d ON d.id = dc.id_docente
		WHERE dc.id_curso = ?
		ORDER BY d.nombre, d.apellido
		`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query teachers: %w", err)
	}
	defer teacherRows.Close()

	var teachers []courseOffering.TeacherInfo
	for teacherRows.Next() {
		var t courseOffering.TeacherInfo
		if err := teacherRows.Scan(&t.Name, &t.Email); err != nil {
			return nil, fmt.Errorf("failed to scan teacher: %w", err)
		}
		teachers = append(teachers, t)
	}
	cs.Teachers = teachers

	return &cs, nil
}

func (s SqliteCourseOfferingStore) GetCoursesSchedules(
	ctx context.Context,
	ids []courseOffering.CourseOfferingID,
) ([]courseOffering.CourseClass, error) {

	if len(ids) == 0 {
		return nil, nil
	}

	// Construir placeholders para el IN
	args := make([]any, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	rows, err := s.db.QueryContext(
		ctx,
		fmt.Sprintf(`
			SELECT 
				ch.curso_id,
				c.nombre,
				ch.dia,
				ch.aula,
				ch.desde,
				ch.hasta
			FROM curso_horarios ch
			JOIN cursos c ON c.id = ch.curso_id
			WHERE ch.curso_id IN (%s)
			ORDER BY ch.curso_id, ch.dia, ch.desde
		`, inClause),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query course schedules: %w", err)
	}
	defer rows.Close()

	var classes []courseOffering.CourseClass

	for rows.Next() {
		var c courseOffering.CourseClass
		var day int
		var startStr, endStr string

		err := rows.Scan(
			&c.CourseID,
			&c.Name,
			&day,
			&c.Room,
			&startStr,
			&endStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan course schedule: %w", err)
		}

		c.Day = courseOffering.WeekDay(day)

		// Parsear horas HH:MM
		c.Start, err = time.Parse("15:04", startStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
		c.End, err = time.Parse("15:04", endStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}

		classes = append(classes, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating course schedules: %w", err)
	}

	return classes, nil
}

func (s SqliteCourseOfferingStore) GetCoursesExams(
	ctx context.Context,
	ids []courseOffering.CourseOfferingID,
) ([]courseOffering.ExamClass, error) {

	if len(ids) == 0 {
		return nil, nil
	}

	// Build placeholders for IN clause
	args := make([]any, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	query := fmt.Sprintf(`
		SELECT 
			e.id,
			c.nombre,
			e.aula,
			e.fecha,
			e.hora,
			e.revision_fecha,
			e.revision_hora,
			e.tipo,
			e.instancia
		FROM examenes e
		JOIN cursos c ON c.id = e.curso_id
		WHERE e.curso_id IN (%s)
		ORDER BY e.curso_id, e.fecha, e.hora
	`, inClause)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query course exams: %w", err)
	}
	defer rows.Close()

	var exams []courseOffering.ExamClass

	for rows.Next() {
		var e courseOffering.ExamClass
		var dateStr, timeStr, tipo string
		var instancia int
		var revDate, revTime sql.NullString

		if err := rows.Scan(
			&e.ID,
			&e.CourseName,
			&e.Room,
			&dateStr,
			&timeStr,
			&revDate,
			&revTime,
			&tipo,
			&instancia,
		); err != nil {
			return nil, fmt.Errorf("failed to scan exam: %w", err)
		}

		// Parse main exam datetime (SQLite stores ISO8601)
		dateTime, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse exam date (RFC3339): %w", err)
		}
		t, err := time.Parse("15:04", timeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse exam time: %w", err)
		}
		e.Date = time.Date(
			dateTime.Year(), dateTime.Month(), dateTime.Day(),
			t.Hour(), t.Minute(), 0, 0,
			dateTime.Location(),
		)

		// Parse revision datetime if present
		if revDate.Valid && revTime.Valid {
			revDateTime, err := time.Parse(time.RFC3339, revDate.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse review date: %w", err)
			}
			revTimePart, err := time.Parse("15:04", revTime.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse review time: %w", err)
			}
			rev := time.Date(
				revDateTime.Year(), revDateTime.Month(), revDateTime.Day(),
				revTimePart.Hour(), revTimePart.Minute(), 0, 0,
				revDateTime.Location(),
			)
			e.Revision = &rev
		}

		// Assign type and instance
		e.Type = courseOffering.ExamType(tipo)
		e.Instance = courseOffering.ExamInstance(instancia)

		exams = append(exams, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating exams: %w", err)
	}

	return exams, nil
}
