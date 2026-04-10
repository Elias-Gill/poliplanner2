package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"github.com/elias-gill/poliplanner2/logger"
)

type SqliteUserStore struct {
	db *sql.DB
}

func NewSqliteUserStore(db *sql.DB) *SqliteUserStore {
	return &SqliteUserStore{
		db: db,
	}
}

func (s SqliteUserStore) Insert(ctx context.Context, u *user.User) error {
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
	u.ID = user.UserID(id)
	return nil
}

// WIP: realmente ni hace falta meter en una transaccion
// FIX: revisar que todo funcione
func (s *SqliteUserStore) Save(ctx context.Context, u *user.User) error {
	var user user.User
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email, recovery_token_hash,
		       recovery_token_expiration, recovery_token_used
		FROM users WHERE user_id = ? FOR UPDATE`,
		u.ID,
	).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.RecoveryTokenHash, &user.RecoveryTokenExpiration, &user.RecoveryTokenUsed,
	)
	if err == sql.ErrNoRows {
		return fmt.Errorf("user not found")
	}
	if err != nil {
		return fmt.Errorf("get user error: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE users SET
			username = ?,
			password = ?,
			email = ?,
			recovery_token_hash = ?,
			recovery_token_expiration = ?,
			recovery_token_used = ?
		WHERE user_id = ?`,
		user.Username, user.Password, user.Email,
		user.RecoveryTokenHash, user.RecoveryTokenExpiration,
		user.RecoveryTokenUsed, u.ID,
	)
	if err != nil {
		return fmt.Errorf("update exec error: %w", err)
	}

	return nil
}

func (s SqliteUserStore) Delete(ctx context.Context, userID user.UserID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete exec error: %w", err)
	}

	return nil
}

func (s SqliteUserStore) GetByID(ctx context.Context, userID user.UserID) (*user.User, error) {
	u := &user.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE user_id = ?`, userID).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err == sql.ErrNoRows {
		logger.Debug("UserID not found", "id", userID)
		return nil, err
	}
	if err != nil {
		logger.Warn("Database error searching userID", "userID", userID, "error", err)
		return nil, err
	}
	return u, nil
}

func (s SqliteUserStore) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	u := &user.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT u.user_id, u.username, u.password, u.email,
		       u.recovery_token_hash, u.recovery_token_expiration, u.recovery_token_used
		FROM users u WHERE u.username = ?`, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err == sql.ErrNoRows {
		logger.Debug("User not found", "username", username)
		return nil, err
	}
	if err != nil {
		logger.Warn("Database error searching user", "username", username, "error", err)
		return nil, err
	}
	return u, nil
}

func (s SqliteUserStore) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u := &user.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		       recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err == sql.ErrNoRows {
		logger.Debug("Email not found", "email", email)
		return nil, err
	}
	if err != nil {
		logger.Warn("Database error searching email", "email", email, "error", err)
		return nil, err
	}
	return u, nil
}

func (s SqliteUserStore) GetByRecoveryToken(ctx context.Context, token string) (*user.User, error) {
	u := &user.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE recovery_token_hash = ?`, token).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err == sql.ErrNoRows {
		logger.Debug("Token not found", "token", token)
		return nil, err
	}
	if err != nil {
		logger.Warn("Database error searching token", "token", token, "error", err)
		return nil, err
	}
	return u, nil
}
