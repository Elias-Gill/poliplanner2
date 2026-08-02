package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
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

func (s *SqliteScheduleStore) ListByUserID(ctx context.Context, ownerID user.UserID) ([]schedule.ScheduleSummaryView, error) {
	rows, err := s.db.QueryContext(ctx, `
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

func (s *SqliteScheduleStore) GetDetailsByID(ctx context.Context, ID schedule.ScheduleID) (*schedule.ScheduleDetails, error) {
	var sch schedule.ScheduleDetails
	var created string

	row := s.db.QueryRowContext(ctx, `
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

	coursesQuery := `
		SELECT 
			c.id, 
			c.seccion, 
			c.turno, 
			c.nombre, 
			c.tipo,
			COALESCE(c.fechas_sabados, ''),
			COALESCE(c.comite_presidente, ''),
			COALESCE(c.comite_miembro1, ''),
			COALESCE(c.comite_miembro2, '')
		FROM horarios_detalle hd
		JOIN cursos c ON hd.curso_id = c.id
		JOIN mallas m ON c.malla = m.id
		JOIN asignaturas a ON m.asignatura = a.id
		WHERE hd.horario_id = ?
		ORDER BY c.id ASC`

	rows, err := s.db.QueryContext(ctx, coursesQuery, ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query courses: %w", err)
	}
	defer rows.Close()

	var courseIDs []any
	courseIdxMap := make(map[int64]int)

	for rows.Next() {
		var c academic.CourseSummaryView
		var courseID int64
		var courseType int
		var satDates, pres, m1, m2 string

		if err := rows.Scan(
			&courseID,
			&c.Section,
			&c.Shift,
			&c.Name,
			&courseType,
			&satDates,
			&pres,
			&m1,
			&m2,
		); err != nil {
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}

		c.ID = academic.CourseID(courseID)
		c.Type = academic.CourseType(courseType)
		c.SaturdayDates = satDates
		c.Committee = academic.Committee{
			President: pres,
			Member1:   m1,
			Member2:   m2,
		}
		c.Teachers = []academic.Teacher{}
		c.Schedules = []academic.ClassSession{}
		c.Exams = []academic.Exam{}

		courseIdxMap[courseID] = len(sch.Courses)
		sch.Courses = append(sch.Courses, c)
		courseIDs = append(courseIDs, courseID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(sch.Courses) == 0 {
		return &sch, nil
	}

	placeholders := strings.Repeat("?,", len(courseIDs)-1) + "?"

	teachersQuery := fmt.Sprintf(`
		SELECT 
			dc.id_curso,
			d.id,
			d.titulo,
			d.nombre,
			d.apellido,
			d.correo
		FROM docentes_curso dc
		JOIN docentes d ON dc.id_docente = d.id
		WHERE dc.id_curso IN (%s)
		ORDER BY dc.id_curso ASC`, placeholders)

	tRows, err := s.db.QueryContext(ctx, teachersQuery, courseIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query teachers: %w", err)
	}
	defer tRows.Close()

	for tRows.Next() {
		var courseID int64
		var teacher academic.Teacher
		var teacherID int64
		var title sql.NullString

		if err := tRows.Scan(&courseID, &teacherID, &title, &teacher.FirstName, &teacher.LastName, &teacher.Email); err != nil {
			return nil, fmt.Errorf("failed to scan teacher: %w", err)
		}

		teacher.ID = academic.TeacherID(teacherID)
		if title.Valid {
			teacher.Title = title.String
		}

		if idx, ok := courseIdxMap[courseID]; ok {
			sch.Courses[idx].Teachers = append(sch.Courses[idx].Teachers, teacher)
		}
	}

	schedulesQuery := fmt.Sprintf(`
		SELECT 
			curso_id,
			dia,
			CAST(desde AS TEXT),
			CAST(hasta AS TEXT),
			COALESCE(aula, '')
		FROM curso_horarios
		WHERE curso_id IN (%s)
		ORDER BY curso_id ASC`, placeholders)

	sRows, err := s.db.QueryContext(ctx, schedulesQuery, courseIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query class schedules: %w", err)
	}
	defer sRows.Close()

	for sRows.Next() {
		var courseID int64
		var day int
		var startTimeStr, endTimeStr, room string

		if err := sRows.Scan(&courseID, &day, &startTimeStr, &endTimeStr, &room); err != nil {
			return nil, fmt.Errorf("failed to scan class schedule: %w", err)
		}

		session := academic.ClassSession{
			Day:  academic.WeekDay(day),
			Room: room,
			Time: academic.TimeSlot{
				Start: parseTimeOnly(startTimeStr),
				End:   parseTimeOnly(endTimeStr),
			},
		}

		if idx, ok := courseIdxMap[courseID]; ok {
			sch.Courses[idx].Schedules = append(sch.Courses[idx].Schedules, session)
		}
	}

	examsQuery := fmt.Sprintf(`
		SELECT 
			curso_id,
			tipo,
			instancia,
			COALESCE(CAST(fecha AS TEXT), ''),
			COALESCE(CAST(hora AS TEXT), ''),
			COALESCE(aula, ''),
			COALESCE(CAST(revision_fecha AS TEXT), ''),
			COALESCE(CAST(revision_hora AS TEXT), '')
		FROM examenes
		WHERE curso_id IN (%s)
		ORDER BY curso_id ASC`, placeholders)

	eRows, err := s.db.QueryContext(ctx, examsQuery, courseIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query exams: %w", err)
	}
	defer eRows.Close()

	for eRows.Next() {
		var courseID int64
		var examTypeStr string
		var instance int
		var examDateStr, examTimeStr, room string
		var revDateStr, revTimeStr string

		if err := eRows.Scan(
			&courseID,
			&examTypeStr,
			&instance,
			&examDateStr,
			&examTimeStr,
			&room,
			&revDateStr,
			&revTimeStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan exam: %w", err)
		}

		exam := academic.Exam{
			Room:     room,
			Type:     academic.ExamType(examTypeStr),
			Instance: academic.ExamInstance(instance),
		}
		exam.SetDate(parseExamDateTime(examDateStr, examTimeStr))
		exam.SetRevision(parseExamDateTime(revDateStr, revTimeStr))

		if idx, ok := courseIdxMap[courseID]; ok {
			sch.Courses[idx].Exams = append(sch.Courses[idx].Exams, exam)
		}
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

// ==================
//  Helper functions
// ==================

func parseExamDateTime(dateStr, timeStr string) *time.Time {
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)

	// Requiere al menos los 10 caracteres de la fecha "YYYY-MM-DD"
	if len(dateStr) < 10 {
		return nil
	}

	// Extrae la parte YYYY-MM-DD ignorando la 'T' e ISO strings si existieran
	cleanDate := dateStr[:10]
	fullStr := cleanDate
	layout := "2006-01-02"

	// Si se cuenta con una hora válida (ej: "18:30" o "18:30:00")
	if len(timeStr) >= 5 {
		fullStr += " " + timeStr[:5]
		layout += " 15:04"
	}

	t, err := time.Parse(layout, fullStr)
	if err != nil {
		return nil
	}
	return &t
}

func parseTimeOnly(timeStr string) *time.Time {
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return nil
	}

	// Intenta parsear los formatos de hora habituales
	formats := []string{"15:04", "15:04:05"}
	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, timeStr); err == nil {
			return &t
		}
	}
	return nil
}
