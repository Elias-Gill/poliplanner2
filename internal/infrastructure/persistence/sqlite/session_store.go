package sqlite

import (
	"context"
	"database/sql"

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
	// TODO: implement
	return auth.NewSession(1), nil
}

func (s SqliteSessionStore) Save(ctx context.Context, ses *auth.Session) error {
	// TODO implement
	return nil
}

func (s SqliteSessionStore) Delete(ctx context.Context, token auth.SessionID) error {
	// TODO: implement
	return nil
}
