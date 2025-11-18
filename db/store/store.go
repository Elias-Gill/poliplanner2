package store

import (
	"context"

	"github.com/elias-gill/poliplanner2/db/models"
)

type UserStorer interface {
	Insert(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, userID int64) error
	GetByID(ctx context.Context, userID int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}
