package store

import (
	"context"
	"database/sql"
	"github.com/elias-gill/poliplanner2/db/model"
)

type SqliteUserStore struct {
	db *sql.DB
}

func NewSqliteUserStore(db *sql.DB) *SqliteUserStore {
	return &SqliteUserStore{db: db}
}

func (s *SqliteUserStore) Insert(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (username, password, email)
		VALUES (?, ?, ?)
	`
	res, err := s.db.ExecContext(ctx, query, u.Username, u.Password, u.Email)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.UserID = id
	return nil
}

func (s *SqliteUserStore) Delete(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE user_id = ?`, userID)
	return err
}

func (s *SqliteUserStore) GetByID(ctx context.Context, userID int64) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE user_id = ?`, userID).
		Scan(&u.UserID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *SqliteUserStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE username = ?`, username).
		Scan(&u.UserID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *SqliteUserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE email = ?`, email).
		Scan(&u.UserID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err != nil {
		return nil, err
	}
	return u, nil
}
