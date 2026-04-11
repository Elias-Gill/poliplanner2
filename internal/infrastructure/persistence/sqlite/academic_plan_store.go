package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/domain/academicPlan"
)

type SqliteAcademicPlanStore struct {
	db *sql.DB
}

func NewSqliteAcademicPlanStore(connection *sql.DB) *SqliteAcademicPlanStore {
	return &SqliteAcademicPlanStore{db: connection}
}

func (s SqliteAcademicPlanStore) ListCareers(ctx context.Context) ([]*academicPlan.Career, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT id, siglas
		FROM carreras
		ORDER BY siglas
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query careers: %w", err)
	}
	defer rows.Close()

	var careers []*academicPlan.Career

	for rows.Next() {
		var c academicPlan.Career

		err := rows.Scan(
			&c.ID,
			&c.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan career: %w", err)
		}

		careers = append(careers, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating careers: %w", err)
	}

	return careers, nil
}

func (s SqliteAcademicPlanStore) GetSubject(
	ctx context.Context,
	id academicPlan.SubjectID,
) (*academicPlan.Subject, error) {

	row := s.db.QueryRowContext(
		ctx,
		`
		SELECT 
			a.id,
			a.nombre,
			d.siglas
		FROM asignaturas a
		JOIN mallas m ON m.asignatura = a.id
		JOIN departamentos d ON d.id = a.departamento
		WHERE a.id = ?
		LIMIT 1
		`,
		id,
	)

	var subject academicPlan.Subject

	err := row.Scan(
		&subject.ID,
		&subject.Name,
		&subject.Department,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get subject: %w", err)
	}

	return &subject, nil
}

func (s SqliteAcademicPlanStore) GetPlanByCareerID(
	ctx context.Context,
	career academicPlan.CareerID,
) ([]academicPlan.Subject, error) {

	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT 
			a.id,
			a.nombre,
			d.siglas,
			m.nivel,
			m.semestre
		FROM mallas m
		JOIN asignaturas a ON a.id = m.asignatura
		JOIN departamentos d ON d.id = a.departamento
		WHERE m.carrera = ?
		ORDER BY 
			CASE 
				WHEN m.semestre > 0 THEN m.semestre
				ELSE m.nivel
			END,
			a.nombre
		`,
		career,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query academic plan subjects: %w", err)
	}
	defer rows.Close()

	var subjects []academicPlan.Subject

	for rows.Next() {
		var subject academicPlan.Subject

		err := rows.Scan(
			&subject.ID,
			&subject.Name,
			&subject.Department,
			&subject.Level,
			&subject.Semester,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subject: %w", err)
		}

		subjects = append(subjects, subject)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subjects: %w", err)
	}

	if len(subjects) == 0 {
		return nil, nil
	}

	return subjects, nil
}

func (s SqliteAcademicPlanStore) GetCareer(
	ctx context.Context,
	id academicPlan.CareerID,
) (*academicPlan.Career, error) {

	row := s.db.QueryRowContext(
		ctx,
		`
		SELECT id, siglas
		FROM carreras
		WHERE id = ?
		LIMIT 1
		`,
		id,
	)

	var career academicPlan.Career

	err := row.Scan(
		&career.ID,
		&career.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get career: %w", err)
	}

	return &career, nil
}
