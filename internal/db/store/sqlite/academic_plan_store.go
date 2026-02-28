package sqlite

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteAcademicPlanStore struct {
	db *sql.DB
}

func NewSqliteAcademicPlanStore(db *sql.DB) *SqliteAcademicPlanStore {
	return &SqliteAcademicPlanStore{
		db: db,
	}
}

func (s SqliteAcademicPlanStore) GetByCareerID(
	ctx context.Context,
	careerID int64,
) (*model.AcademicPlan, error) {

	const query = `
		SELECT
			m.semestre,
			m.nivel,
			a.id,
			a.nombre,
			d.siglas
		FROM mallas m
		JOIN asignaturas a ON a.id = m.asignatura
		JOIN departamentos d ON d.id = a.departamento
		WHERE m.carrera = ?
		ORDER BY m.semestre ASC, a.nombre ASC;
	`

	rows, err := s.db.QueryContext(ctx, query, careerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plan := &model.AcademicPlan{
		ID:        careerID,
		Semesters: []model.Semester{},
	}

	var (
		currentSemesterNumber int
		currentSemester       *model.Semester
		firstRow              = true
	)

	for rows.Next() {
		var (
			semester   int
			level      int
			assignment model.Assignment
			name       string
			dept       string
		)

		if err := rows.Scan(
			&semester,
			&level,
			&assignment.ID,
			&name,
			&dept,
		); err != nil {
			return nil, err
		}

		assignment.Semester = semester
		assignment.Level = level
		assignment.Name = name
		assignment.Department = dept

		if firstRow || semester != currentSemesterNumber {
			if currentSemester != nil {
				plan.Semesters = append(plan.Semesters, *currentSemester)
			}

			currentSemesterNumber = semester
			currentSemester = &model.Semester{
				Number:      semester,
				Assignments: []model.Assignment{},
			}
			firstRow = false
		}

		currentSemester.Assignments = append(currentSemester.Assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if currentSemester != nil {
		plan.Semesters = append(plan.Semesters, *currentSemester)
	}

	if len(plan.Semesters) == 0 {
		return nil, nil
	}

	return plan, nil
}
