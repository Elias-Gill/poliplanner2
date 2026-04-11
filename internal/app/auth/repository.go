package auth

import "context"

type SessionRepository interface {
	Get(ctx context.Context, token SessionID) (*Session, error)
	Save(ctx context.Context, s *Session) error
	Delete(ctx context.Context, token SessionID) error
}
