package user

import (
	"context"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"strings"
)

type UserService struct {
	userStorer user.UserStorer
}

func NewUserService(userStorer user.UserStorer) *UserService {
	return &UserService{userStorer: userStorer}
}

func (s *UserService) CreateUser(
	ctx context.Context,
	username,
	email,
	rawPassword,
	confirmPassword string,
) error {
	// Valid and create user fields
	u, err := user.NewUser(username, email, rawPassword, confirmPassword)
	if err != nil {
		return err
	}

	// Check if username is already taken
	_, err = s.userStorer.GetByUsername(ctx, u.Username)
	if err == nil {
		return user.ErrUsernameTaken
	}

	// Check if email is already taken
	_, err = s.userStorer.GetByEmail(ctx, u.Email)
	if err == nil {
		return user.ErrEmailTaken
	}

	return s.userStorer.Insert(ctx, u)
}

func (s *UserService) StartPasswordRecovery(ctx context.Context, email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := user.ValidateEmailInput(email); err != nil {
		return "", err
	}

	u, err := s.userStorer.GetByEmail(ctx, email)
	if err != nil {
		// No leak existence
		return "", nil
	}

	token := u.SetupRecovery()

	err = s.userStorer.Save(ctx, u)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) CommitPasswordRecovery(ctx context.Context, token, newPassword, confirmPassword string) error {
	// Validate passwords
	if err := user.ValidatePasswordInput(newPassword, confirmPassword); err != nil {
		return err
	}

	// Retrieve user
	u, err := s.userStorer.GetByRecoveryToken(ctx, token)
	if err != nil {
		return user.ErrInvalidToken
	}

	err = u.ConfirmRecovery(newPassword)
	if err != nil {
		return err
	}

	// Save
	return s.userStorer.Save(ctx, u)
}
