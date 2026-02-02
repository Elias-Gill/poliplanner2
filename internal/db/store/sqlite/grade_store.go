package sqlite

import (
	"context"
	"database/sql"
	"fmt"

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

func (s *SqliteGradeStore) FindById(ctx context.Context, id int64) (*model.GradeModel, error) {
	// TODO: implement
	return nil, nil
}

func (s *SqliteGradeStore) Upsert(
	ctx context.Context,
	insertFn func(persist func(model.GradeModel) error) error,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback() // automatic rollback if no commit

	// prepare the persist function
	persist := func(grade model.GradeModel) error {
		var err error
		var id int64

		// update or save the period data
		err = tx.QueryRowContext(ctx, `
            INSERT INTO periodos (year, periodo)
            VALUES (?, ?)
            ON CONFLICT(year, periodo) DO UPDATE SET year = excluded.year
            RETURNING id
        `, grade.Period.Year, grade.Period.Period).Scan(&id)
		if err != nil {
			return fmt.Errorf("upsert periodo: %w", err)
		}
		periodID := id

		// update or save all teachers data
		var teachersIDs []int64
		for _, teacher := range grade.Teachers {
			var teacherID int64

			// FIX: esta mal lo que esta haciendo aca
			err = tx.QueryRowContext(ctx, `
                INSERT INTO docentes (nombre, correo, search_key)
                VALUES (?, ?, ?)
                ON CONFLICT(correo) DO UPDATE SET nombre = excluded.nombre, search_key = excluded.search_key
                RETURNING id
            `, teacher.Name, teacher.Email, "").Scan(&teacherID) // search_key vacío por ahora

			if err != nil {
				// TODO: si no encuentra por email, acá podés decidir qué hacer (buscar por search_key, crear nuevo, skip, etc.)
				// Por ahora falla si hay conflicto o error
				return fmt.Errorf("error saving teacher %s (%s): %w", teacher.Name, teacher.Email, err)
			}

			teachersIDs = append(teachersIDs, teacherID)
		}

		// --- Update Curriculum data ---

		// Departament data
		var deptID int64
		deptName := grade.Curriculum.Subject.Department // ej: "DIF", "DCB"
		err = tx.QueryRowContext(ctx, `
            INSERT INTO departamentos (siglas)
            VALUES (?)
            ON CONFLICT(siglas) DO UPDATE SET siglas = excluded.siglas
            RETURNING id
        `, deptName).Scan(&deptID)
		if err != nil {
			return fmt.Errorf("error saving department: %w", err)
		}

		// Subject data
		var subjectID int64
		err = tx.QueryRowContext(ctx, `
            INSERT INTO asignaturas (nombre, departamento)
            VALUES (?, ?)
            ON CONFLICT(nombre) DO UPDATE SET departamento = excluded.departamento
            RETURNING id
        `, grade.Curriculum.Subject.Name, deptID).Scan(&subjectID)
		if err != nil {
			return fmt.Errorf("error saving subject: %w", err)
		}

		// Carrer data
		var carrerID int64
		err = tx.QueryRowContext(ctx, `
            INSERT INTO carreras (siglas)
            VALUES (?)
            ON CONFLICT(siglas) DO UPDATE SET siglas = excluded.siglas
            RETURNING id
        `, grade.Curriculum.Career).Scan(&carrerID)
		if err != nil {
			return fmt.Errorf("error saving career: %w", err)
		}

		// Curriculum data (career + subject + semester)
		var curriculumID int64
		err = tx.QueryRowContext(ctx, `
            INSERT INTO mallas (carrera, asignatura, semestre)
            VALUES (?, ?, ?)
            ON CONFLICT(carrera, asignatura) DO UPDATE SET semestre = excluded.semestre
            RETURNING id
        `, carrerID, subjectID, grade.Curriculum.Semester).Scan(&curriculumID)
		if err != nil {
			return fmt.Errorf("error saving curriculum: %w", err)
		}

		// Save all grade data (find by curriculum + section + period)
		var gradeID int64
		err = tx.QueryRowContext(ctx, `
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
			curriculumID, periodID, grade.Name, grade.Section,
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
		).Scan(&gradeID)
		if err != nil {
			return fmt.Errorf("error saving grade: %w", err)
		}

		// Update grade asociated teachers
		_, err = tx.ExecContext(ctx, "DELETE FROM docentes_curso WHERE id_curso = ?", gradeID)
		if err != nil {
			return fmt.Errorf("error deleting old teachers from grade: %w", err)
		}

		for _, teachID := range teachersIDs {
			_, err = tx.ExecContext(ctx, `
                INSERT INTO docentes_curso (id_docente, id_curso)
                VALUES (?, ?)
            `, teachID, gradeID)
			if err != nil {
				return fmt.Errorf("error asociating teachers to grade: %w", err)
			}
		}

		return nil
	}

	// run service layer logic
	if err := insertFn(persist); err != nil {
		return fmt.Errorf("insert callback: %w", err)
	}

	return tx.Commit()
}
