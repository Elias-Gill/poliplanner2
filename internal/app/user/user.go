package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userStorer user.UserStorer
}

func NewUserService(userStorer user.UserStorer) *UserService {
	return &UserService{userStorer: userStorer}
}

func (s *UserService) AuthenticateUser(ctx context.Context, login string, rawPassword string) (*user.User, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	u, err := s.userStorer.GetByUsername(ctx, login)
	if err != nil {
		u, err = s.userStorer.GetByEmail(ctx, login)
		if err != nil {
			return nil, user.ErrInvalidCredentials
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(rawPassword)); err != nil {
		return nil, user.ErrInvalidCredentials
	}

	return u, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	username,
	email,
	rawPassword,
	confirmPassword string,
) error {

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

	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(rawPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	u.SetPasswordHash(string(hashed))

	return s.userStorer.Insert(ctx, u)
}

func (s *UserService) StartPasswordRecovery(ctx context.Context, email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if !isValidEmail(email) {
		return "", user.ValidationError{Field: "email", Message: "invalid email format"}
	}

	u, err := s.userStorer.GetByEmail(ctx, email)
	if err != nil {
		// No leak existence
		return "", nil
	}

	token := newRecoveryToken()
	expiration := time.Now().In(timezone.ParaguayTZ).Add(15 * time.Minute)

	err = s.userStorer.Update(
		ctx, u.ID,
		func(u *user.User) error {
			u.RecoveryTokenHash = &token
			u.RecoveryTokenExpiration = &expiration
			u.RecoveryTokenUsed = false
			return nil
		})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) CommitPasswordRecovery(ctx context.Context, token, newPassword, confirmPassword string) error {
	if newPassword != confirmPassword {
		return user.ValidationError{Field: "confirm_password", Message: "passwords do not match"}
	}

	if len(newPassword) < 6 {
		return user.ValidationError{Field: "password", Message: "must be at least 6 characters"}
	}

	u, err := s.userStorer.GetByRecoveryToken(ctx, token)
	if err != nil {
		return user.ErrInvalidToken
	}

	if u.RecoveryTokenUsed || u.RecoveryTokenExpiration.Before(time.Now().In(timezone.ParaguayTZ)) {
		return user.ErrInvalidToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	return s.userStorer.Update(ctx, u.ID, func(u *user.User) error {
		u.Password = string(hashed)
		u.RecoveryTokenUsed = true
		u.RecoveryTokenHash = nil
		u.RecoveryTokenExpiration = nil
		return nil
	})
}

func newRecoveryToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
