package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteUserStore struct {
	db *sql.DB
}

func NewSqliteUserStore(db *sql.DB) *SqliteUserStore {
	return &SqliteUserStore{
		db: db,
	}
}

func (s SqliteUserStore) Insert(ctx context.Context, u *model.User) error {
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
	u.ID = id
	return nil
}

func (s *SqliteUserStore) Update(ctx context.Context, userID int64, updateFn func(user *model.User) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var user model.User
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, username, password, email, recovery_token_hash,
		       recovery_token_expiration, recovery_token_used
		FROM users WHERE user_id = ? FOR UPDATE`,
		userID,
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

	if err := updateFn(&user); err != nil {
		return fmt.Errorf("update canceled: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
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
		user.RecoveryTokenUsed, userID,
	)
	if err != nil {
		return fmt.Errorf("update exec error: %w", err)
	}

	return tx.Commit()
}

func (s SqliteUserStore) Delete(ctx context.Context, userID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = s.db.ExecContext(ctx, `DELETE FROM users WHERE user_id = ?`, userID)

	return tx.Commit()
}

func (s SqliteUserStore) GetByID(ctx context.Context, userID int64) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
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

func (s SqliteUserStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
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

func (s SqliteUserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
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

func (s SqliteUserStore) GetByRecoveryToken(ctx context.Context, token string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, password, email,
		recovery_token_hash, recovery_token_expiration, recovery_token_used
		FROM users WHERE recovery_token_hash = ?`, token).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email,
			&u.RecoveryTokenHash, &u.RecoveryTokenExpiration, &u.RecoveryTokenUsed)
	if err != nil {
		return nil, err
	}
	return u, nil
}
