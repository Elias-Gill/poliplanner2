package user

type UserStorer interface {
	Insert(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, userID int64) error
	Update(ctx context.Context, userID int64, updateFn func(user *model.User) error) error

	GetByID(ctx context.Context, userID int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByRecoveryToken(ctx context.Context, token string) (*model.User, error)
}
