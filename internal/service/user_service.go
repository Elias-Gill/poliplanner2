package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"golang.org/x/crypto/bcrypt"
)

func AuthenticateUser(username string, rawPassword string) (*model.User, error) {
	// Search first by username. If not found then search by email
	user, err := userStorer.GetByUsername(context.TODO(), username)
	if err != nil {
		user, err = userStorer.GetByEmail(context.TODO(), username)
		if err != nil {
			return nil, fmt.Errorf("Cannot find user")
		}
	}

	// Verify credentials
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword))
	if err != nil {
		return nil, fmt.Errorf("Incorrect password")
	}

	return user, nil
}

func CreateUser(username string, email string, rawPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return userStorer.Insert(context.TODO(), &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	})
}
