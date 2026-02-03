package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	"golang.org/x/crypto/bcrypt"
)

var (
	CannotFindUserError = errors.New("Cannot find user")
	BadCredentialsError = errors.New("Bad credentials")
)

type UserService struct {
	userStorer store.UserStorer
}

func NewUserService(
	userStorer store.UserStorer,
) *UserService {
	return &UserService{
		userStorer: userStorer,
	}
}

func (s *UserService) AuthenticateUser(ctx context.Context, login, rawPassword string) (*model.User, error) {
	login = strings.ToLower(login)

	user, err := s.userStorer.GetByUsername(ctx, login)
	if err != nil {
		user, err = s.userStorer.GetByEmail(ctx, login)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, username, email, rawPassword string) error {
	username = strings.ToLower(username)
	email = strings.ToLower(email)

	hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.userStorer.Insert(ctx, &model.User{
		Username: username,
		Password: string(hashed),
		Email:    email,
	})
}

func (s *UserService) StartPasswordRecovery(ctx context.Context, email string) (string, error) {
	email = strings.ToLower(email)

	user, err := s.userStorer.GetByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	token := newRecoveryToken()
	expiration := time.Now().Add(15 * time.Minute)

	err = s.userStorer.Update(ctx, user.ID, func(u *model.User) error {
		u.RecoveryTokenHash = token
		u.RecoveryTokenExpiration = &expiration
		u.RecoveryTokenUsed = false
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("update recovery token: %w", err)
	}

	return token, nil
}

func (s *UserService) CommitPasswordRecovery(ctx context.Context, token, newPassword string) error {
	user, err := s.userStorer.GetByRecoveryToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if user.RecoveryTokenUsed || user.RecoveryTokenExpiration.Before(time.Now()) {
		return errors.New("invalid or expired token")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	return s.userStorer.Update(ctx, user.ID, func(u *model.User) error {
		u.Password = string(hashed)
		u.RecoveryTokenUsed = true
		u.RecoveryTokenHash = ""
		u.RecoveryTokenExpiration = nil
		return nil
	})
}

func newRecoveryToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
