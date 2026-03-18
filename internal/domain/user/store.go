package user

import "context"

type UserStorer interface {
	Insert(ctx context.Context, u *User) error
	Delete(ctx context.Context, userID UserID) error
	Update(ctx context.Context, userID UserID, updateFn func(user *User) error) error

	GetByID(ctx context.Context, userID UserID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByRecoveryToken(ctx context.Context, token string) (*User, error)
}
