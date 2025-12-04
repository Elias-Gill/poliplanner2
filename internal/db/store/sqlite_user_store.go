package store

import (
	"context"
	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteUserStore struct {
}

func NewSqliteUserStore() *SqliteUserStore {
	return &SqliteUserStore{}
}

func (s SqliteUserStore) Insert(ctx context.Context, exec Executor, u *model.User) error {
	query := `
		INSERT INTO users (username, password, email)
		VALUES (?, ?, ?)
	`
	res, err := exec.ExecContext(ctx, query, u.Username, u.Password, u.Email)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (s SqliteUserStore) Delete(ctx context.Context, exec Executor, userID int64) error {
	_, err := exec.ExecContext(ctx, `DELETE FROM users WHERE user_id = ?`, userID)
	return err
}

func (s SqliteUserStore) GetByID(ctx context.Context, exec Executor, userID int64) (*model.User, error) {
	u := &model.User{}
	err := exec.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE user_id = ?`, userID).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s SqliteUserStore) GetByUsername(ctx context.Context, exec Executor, username string) (*model.User, error) {
	u := &model.User{}
	err := exec.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s SqliteUserStore) GetByEmail(ctx context.Context, exec Executor, email string) (*model.User, error) {
	u := &model.User{}
	err := exec.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err != nil {
		return nil, err
	}
	return u, nil
}
