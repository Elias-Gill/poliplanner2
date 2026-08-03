package sqlite

import (
	"context"
	"database/sql"
	"errors"

	txManager "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
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

	// Prepare database email parameter: NULL when empty, string otherwise
	var dbEmail any = t.Email
	if t.Email == "" {
		dbEmail = nil
	}

	var existingID int64
	var existingEmail string

	// Primary search: match by full name combination.
	// COALESCE avoids scan errors in Go if stored 'correo' is SQL NULL.
	err := exec.QueryRowContext(ctx, `
		SELECT id, COALESCE(correo, '') 
		FROM docentes 
		WHERE nombre = ? AND apellido = ? 
		LIMIT 1
	`, t.FirstName, t.LastName).Scan(&existingID, &existingEmail)

	// Primary match found: update teacher info while preserving existing email if incoming email is empty.
	if err == nil {
		var finalEmail any = t.Email
		if t.Email == "" && existingEmail != "" {
			finalEmail = existingEmail
		} else if t.Email == "" {
			finalEmail = nil
		}

		_, updateErr := exec.ExecContext(ctx, `
			UPDATE docentes 
			SET titulo = ?, nombre = ?, apellido = ?, correo = ? 
			WHERE id = ?
		`, t.Title, t.FirstName, t.LastName, finalEmail, existingID)
		if updateErr != nil {
			return 0, updateErr
		}

		return academic.TeacherID(existingID), nil
	}

	if err != sql.ErrNoRows {
		return 0, err
	}

	// Secondary search: match by email (only executed when an incoming email is present).
	if t.Email != "" {
		err = exec.QueryRowContext(ctx, `
			SELECT id 
			FROM docentes 
			WHERE correo = ? 
			LIMIT 1
		`, t.Email).Scan(&existingID)

		if err == nil {
			_, updateErr := exec.ExecContext(ctx, `
				UPDATE docentes 
				SET titulo = ?, nombre = ?, apellido = ?, correo = ? 
				WHERE id = ?
			`, t.Title, t.FirstName, t.LastName, dbEmail, existingID)
			if updateErr != nil {
				return 0, updateErr
			}

			return academic.TeacherID(existingID), nil
		}

		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	// No match found: insert a new teacher record.
	var newID int64
	err = exec.QueryRowContext(ctx, `
		INSERT INTO docentes (titulo, nombre, apellido, correo) 
		VALUES (?, ?, ?, ?) 
		RETURNING id
	`, t.Title, t.FirstName, t.LastName, dbEmail).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return academic.TeacherID(newID), nil
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
