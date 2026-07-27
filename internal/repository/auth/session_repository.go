package auth

import (
	"context"
	model "github.com/elias-gill/poliplanner2/internal/model/auth"
)

type AuthRepository interface {
	Get(ctx context.Context, token model.SessionID) (*model.Session, error)
	Save(ctx context.Context, s *model.Session) error
	Delete(ctx context.Context, token model.SessionID) error
}
