package user

import (
	"context"
	model "github.com/elias-gill/poliplanner2/internal/model/user"
)

type UserRepository interface {
	Insert(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, userID model.UserID) error
	Save(ctx context.Context, user *model.User) error

	GetByID(ctx context.Context, userID model.UserID) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByRecoveryToken(ctx context.Context, token string) (*model.User, error)
}
