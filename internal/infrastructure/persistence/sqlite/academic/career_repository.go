package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	txManager "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type CareerRepository struct {
	db *sql.DB
}

func NewCareerRepository(db *sql.DB) *CareerRepository {
	return &CareerRepository{db: db}
}

// Upsert inserts a new career or updates the name if the code (siglas) already exists.
func (r *CareerRepository) Upsert(ctx context.Context, c academic.Career) (academic.CareerID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	_, err := exec.ExecContext(ctx, `
		INSERT INTO carreras (siglas, nombre)
		VALUES (?, ?)
		ON CONFLICT(siglas) DO UPDATE SET nombre = excluded.nombre
	`, c.Code, c.Name)
	if err != nil {
		return 0, err
	}

	var id int64
	err = exec.QueryRowContext(ctx, `SELECT id FROM carreras WHERE siglas = ?`, c.Code).Scan(&id)
	if err != nil {
		return 0, err
	}

	return academic.CareerID(id), nil
}

// GetByID retrieves a career by its database ID.
func (r *CareerRepository) GetByID(ctx context.Context, id academic.CareerID) (*academic.Career, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	var c academic.Career
	err := exec.QueryRowContext(ctx, `
		SELECT siglas, nombre
		FROM carreras
		WHERE id = ?
	`, id).Scan(&c.Code, &c.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}

// List returns all registered careers ordered alphabetically by code.
func (r *CareerRepository) List(ctx context.Context) ([]*academic.Career, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	rows, err := exec.QueryContext(ctx, `
		SELECT siglas, nombre, id
		FROM carreras
		ORDER BY siglas ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var careers []*academic.Career
	for rows.Next() {
		var c academic.Career
		if err := rows.Scan(&c.Code, &c.Name, &c.ID); err != nil {
			return nil, err
		}
		careers = append(careers, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return careers, nil
}

// ListPlans retrieves all unique study plans associated with a given career ID.
func (r *CareerRepository) ListPlans(ctx context.Context, id academic.CareerID) ([]academic.Plan, error) {
	const query = `
		SELECT p.codigo
		FROM planes p
		WHERE p.carrera = ?
		ORDER BY p.codigo ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans for career %d: %w", id, err)
	}
	defer rows.Close()

	var plans []academic.Plan
	for rows.Next() {
		var p academic.Plan
		if err := rows.Scan(&p.Code); err != nil {
			return nil, fmt.Errorf("failed to scan plan row: %w", err)
		}
		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plan rows: %w", err)
	}

	// Ensure an empty slice is returned instead of nil for consistent JSON serialization
	if plans == nil {
		plans = make([]academic.Plan, 0)
	}

	return plans, nil
}
