package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	txManager "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
)

type CurriculumRepository struct {
	db *sql.DB
}

func NewCurriculumRepository(db *sql.DB) *CurriculumRepository {
	return &CurriculumRepository{db: db}
}

func (r *CurriculumRepository) Upsert(ctx context.Context, c academicRepo.CurriculumSaveParams) (academic.CurriculumID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	// Insert plan data
	var planID int64
	err := exec.QueryRowContext(ctx, `
		INSERT INTO planes (carrera, codigo)
		VALUES (?, ?)
		ON CONFLICT(codigo, carrera) DO UPDATE SET codigo = excluded.codigo
		RETURNING id
	`, c.CareerID, c.Curriculum.Plan.Code).Scan(&planID)
	if err != nil {
		return 0, fmt.Errorf("error al upsert de plan: %w", err)
	}

	// Insert the curriculum
	var mallaID int64
	err = exec.QueryRowContext(ctx, `
		INSERT INTO mallas (carrera, plan, asignatura, semestre, nivel)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(carrera, asignatura, plan) DO UPDATE SET
			semestre = excluded.semestre,
			nivel = excluded.nivel
		RETURNING id
	`, c.CareerID, planID, c.SubjectID, c.Curriculum.Semester, c.Curriculum.Level).Scan(&mallaID)
	if err != nil {
		return 0, fmt.Errorf("error al upsert de malla: %w", err)
	}

	// Insert and link emphases
	for _, emp := range c.Curriculum.Emphases {
		careerID := emp.Career
		if careerID == 0 {
			careerID = c.CareerID
		}

		// insert
		var enfasisID int64
		err := exec.QueryRowContext(ctx, `
			INSERT INTO enfasis (carrera, codigo, nombre)
			VALUES (?, ?, ?)
			ON CONFLICT(codigo, carrera) DO UPDATE SET
				nombre = excluded.nombre
			RETURNING id
		`, careerID, emp.Code, emp.Name).Scan(&enfasisID)
		if err != nil {
			return 0, fmt.Errorf("error al upsert de enfasis (%s): %w", emp.Code, err)
		}

		// link
		_, err = exec.ExecContext(ctx, `
			INSERT INTO enfasis_materia (malla, enfasis)
			VALUES (?, ?)
			ON CONFLICT(malla, enfasis) DO NOTHING
		`, mallaID, enfasisID)
		if err != nil {
			return 0, fmt.Errorf("error al vincular enfasis (%s) a la malla: %w", emp.Code, err)
		}
	}

	return academic.CurriculumID(mallaID), nil
}

func (r *CurriculumRepository) GetByCareerID(ctx context.Context, career academic.CareerID) ([]academic.CurriculumSubjectItem, error) {
	query := `
		SELECT m.id, m.semestre, m.nivel, s.nombre, p.codigo
		FROM mallas m join asignaturas s
		on s.id = m.asignatura
		join planes p on p.id = m.plan
		WHERE m.carrera = $1
		ORDER BY semestre ASC, s.nombre ASC
	`

	rows, err := r.db.QueryContext(ctx, query, career)
	if err != nil {
		return nil, fmt.Errorf("error al consultar materias para la carrera %v: %w", career, err)
	}
	defer rows.Close()

	// Initialize with 0 to prevent nil values when the career does not have curriculums
	subjects := make([]academic.CurriculumSubjectItem, 0)

	for rows.Next() {
		var s academic.CurriculumSubjectItem
		if err := rows.Scan(&s.ID, &s.Semester, &s.Level, &s.Name, &s.Plan); err != nil {
			return nil, fmt.Errorf("error al escanear materia del currículum: %w", err)
		}
		subjects = append(subjects, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando filas de materias: %w", err)
	}

	return subjects, nil
}
