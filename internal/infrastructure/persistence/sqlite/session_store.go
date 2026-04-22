package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elias-gill/poliplanner2/internal/app/auth"
)

type SqliteSessionStore struct {
	db *sql.DB
}

func NewSqliteSessionStore(db *sql.DB) *SqliteSessionStore {
	return &SqliteSessionStore{
		db: db,
	}
}

func (s SqliteSessionStore) Get(ctx context.Context, token auth.SessionID) (*auth.Session, error) {
	const query = `
		SELECT 
			session_token,
			user_id,
			created_at,
			expires_at,
			last_used_at
		FROM user_sessions
		WHERE session_token = ?
	`

	row := s.db.QueryRowContext(ctx, query, token)

	var ses auth.Session
	var sessionToken string

	err := row.Scan(
		&sessionToken,
		&ses.User,
		&ses.CreatedAt,
		&ses.Expiration,
		&ses.LastUse,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No token found, but no error too
			return nil, nil
		}
		return nil, err
	}

	ses.ID = auth.SessionID(sessionToken)

	return &ses, nil
}

func (s SqliteSessionStore) Save(ctx context.Context, ses *auth.Session) error {
	const query = `
		INSERT INTO user_sessions (
			session_token,
			user_id,
			expires_at,
			created_at,
			last_used_at
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(session_token) DO UPDATE SET
			user_id = excluded.user_id,
			expires_at = excluded.expires_at,
			created_at = excluded.created_at,
			last_used_at = excluded.last_used_at
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		ses.ID,
		ses.User,
		ses.Expiration,
		ses.CreatedAt,
		ses.LastUse,
	)

	return err
}

func (s SqliteSessionStore) Delete(ctx context.Context, token auth.SessionID) error {
	const query = `
		DELETE FROM user_sessions
		WHERE session_token = ?
	`

	_, err := s.db.ExecContext(ctx, query, token)
	return err
}
