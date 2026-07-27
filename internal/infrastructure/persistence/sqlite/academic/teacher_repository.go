package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type TeacherRepository struct {
	db *sql.DB
}

func NewTeacherRepository(db *sql.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

// Upsert inserts or updates a teacher record by email.
func (r *TeacherRepository) Upsert(ctx context.Context, t academic.Teacher) (academic.TeacherID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	var id int64
	err := exec.QueryRowContext(ctx, `
		INSERT INTO docentes (titulo, nombre, apellido, correo)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(correo) DO UPDATE SET
			titulo = excluded.titulo,
			nombre = excluded.nombre,
			apellido = excluded.apellido
		RETURNING id
	`, t.Title, t.FirstName, t.LastName, t.Email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return academic.TeacherID(id), nil
}

// GetByID retrieves a teacher by their database ID.
func (r *TeacherRepository) GetByID(ctx context.Context, id academic.TeacherID) (*academic.Teacher, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	var t academic.Teacher
	err := exec.QueryRowContext(ctx, `
		SELECT titulo, nombre, apellido, correo
		FROM docentes
		WHERE id = ?
	`, id).Scan(&t.Title, &t.FirstName, &t.LastName, &t.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}
