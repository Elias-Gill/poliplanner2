package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	excelimport "github.com/elias-gill/poliplanner2/internal/app/excelImport"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/period"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
)

func NewSqliteExcelImportStorer(db *sql.DB) *SqliteExcelImportStorer {
	return &SqliteExcelImportStorer{
		db: db,
	}
}

type SqliteExcelImportStorer struct {
	db *sql.DB
}

type sqliteExcelImportWritter struct {
	tx *sql.Tx
}

func (s *SqliteExcelImportStorer) RunImport(
	ctx context.Context,
	fn func(excelimport.ImportWriter) error,
) (err error) {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure rollback on panic or early return
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		// If err is set, rollback the transaction
		if err != nil {
			_ = tx.Rollback()
			return
		}

		// Otherwise attempt to commit
		if cErr := tx.Commit(); cErr != nil {
			err = cErr
		}
	}()

	// Execute transactional logic
	err = fn(sqliteExcelImportWritter{tx})
	return
}

func (s *SqliteExcelImportStorer) SaveAudit(
	ctx context.Context,
	meta source.ExcelSourceMetadata,
	success bool,
	errorMsg error,
) error {
	// Convert bool to SQLite-compatible integer (0/1)
	successInt := 0
	if success {
		successInt = 1
	}

	// Extract error message if present
	var errStr *string
	if errorMsg != nil {
		msg := errorMsg.Error()
		errStr = &msg
	}

	// Insert audit record
	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO sheet_version (
			file_name,
			url,
			success,
			error_message
		) VALUES (?, ?, ?, ?)
		`,
		meta.Name,
		meta.URI,
		successInt,
		errStr,
	)
	if err != nil {
		return fmt.Errorf("failed to insert sheet audit record: %w", err)
	}

	return nil
}

func (w sqliteExcelImportWritter) EnsurePeriod(p period.Period) (period.PeriodID, error) {
	// Try to insert the period (no-op if it already exists due to UNIQUE constraint)
	_, err := w.tx.Exec(
		`
		INSERT INTO periodos (year, periodo)
		VALUES (?, ?)
		ON CONFLICT(year, periodo) DO NOTHING
		`,
		p.Year,
		p.Semester,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert period: %w", err)
	}

	// Retrieve the period ID (either newly inserted or existing)
	var id int64
	err = w.tx.QueryRow(
		`
		SELECT id FROM periodos
		WHERE year = ? AND periodo = ?
		`,
		p.Year,
		p.Semester,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to retrieve period id: %w", err)
	}

	return period.PeriodID(id), nil
}

func (w sqliteExcelImportWritter) SaveCourseOffering(off excelimport.Offering) error {
	// SUBJECT
	mallaID, err := w.ensureSubject(off)
	if err != nil {
		return err
	}

	// TEACHERS
	teacherIDs, err := w.ensureTeachers(off)
	if err != nil {
		return err
	}

	// COURSE
	courseID, err := w.upsertCourse(off, mallaID)
	if err != nil {
		return err
	}

	// LINK TEACHERS
	err = w.replaceCourseTeachers(courseID, teacherIDs)
	if err != nil {
		return err
	}

	return nil
}

// =========================
// Private methods
// =========================

func (w sqliteExcelImportWritter) ensureSubject(s excelimport.Offering) (int64, error) {
	// =========================
	// Carrera
	// =========================
	var careerID int64
	err := w.tx.QueryRow(`
		INSERT INTO carreras (siglas)
		VALUES (?)
		ON CONFLICT(siglas) DO UPDATE SET siglas = excluded.siglas
		RETURNING id
	`, s.Subject.Career).Scan(&careerID)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert carrera: %w", err)
	}

	// =========================
	// Departamento
	// =========================
	var deptID int64
	err = w.tx.QueryRow(`
		INSERT INTO departamentos (siglas)
		VALUES (?)
		ON CONFLICT(siglas) DO UPDATE SET siglas = excluded.siglas
		RETURNING id
	`, s.Subject.Department).Scan(&deptID)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert departamento: %w", err)
	}

	// =========================
	// Asignatura
	// =========================
	var subjectID int64
	err = w.tx.QueryRow(`
		INSERT INTO asignaturas (nombre, departamento)
		VALUES (?, ?)
		ON CONFLICT(nombre) DO UPDATE SET
			departamento = excluded.departamento
		RETURNING id
	`, s.Subject.Name, deptID).Scan(&subjectID)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert asignatura: %w", err)
	}

	// =========================
	// Malla
	// =========================
	var mallaID int64
	err = w.tx.QueryRow(`
		INSERT INTO mallas (carrera, asignatura, semestre, nivel)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(carrera, asignatura) DO UPDATE SET
			semestre = excluded.semestre,
			nivel = excluded.nivel
		RETURNING id
	`, careerID, subjectID, s.Subject.Semester, s.Subject.Level).Scan(&mallaID)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert malla: %w", err)
	}

	return mallaID, nil
}

func (w sqliteExcelImportWritter) ensureTeachers(off excelimport.Offering) ([]int64, error) {
	var ids []int64

	for _, t := range off.Teachers {
		searchKey := t.SearchKey

		// =========================
		//  Exact match by email
		// =========================
		var id int64
		err := w.tx.QueryRow(`
			SELECT id FROM docentes WHERE correo = ?
			`, t.Email).Scan(&id)

		if err == nil {
			ids = append(ids, id)
			continue
		}

		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to query teacher by email: %w", err)
		}

		// =========================
		//  Search by search_key
		// =========================
		rows, err := w.tx.Query(`
			SELECT id, nombre, correo FROM docentes WHERE search_key = ?
			`, searchKey)
		if err != nil {
			return nil, fmt.Errorf("failed to query teachers by search_key: %w", err)
		}

		var matchedID int64
		found := false

		for rows.Next() {
			var candidateID int64
			var name string
			var email string

			if err := rows.Scan(&candidateID, &name, &email); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan teacher candidate: %w", err)
			}

			if t.IsSimilar(name) {
				matchedID = candidateID
				found = true
				break
			}
		}
		rows.Close()

		// ============================
		// 3. If match found -> UPDATE
		// ============================
		if found {
			_, err := w.tx.Exec(`
				UPDATE docentes
				SET nombre = ?, correo = ?, search_key = ?
				WHERE id = ?
				`, t.Name, t.Email, searchKey, matchedID)
			if err != nil {
				return nil, fmt.Errorf("failed to update teacher: %w", err)
			}

			ids = append(ids, matchedID)
			continue
		}

		// =========================
		// 4. INSERT new teacher
		// =========================
		var newID int64
		err = w.tx.QueryRow(`
			INSERT INTO docentes (nombre, correo, search_key)
			VALUES (?, ?, ?)
			RETURNING id
			`, t.Name, t.Email, searchKey).Scan(&newID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert teacher: %w", err)
		}

		ids = append(ids, newID)
	}

	return ids, nil
}

func (w sqliteExcelImportWritter) upsertCourse(
	off excelimport.Offering,
	mallaID int64,
) (int64, error) {

	// =========================
	// Helpers
	// =========================

	formatDate := func(t *time.Time) any {
		if t == nil {
			return nil
		}
		return t.Format("2006-01-02")
	}

	formatTime := func(t *time.Time) any {
		if t == nil {
			return nil
		}
		return t.Format("15:04")
	}

	getDay := func(day courseOffering.WeekDay) (desde, hasta, aula any) {
		d, ok := off.Schedule[day]
		if !ok {
			return nil, nil, nil
		}
		return formatTime(d.Time.Start), formatTime(d.Time.End), d.Room
	}

	// =========================
	// Schedule
	// =========================

	lunD, lunH, lunA := getDay(courseOffering.Monday)
	marD, marH, marA := getDay(courseOffering.Tuesday)
	mieD, mieH, mieA := getDay(courseOffering.Wednesday)
	jueD, jueH, jueA := getDay(courseOffering.Thursday)
	vieD, vieH, vieA := getDay(courseOffering.Friday)
	sabD, sabH, sabA := getDay(courseOffering.Saturday)

	// =========================
	// Upsert
	// =========================

	var courseID int64
	err := w.tx.QueryRow(`
		INSERT INTO cursos (
		malla, periodo, nombre, seccion, tipo,

		lunes_desde, lunes_hasta, lunes_aula,
		martes_desde, martes_hasta, martes_aula,
		miercoles_desde, miercoles_hasta, miercoles_aula,
		jueves_desde, jueves_hasta, jueves_aula,
		viernes_desde, viernes_hasta, viernes_aula,
		sabado_desde, sabado_hasta, sabado_aula, sabado_night_fechas,

		partial1_fecha, partial1_hora, partial1_aula,
		partial2_fecha, partial2_hora, partial2_aula,

		final1_fecha, final1_hora, final1_aula,
		final1_fecha_revision, final1_hora_revision,

		final2_fecha, final2_hora, final2_aula,
		final2_fecha_revision, final2_hora_revision,

		comite_presidente, comite_miembro1, comite_miembro2
		)
		VALUES (?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
		ON CONFLICT(malla, seccion, periodo) DO UPDATE SET
		nombre = excluded.nombre,
		tipo = excluded.tipo,

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
		mallaID,
		off.Period,
		off.CourseName,
		off.Section,
		off.CourseType,

		lunD, lunH, lunA,
		marD, marH, marA,
		mieD, mieH, mieA,
		jueD, jueH, jueA,
		vieD, vieH, vieA,
		sabD, sabH, sabA, off.SaturdayDates,

		formatDate(off.Partial1.Date), formatTime(off.Partial1.Date), off.Partial1.Room,
		formatDate(off.Partial2.Date), formatTime(off.Partial2.Date), off.Partial2.Room,

		formatDate(off.Final1.Date), formatTime(off.Final1.Date), off.Final1.Room,
		formatDate(off.Final1.Revision), formatTime(off.Final1.Revision),

		formatDate(off.Final2.Date), formatTime(off.Final2.Date), off.Final2.Room,
		formatDate(off.Final2.Revision), formatTime(off.Final2.Revision),

		off.CommitteePresident,
		off.CommitteeMember1,
		off.CommitteeMember2,
	).Scan(&courseID)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert course: %w", err)
	}

	return courseID, nil
}

func (w sqliteExcelImportWritter) replaceCourseTeachers(courseID int64, teacherIDs []int64) error {
	// Delete all existing teachers for this course
	_, err := w.tx.Exec(`
		DELETE FROM docentes_curso
		WHERE id_curso = ?
		`, courseID)
	if err != nil {
		return fmt.Errorf("failed to delete existing course teachers: %w", err)
	}

	// Insert all new teacher links
	stmt, err := w.tx.Prepare(`
		INSERT INTO docentes_curso (id_docente, id_curso)
		VALUES (?, ?)
		`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert for course teachers: %w", err)
	}
	defer stmt.Close()

	for _, teacherID := range teacherIDs {
		if _, err := stmt.Exec(teacherID, courseID); err != nil {
			return fmt.Errorf("failed to insert teacher %d for course %d: %w", teacherID, courseID, err)
		}
	}

	return nil
}
