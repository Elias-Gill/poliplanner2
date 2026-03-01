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

	var plan *model.AcademicPlan
	var currentSemesterNumber int

	for rows.Next() {
		var (
			semester int
			level    int
			id       int64
			name     string
			dept     string
		)

		// read row values
		if err := rows.Scan(&semester, &level, &id, &name, &dept); err != nil {
			return nil, err
		}

		// Initialize plan if not exists
		if plan == nil {
			plan = &model.AcademicPlan{
				ID:        careerID,
				Semesters: make([]model.Semester, 0),
			}
		}

		// Create a new semester if necesary
		if len(plan.Semesters) == 0 || semester != currentSemesterNumber {
			plan.Semesters = append(plan.Semesters, model.Semester{
				Number:      semester,
				Assignments: make([]model.Assignment, 0),
			})
			currentSemesterNumber = semester
		}

		// Append semester
		lastIdx := len(plan.Semesters) - 1
		plan.Semesters[lastIdx].Assignments = append(
			plan.Semesters[lastIdx].Assignments,
			model.Assignment{
				ID:         id,
				Name:       name,
				Semester:   semester,
				Level:      level,
				Department: dept,
			},
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// if there is no plan, then this returns "nil", "nil"
	return plan, nil
}
