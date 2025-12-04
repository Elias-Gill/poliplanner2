package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	CannotFindUserError = errors.New("Cannot find user")
	BadCredentialsError = errors.New("Bad credentials")
)

func AuthenticateUser(ctx context.Context, username string, rawPassword string) (*model.User, error) {
	// Search first by username. If not found then search by email
	user, err := userStorer.GetByUsername(ctx, db, username)
	if err != nil {
		user, err = userStorer.GetByEmail(ctx, db, username)
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

func CreateUser(ctx context.Context, username string, email string, rawPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("Cannot encrypt passwords: %+v", err))
	}

	return userStorer.Insert(ctx, db, &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	})
}
