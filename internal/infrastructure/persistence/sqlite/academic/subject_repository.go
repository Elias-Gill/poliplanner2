package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type SubjectRepository struct {
	db *sql.DB
}

func NewSubjectRepository(db *sql.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

// Upsert inserts or updates a department and its associated subject.
func (r *SubjectRepository) Upsert(ctx context.Context, s academic.Subject) (academic.SubjectID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	// Upsert department (now includes 'nombre')
	_, err := exec.ExecContext(ctx, `
		INSERT INTO departamentos (siglas, nombre)
		VALUES (?, ?)
		ON CONFLICT(siglas) DO UPDATE SET nombre = excluded.nombre
	`, s.Department.Code, s.Department.Name)
	if err != nil {
		return 0, err
	}

	var deptID int64
	err = exec.QueryRowContext(ctx, `SELECT id FROM departamentos WHERE siglas = ?`, s.Department.Code).Scan(&deptID)
	if err != nil {
		return 0, err
	}

	// Upsert subject
	_, err = exec.ExecContext(ctx, `
		INSERT INTO asignaturas (nombre, departamento)
		VALUES (?, ?)
		ON CONFLICT(nombre) DO UPDATE SET departamento = excluded.departamento
	`, s.Name, deptID)
	if err != nil {
		return 0, err
	}

	var id int64
	err = exec.QueryRowContext(ctx, `SELECT id FROM asignaturas WHERE nombre = ?`, s.Name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return academic.SubjectID(id), nil
}

// GetByID retrieves a subject by its database ID, joining its corresponding department.
func (r *SubjectRepository) GetByID(ctx context.Context, id academic.SubjectID) (*academic.Subject, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	var s academic.Subject
	err := exec.QueryRowContext(ctx, `
		SELECT a.nombre, d.siglas, d.nombre
		FROM asignaturas a
		INNER JOIN departamentos d ON a.departamento = d.id
		WHERE a.id = ?
	`, id).Scan(&s.Name, &s.Department.Code, &s.Department.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}
