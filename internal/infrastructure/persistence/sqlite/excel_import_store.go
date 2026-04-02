package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	excelimport "github.com/elias-gill/poliplanner2/internal/app/excelImport"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/domain/sheetVersion"
)

// ============================================================
// Writer
// ============================================================

type sqliteExcelImportWritter struct {
	tx *sql.Tx
}

// ============================================================
// Public Storer
// ============================================================

type SqliteExcelImportStorer struct {
	db *sql.DB
}

func NewSqliteExcelImportStorer(db *sql.DB) *SqliteExcelImportStorer {
	return &SqliteExcelImportStorer{
		db: db,
	}
}

func (s *SqliteExcelImportStorer) RunImport(
	ctx context.Context,
	fn func(excelimport.ImportWriter) error,
) (err error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			_ = tx.Rollback()
			return
		}

		if cErr := tx.Commit(); cErr != nil {
			err = cErr
		}
	}()

	err = fn(sqliteExcelImportWritter{tx})
	return
}

func (s *SqliteExcelImportStorer) SaveAudit(
	ctx context.Context,
	version *sheetVersion.SheetVersion,
) (sheetVersion.SheetVersionID, error) {

	successInt := 1
	if version.Error != "" {
		successInt = 0
	}

	result, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO sheet_version (
			file_name,
			url,
			success,
			error_message,
			period,
			parsed_at,
			parsed_sheets
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		version.FileName,
		version.URL,
		successInt,
		version.Error,
		version.Period,
		version.ParsedAt,
		version.ParsedSheets,
	)

	if err != nil {
		return -1, fmt.Errorf("failed to insert sheet audit record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to get last inserted record id: %w", err)
	}

	return sheetVersion.SheetVersionID(id), nil
}

// ============================================================
// Writer Interface Implementation
// ============================================================

func (w sqliteExcelImportWritter) EnsurePeriod(p period.Period) (period.PeriodID, error) {

	_, err := w.tx.Exec(`
		INSERT INTO periodos (year, periodo)
		VALUES (?, ?)
		ON CONFLICT(year, periodo) DO NOTHING
	`, p.Year, p.Semester)
	if err != nil {
		return 0, fmt.Errorf("failed to insert period: %w", err)
	}

	var id int64
	err = w.tx.QueryRow(`
		SELECT id FROM periodos
		WHERE year = ? AND periodo = ?
	`, p.Year, p.Semester).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve period id: %w", err)
	}

	return period.PeriodID(id), nil
}

func (w sqliteExcelImportWritter) SaveCourseOffering(off excelimport.Offering) error {

	// 1. Carrera
	careerID, err := w.ensureCareer(off.Subject.Career)
	if err != nil {
		return err
	}

	// 2. Departamento
	deptID, err := w.ensureDepartment(off.Subject.Department)
	if err != nil {
		return err
	}

	// 3. Asignatura
	subjectID, err := w.ensureSubject(off.Subject.Name, deptID)
	if err != nil {
		return err
	}

	// 4. Malla
	mallaID, err := w.ensureMalla(careerID, subjectID, off.Subject.Semester, off.Subject.Level)
	if err != nil {
		return err
	}

	// 5. Curso
	courseID, err := w.upsertCourse(off, mallaID)
	if err != nil {
		return err
	}

	// 6. Teachers
	teacherIDs, err := w.ensureTeachers(off)
	if err != nil {
		return err
	}

	if err := w.replaceCourseTeachers(courseID, teacherIDs); err != nil {
		return err
	}

	return nil
}

// ============================================================
// Core Upserts
// ============================================================

func (w sqliteExcelImportWritter) ensureCareer(siglas string) (int64, error) {
	_, err := w.tx.Exec(`
		INSERT INTO carreras (siglas)
		VALUES (?)
		ON CONFLICT(siglas) DO NOTHING
	`, siglas)
	if err != nil {
		return 0, err
	}

	var id int64
	err = w.tx.QueryRow(`SELECT id FROM carreras WHERE siglas = ?`, siglas).Scan(&id)
	return id, err
}

func (w sqliteExcelImportWritter) ensureDepartment(siglas string) (int64, error) {
	_, err := w.tx.Exec(`
		INSERT INTO departamentos (siglas)
		VALUES (?)
		ON CONFLICT(siglas) DO NOTHING
	`, siglas)
	if err != nil {
		return 0, err
	}

	var id int64
	err = w.tx.QueryRow(`SELECT id FROM departamentos WHERE siglas = ?`, siglas).Scan(&id)
	return id, err
}

func (w sqliteExcelImportWritter) ensureSubject(name string, deptID int64) (int64, error) {
	_, err := w.tx.Exec(`
		INSERT INTO asignaturas (nombre, departamento)
		VALUES (?, ?)
		ON CONFLICT(nombre) DO NOTHING
	`, name, deptID)
	if err != nil {
		return 0, err
	}

	var id int64
	err = w.tx.QueryRow(`SELECT id FROM asignaturas WHERE nombre = ?`, name).Scan(&id)
	return id, err
}

func (w sqliteExcelImportWritter) ensureMalla(
	careerID int64,
	subjectID int64,
	semester int,
	level int,
) (int64, error) {

	_, err := w.tx.Exec(`
		INSERT INTO mallas (carrera, asignatura, semestre, nivel)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(carrera, asignatura) DO UPDATE SET
			semestre = excluded.semestre,
			nivel = excluded.nivel
	`, careerID, subjectID, semester, level)
	if err != nil {
		return 0, err
	}

	var id int64
	err = w.tx.QueryRow(`
		SELECT id FROM mallas
		WHERE carrera = ? AND asignatura = ?
	`, careerID, subjectID).Scan(&id)

	return id, err
}

// ============================================================
// Course
// ============================================================

func (w sqliteExcelImportWritter) upsertCourse(
	off excelimport.Offering,
	mallaID int64,
) (int64, error) {

	var courseID int64

	err := w.tx.QueryRow(`
		INSERT INTO cursos (
			malla, periodo, nombre, seccion, tipo,
			comite_presidente, comite_miembro1, comite_miembro2, fechas_sabados
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(malla, seccion, periodo) DO UPDATE SET
			nombre = excluded.nombre,
			tipo = excluded.tipo,
			comite_presidente = excluded.comite_presidente,
			comite_miembro1 = excluded.comite_miembro1,
			comite_miembro2 = excluded.comite_miembro2,
			fechas_sabados = excluded.fechas_sabados
		RETURNING id
	`,
		mallaID,
		off.Period,
		off.CourseName,
		off.Section,
		off.CourseType,
		off.CommitteePresident,
		off.CommitteeMember1,
		off.CommitteeMember2,
		off.SaturdayDates,
	).Scan(&courseID)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert course: %w", err)
	}

	if err := w.replaceSchedule(courseID, off.Schedule); err != nil {
		return 0, err
	}

	if err := w.replaceExams(courseID, off.Exams); err != nil {
		return 0, err
	}

	return courseID, nil
}

// ============================================================
// Schedule
// ============================================================

func (w sqliteExcelImportWritter) replaceSchedule(
	courseID int64,
	schedule []excelimport.ScheduleEntry,
) error {

	_, err := w.tx.Exec(`DELETE FROM curso_horarios WHERE curso_id = ?`, courseID)
	if err != nil {
		return fmt.Errorf("failed to clear schedule: %w", err)
	}

	stmt, err := w.tx.Prepare(`
		INSERT INTO curso_horarios (curso_id, dia, desde, hasta, aula)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare schedule insert: %w", err)
	}
	defer stmt.Close()

	for _, s := range schedule {
		if s.Start == nil || s.End == nil {
			continue
		}

		_, err := stmt.Exec(
			courseID,
			int(s.Day),
			s.Start.Format("15:04"),
			s.End.Format("15:04"),
			s.Room,
		)
		if err != nil {
			return fmt.Errorf("failed to insert schedule: %w", err)
		}
	}

	return nil
}

// ============================================================
// Exams
// ============================================================

func (w sqliteExcelImportWritter) replaceExams(
	courseID int64,
	exams []excelimport.Exam,
) error {

	_, err := w.tx.Exec(`DELETE FROM examenes WHERE curso_id = ?`, courseID)
	if err != nil {
		return fmt.Errorf("failed to clear exams: %w", err)
	}

	stmt, err := w.tx.Prepare(`
		INSERT INTO examenes (
			curso_id, tipo, instancia,
			fecha, hora, aula,
			revision_fecha, revision_hora
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare exam insert: %w", err)
	}
	defer stmt.Close()

	for _, e := range exams {
		if e.Date == nil {
			continue
		}

		var revDate any
		var revTime any

		if e.Revision != nil {
			revDate = e.Revision.Format("2006-01-02")
			revTime = e.Revision.Format("15:04")
		}

		_, err := stmt.Exec(
			courseID,
			examTypeToDB(e.Type),
			e.Instance,
			e.Date.Format("2006-01-02"),
			e.Date.Format("15:04"),
			e.Room,
			revDate,
			revTime,
		)
		if err != nil {
			return fmt.Errorf("failed to insert exam: %w", err)
		}
	}

	return nil
}

// ============================================================
// Teachers
// ============================================================

func (w sqliteExcelImportWritter) ensureTeachers(
	off excelimport.Offering,
) ([]int64, error) {

	var ids []int64

	for _, t := range off.Teachers {

		var id int64
		err := w.tx.QueryRow(`
			SELECT id FROM docentes WHERE correo = ?
		`, t.Email).Scan(&id)

		if err == nil {
			ids = append(ids, id)
			continue
		}

		if err != sql.ErrNoRows {
			return nil, err
		}

		rows, err := w.tx.Query(`
			SELECT id, nombre, apellido, correo 
			FROM docentes WHERE search_key = ?
		`, t.SearchKey)
		if err != nil {
			return nil, err
		}

		var matchedID int64
		found := false

		for rows.Next() {
			var cid int64
			var fn, ln, email string

			if err := rows.Scan(&cid, &fn, &ln, &email); err != nil {
				rows.Close()
				return nil, err
			}

			if t.IsSimilar(fn + " " + ln) {
				matchedID = cid
				found = true
				break
			}
		}
		rows.Close()

		if found {
			_, err := w.tx.Exec(`
				UPDATE docentes
				SET nombre = ?, apellido = ?, correo = ?, search_key = ?
				WHERE id = ?
			`, t.FirstName, t.LastName, t.Email, t.SearchKey, matchedID)
			if err != nil {
				return nil, err
			}

			ids = append(ids, matchedID)
			continue
		}

		var newID int64
		err = w.tx.QueryRow(`
			INSERT INTO docentes (nombre, apellido, correo, search_key)
			VALUES (?, ?, ?, ?)
			RETURNING id
		`, t.FirstName, t.LastName, t.Email, t.SearchKey).Scan(&newID)
		if err != nil {
			return nil, err
		}

		ids = append(ids, newID)
	}

	return ids, nil
}

func (w sqliteExcelImportWritter) replaceCourseTeachers(
	courseID int64,
	teacherIDs []int64,
) error {

	_, err := w.tx.Exec(`DELETE FROM docentes_curso WHERE id_curso = ?`, courseID)
	if err != nil {
		return err
	}

	stmt, err := w.tx.Prepare(`
		INSERT INTO docentes_curso (id_docente, id_curso)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tid := range teacherIDs {
		if _, err := stmt.Exec(tid, courseID); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================
// Utils
// ============================================================

func examTypeToDB(t courseOffering.ExamType) string {
	switch t {
	case courseOffering.ExamPartial:
		return "partial"
	case courseOffering.ExamFinal:
		return "final"
	default:
		return "unknown"
	}
}
