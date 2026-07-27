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

	// Get schedule details
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

	// Get courses details
	coursesQuery := `
		SELECT 
			c.id, 
			c.seccion, 
			c.turno, 
			c.nombre, 
			c.tipo
		FROM horarios_detalle hd
		JOIN cursos c ON hd.curso_id = c.id
		JOIN mallas m ON c.malla = m.id
		JOIN asignaturas a ON m.asignatura = a.id
		WHERE hd.horario_id = ?`

	rows, err := s.db.QueryContext(ctx, coursesQuery, ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query courses: %w", err)
	}
	defer rows.Close()

	courseMap := make(map[academic.CourseID]*academic.CourseSummaryView)
	var courseIDs []any

	for rows.Next() {
		var c academic.CourseSummaryView
		var courseID int64
		var courseType int

		if err := rows.Scan(&courseID, &c.Section, &c.Shift, &c.Name, &courseType); err != nil {
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}

		c.ID = academic.CourseID(courseID)
		c.Type = academic.CourseType(courseType)
		c.Teachers = []academic.Teacher{}
		c.Schedules = []academic.ClassSession{}

		sch.Courses = append(sch.Courses, c)
		courseIDs = append(courseIDs, courseID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty if schedule has no course associated
	if len(sch.Courses) == 0 {
		return &sch, nil
	}

	for i := range sch.Courses {
		courseMap[sch.Courses[i].ID] = &sch.Courses[i]
	}

	placeholders := strings.Repeat("?,", len(courseIDs)-1) + "?"

	// Load teachers
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
		WHERE dc.id_curso IN (%s)`, placeholders)

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

		if course, ok := courseMap[academic.CourseID(courseID)]; ok {
			course.Teachers = append(course.Teachers, teacher)
		}
	}

	// Load class sessions
	schedulesQuery := fmt.Sprintf(`
		SELECT 
			curso_id,
			dia,
			desde,
			hasta,
			COALESCE(aula, '')
		FROM curso_horarios
		WHERE curso_id IN (%s)
		ORDER BY dia ASC, desde ASC`, placeholders)

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

		startTime := parseTimeOnly(startTimeStr)
		endTime := parseTimeOnly(endTimeStr)

		session := academic.ClassSession{
			Day:  academic.WeekDay(day),
			Room: room,
			Time: academic.TimeSlot{
				Start: startTime,
				End:   endTime,
			},
		}

		if course, ok := courseMap[academic.CourseID(courseID)]; ok {
			course.Schedules = append(course.Schedules, session)
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

func parseTimeOnly(timeStr string) *time.Time {
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
