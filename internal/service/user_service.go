package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	"golang.org/x/crypto/bcrypt"
)

var (
	CannotFindUserError = errors.New("Cannot find user")
	BadCredentialsError = errors.New("Bad credentials")
)

type UserService struct {
	db         *sql.DB
	userStorer store.UserStorer
}

func NewUserService(
	db *sql.DB,
	userStorer store.UserStorer,
) *UserService {
	return &UserService{
		db:         db,
		userStorer: userStorer,
	}
}

func (s *UserService) AuthenticateUser(
	ctx context.Context,
	username string,
	rawPassword string,
) (*model.User, error) {

	// Search first by username. If not found then search by email
	user, err := s.userStorer.GetByUsername(ctx, s.db, username)
	if err != nil {
		user, err = s.userStorer.GetByEmail(ctx, s.db, username)
		if err != nil {
			return nil, CannotFindUserError
		}
	}

	// Verify credentials
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword))
	if err != nil {
		return nil, BadCredentialsError
	}

	return user, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	username string,
	email string,
	rawPassword string,
) error {

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(rawPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		panic(fmt.Sprintf("Cannot encrypt passwords: %+v", err))
	}

	return s.userStorer.Insert(ctx, s.db, &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	})
}
