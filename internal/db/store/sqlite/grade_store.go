package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteGradeStore struct {
	db *sql.DB
}

func NewSqliteGradeStore(db *sql.DB) *SqliteGradeStore {
	return &SqliteGradeStore{
		db: db,
	}
}

// ==========================================================
// =                     PUBLIC API                         =
// ==========================================================

func (s *SqliteGradeStore) FindById(ctx context.Context, id int64) (*model.GradeModel, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT 
			c.id, c.nombre, c.seccion,
			p.year, p.periodo,
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
			c.comite_presidente, c.comite_miembro1, c.comite_miembro2,
			m.semestre,
			a.nombre AS subject_name,
			d.siglas AS department_siglas,
			ca.siglas AS career_siglas
		FROM cursos c
		JOIN periodos p ON c.periodo = p.id
		JOIN mallas m ON c.malla = m.id
		JOIN asignaturas a ON m.asignatura = a.id
		JOIN departamentos d ON a.departamento = d.id
		JOIN carreras ca ON m.carrera = ca.id
		WHERE c.id = ?
	`, id)

	gm, err := scanGradeModel(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Cargar docentes aparte
	err = s.loadTeachersForGrade(ctx, id, gm)
	if err != nil {
		return nil, err
	}

	return gm, nil
}

func (s *SqliteGradeStore) ListByCareerAndPeriod(ctx context.Context, careerID int64, periodID int64) ([]*model.GradeListItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			c.id, c.nombre, c.seccion,
			m.semestre,
			a.nombre AS subject_name,
			GROUP_CONCAT(d.nombre) AS teachers
		FROM cursos c
		JOIN mallas m ON c.malla = m.id
		JOIN asignaturas a ON m.asignatura = a.id
		JOIN carreras ca ON m.carrera = ca.id
		LEFT JOIN docentes_curso dc ON c.id = dc.id_curso
		LEFT JOIN docentes d ON dc.id_docente = d.id
		WHERE ca.id = ? AND c.periodo = ?
		GROUP BY c.id
		ORDER BY m.semestre, c.seccion
	`, careerID, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.GradeListItem
	for rows.Next() {
		item := &model.GradeListItem{}
		var teachersStr string

		err := rows.Scan(
			&item.ID,
			&item.SubjectName,
			&item.Section,
			&item.Semester,
			&teachersStr,
		)
		if err != nil {
			return nil, err
		}

		if teachersStr != "" {
			item.TeacherName = strings.Split(teachersStr, ",")[0]
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *SqliteGradeStore) Upsert(
	ctx context.Context,
	insertFn func(persist func(model.GradeModel) error) error,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	persist := func(grade model.GradeModel) error {
		periodID, err := s.upsertPeriod(tx, ctx, grade.Period)
		if err != nil {
			return err
		}

		teacherIDs, err := s.upsertTeachers(tx, ctx, grade.Teachers)
		if err != nil {
			return err
		}

		mallaID, err := s.upsertCurriculum(tx, ctx, grade.Curriculum)
		if err != nil {
			return err
		}

		courseID, err := s.upsertCourse(tx, ctx, grade, mallaID, periodID)
		if err != nil {
			return err
		}

		return s.linkTeachersToCourse(tx, ctx, courseID, teacherIDs)
	}

	if err := insertFn(persist); err != nil {
		return fmt.Errorf("insert callback failed: %w", err)
	}

	return tx.Commit()
}

// ==========================================================
// =                  Private Methods                       =
// ==========================================================

// saves or updates the period and returns its ID
func (s *SqliteGradeStore) upsertPeriod(tx *sql.Tx, ctx context.Context, p model.Period) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
        INSERT INTO periodos (year, periodo)
        VALUES (?, ?)
        ON CONFLICT(year, periodo) DO UPDATE SET year = excluded.year
        RETURNING id
    `, p.Year, p.Period).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert period: %w", err)
	}
	return id, nil
}

// saves or updates all teachers and returns their IDs
func (s *SqliteGradeStore) upsertTeachers(tx *sql.Tx, ctx context.Context, teachers []model.Teacher) ([]int64, error) {
	ids := make([]int64, 0, len(teachers))

	for _, t := range teachers {
		id, err := s.upsertSingleTeacher(tx, ctx, t)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (s *SqliteGradeStore) upsertSingleTeacher(tx *sql.Tx, ctx context.Context, teacher model.Teacher) (int64, error) {
	// Try by email if present
	if teacher.Email != "" {
		var id int64
		err := tx.QueryRowContext(ctx, `
            INSERT INTO docentes (nombre, correo, search_key)
            VALUES (?, ?, ?)
            ON CONFLICT(correo) DO UPDATE SET
                nombre = excluded.nombre,
                search_key = excluded.search_key
            RETURNING id
        `, teacher.Name, teacher.Email, teacher.GetSearchKey()).Scan(&id)

		if err == nil {
			return id, nil
		}
	}

	// Search candidates by search_key
	rows, err := tx.QueryContext(ctx, `
        SELECT id, nombre, correo
        FROM docentes
        WHERE search_key = ?
    `, teacher.GetSearchKey())
	if err != nil {
		return 0, fmt.Errorf("search by search_key: %w", err)
	}
	defer rows.Close()

	candidates := []struct {
		ID          int64
		Name, Email string
	}{}
	for rows.Next() {
		var c struct {
			ID          int64
			Name, Email string
		}
		if err := rows.Scan(&c.ID, &c.Name, &c.Email); err != nil {
			return 0, fmt.Errorf("scan candidate: %w", err)
		}
		candidates = append(candidates, c)
	}

	// Try to match using the model's comparison method
	for _, c := range candidates {
		if teacher.IsSimilar(c.Name) {
			// Update if better data is available
			_, err = tx.ExecContext(ctx, `
                UPDATE docentes SET
                    nombre = COALESCE(NULLIF(?, ''), nombre),
                    correo = COALESCE(NULLIF(?, ''), correo)
                WHERE id = ?
            `, teacher.Name, teacher.Email, c.ID)
			if err != nil {
				return 0, fmt.Errorf("update matched teacher: %w", err)
			}
			return c.ID, nil
		}
	}

	// No match: insert new
	var newID int64
	err = tx.QueryRowContext(ctx, `
        INSERT INTO docentes (nombre, correo, search_key)
        VALUES (?, ?, ?)
        RETURNING id
    `, teacher.Name, teacher.Email, teacher.GetSearchKey()).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("insert new teacher: %w", err)
	}

	return newID, nil
}

func (s *SqliteGradeStore) upsertCurriculum(tx *sql.Tx, ctx context.Context, c model.Curriculum) (int64, error) {
	deptID, err := s.upsertDepartment(tx, ctx, c.Subject.Department)
	if err != nil {
		return 0, err
	}

	subjectID, err := s.upsertSubject(tx, ctx, c.Subject.Name, deptID)
	if err != nil {
		return 0, err
	}

	careerID, err := s.upsertCareer(tx, ctx, c.Career)
	if err != nil {
		return 0, err
	}

	var mallaID int64
	err = tx.QueryRowContext(ctx, `
        INSERT INTO mallas (carrera, asignatura, semestre)
        VALUES (?, ?, ?)
        ON CONFLICT(carrera, asignatura) DO UPDATE SET semestre = excluded.semestre
        RETURNING id
    `, careerID, subjectID, c.Semester).Scan(&mallaID)
	if err != nil {
		return 0, fmt.Errorf("upsert malla: %w", err)
	}

	return mallaID, nil
}

func (s *SqliteGradeStore) upsertDepartment(tx *sql.Tx, ctx context.Context, siglas string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
        INSERT INTO departamentos (siglas)
        VALUES (?)
        ON CONFLICT(siglas) DO UPDATE SET siglas = excluded.siglas
        RETURNING id
    `, siglas).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert department: %w", err)
	}
	return id, nil
}

func (s *SqliteGradeStore) upsertSubject(tx *sql.Tx, ctx context.Context, name string, deptID int64) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
        INSERT INTO asignaturas (nombre, departamento)
        VALUES (?, ?)
        ON CONFLICT(nombre) DO UPDATE SET departamento = excluded.departamento
        RETURNING id
    `, name, deptID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert subject: %w", err)
	}
	return id, nil
}

func (s *SqliteGradeStore) upsertCareer(tx *sql.Tx, ctx context.Context, siglas string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
        INSERT INTO carreras (siglas)
        VALUES (?)
        ON CONFLICT(siglas) DO UPDATE SET siglas = excluded.siglas
        RETURNING id
    `, siglas).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert career: %w", err)
	}
	return id, nil
}

// upsertCourse saves the course itself and returns its ID
func (s *SqliteGradeStore) upsertCourse(tx *sql.Tx, ctx context.Context, grade model.GradeModel, mallaID, periodID int64) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
        INSERT INTO cursos (
            malla, periodo, nombre, seccion, solo_examen_final,
            lunes_desde, lunes_hasta, lunes_aula,
            martes_desde, martes_hasta, martes_aula,
            miercoles_desde, miercoles_hasta, miercoles_aula,
            jueves_desde, jueves_hasta, jueves_aula,
            viernes_desde, viernes_hasta, viernes_aula,
            sabado_desde, sabado_hasta, sabado_aula, sabado_night_fechas,
            partial1_fecha, partial1_hora, partial1_aula,
            partial2_fecha, partial2_hora, partial2_aula,
            final1_fecha, final1_hora, final1_aula, final1_fecha_revision, final1_hora_revision,
            final2_fecha, final2_hora, final2_aula, final2_fecha_revision, final2_hora_revision,
            comite_presidente, comite_miembro1, comite_miembro2
        ) VALUES (
            ?, ?, ?, ?, 1,
            ?, ?, ?,
            ?, ?, ?,
            ?, ?, ?,
            ?, ?, ?,
            ?, ?, ?,
            ?, ?, ?, ?,
            ?, ?, ?,
            ?, ?, ?,
            ?, ?, ?, ?, ?,
            ?, ?, ?, ?, ?,
            ?, ?, ?
        ) ON CONFLICT(malla, seccion, periodo) DO UPDATE SET
            nombre = excluded.nombre,
            solo_examen_final = excluded.solo_examen_final,
            lunes_desde = excluded.lunes_desde,
            lunes_hasta = excluded.lunes_hasta,
            lunes_aula = excluded.lunes_aula,
            martes_desde = excluded.martes_desde,
            martes_hasta = excluded.martes_hasta,
            martes_aula = excluded.martes_aula,
            miercoles_desde = excluded.miercoles_desde,
            miercoles_hasta = excluded.miercoles_hasta,
            miercoles_aula = excluded.miercoles_aula,
            jueves_desde = excluded.jueves_desde,
            jueves_hasta = excluded.jueves_hasta,
            jueves_aula = excluded.jueves_aula,
            viernes_desde = excluded.viernes_desde,
            viernes_hasta = excluded.viernes_hasta,
            viernes_aula = excluded.viernes_aula,
            sabado_desde = excluded.sabado_desde,
            sabado_hasta = excluded.sabado_hasta,
            sabado_aula = excluded.sabado_aula,
            sabado_night_fechas = excluded.sabado_night_fechas,
            partial1_fecha = excluded.partial1_fecha,
            partial1_hora = excluded.partial1_hora,
            partial1_aula = excluded.partial1_aula,
            partial2_fecha = excluded.partial2_fecha,
            partial2_hora = excluded.partial2_hora,
            partial2_aula = excluded.partial2_aula,
            final1_fecha = excluded.final1_fecha,
            final1_hora = excluded.final1_hora,
            final1_aula = excluded.final1_aula,
            final1_fecha_revision = excluded.final1_fecha_revision,
            final1_hora_revision = excluded.final1_hora_revision,
            final2_fecha = excluded.final2_fecha,
            final2_hora = excluded.final2_hora,
            final2_aula = excluded.final2_aula,
            final2_fecha_revision = excluded.final2_fecha_revision,
            final2_hora_revision = excluded.final2_hora_revision,
            comite_presidente = excluded.comite_presidente,
            comite_miembro1 = excluded.comite_miembro1,
            comite_miembro2 = excluded.comite_miembro2
        RETURNING id
    `,
		mallaID, periodID, grade.Name, grade.Section,
		grade.Monday.Start, grade.Monday.End, grade.MondayRoom,
		grade.Tuesday.Start, grade.Tuesday.End, grade.TuesdayRoom,
		grade.Wednesday.Start, grade.Wednesday.End, grade.WednesdayRoom,
		grade.Thursday.Start, grade.Thursday.End, grade.ThursdayRoom,
		grade.Friday.Start, grade.Friday.End, grade.FridayRoom,
		grade.Saturday.Start, grade.Saturday.End, grade.SaturdayRoom, grade.SaturdayDates,
		grade.Partial1Date, grade.Partial1Time, grade.Partial1Room,
		grade.Partial2Date, grade.Partial2Time, grade.Partial2Room,
		grade.Final1Date, grade.Final1Time, grade.Final1Room, grade.Final1RevDate, grade.Final1RevTime,
		grade.Final2Date, grade.Final2Time, grade.Final2Room, grade.Final2RevDate, grade.Final2RevTime,
		grade.CommitteePresident, grade.CommitteeMember1, grade.CommitteeMember2,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("upsert course: %w", err)
	}

	return id, nil
}

// removes old teachers links and adds new ones
func (s *SqliteGradeStore) linkTeachersToCourse(tx *sql.Tx, ctx context.Context, courseID int64, teacherIDs []int64) error {
	_, err := tx.ExecContext(ctx, `
        DELETE FROM docentes_curso WHERE id_curso = ?
    `, courseID)
	if err != nil {
		return fmt.Errorf("delete old teacher links: %w", err)
	}

	for _, tid := range teacherIDs {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO docentes_curso (id_docente, id_curso)
            VALUES (?, ?)
        `, tid, courseID)
		if err != nil {
			return fmt.Errorf("link teacher %d to course: %w", tid, err)
		}
	}

	return nil
}

// scanGradeModel carga todos los campos del row a GradeModel
func scanGradeModel(row *sql.Row) (*model.GradeModel, error) {
	gm := &model.GradeModel{}

	err := row.Scan(
		&gm.ID,
		&gm.Name,
		&gm.Section,
		&gm.Period.Year,
		&gm.Period.Period,
		&gm.Monday.Start,
		&gm.Monday.End,
		&gm.MondayRoom,
		&gm.Tuesday.Start,
		&gm.Tuesday.End,
		&gm.TuesdayRoom,
		&gm.Wednesday.Start,
		&gm.Wednesday.End,
		&gm.WednesdayRoom,
		&gm.Thursday.Start,
		&gm.Thursday.End,
		&gm.ThursdayRoom,
		&gm.Friday.Start,
		&gm.Friday.End,
		&gm.FridayRoom,
		&gm.Saturday.Start,
		&gm.Saturday.End,
		&gm.SaturdayRoom,
		&gm.SaturdayDates,
		&gm.Partial1Date,
		&gm.Partial1Time,
		&gm.Partial1Room,
		&gm.Partial2Date,
		&gm.Partial2Time,
		&gm.Partial2Room,
		&gm.Final1Date,
		&gm.Final1Time,
		&gm.Final1Room,
		&gm.Final1RevDate,
		&gm.Final1RevTime,
		&gm.Final2Date,
		&gm.Final2Time,
		&gm.Final2Room,
		&gm.Final2RevDate,
		&gm.Final2RevTime,
		&gm.CommitteePresident,
		&gm.CommitteeMember1,
		&gm.CommitteeMember2,
		&gm.Curriculum.Semester,
		&gm.Curriculum.Subject.Name,
		&gm.Curriculum.Subject.Department,
		&gm.Curriculum.Career,
	)
	if err != nil {
		return nil, err
	}

	return gm, nil
}

// loadTeachersForGrade carga los nombres de docentes para un curso
func (s *SqliteGradeStore) loadTeachersForGrade(ctx context.Context, courseID int64, gm *model.GradeModel) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.nombre
		FROM docentes_curso dc
		JOIN docentes d ON dc.id_docente = d.id
		WHERE dc.id_curso = ?
		`, courseID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	gm.Teachers = make([]model.Teacher, len(names))
	for i, name := range names {
		gm.Teachers[i].Name = name
	}

	return nil
}
