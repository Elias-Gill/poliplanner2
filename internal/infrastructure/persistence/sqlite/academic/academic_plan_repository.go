package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type SqliteAcademicPlanStore struct {
	db *sql.DB
}

func NewSqliteAcademicPlanStore(connection *sql.DB) *SqliteAcademicPlanStore {
	return &SqliteAcademicPlanStore{db: connection}
}

func (s SqliteAcademicPlanStore) ListCareers(ctx context.Context) ([]*academic.Career, error) {
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

	var careers []*academic.Career

	for rows.Next() {
		var c academic.Career

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
	id academic.SubjectID,
) (*academic.Subject, error) {

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

	var subject academic.Subject

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
