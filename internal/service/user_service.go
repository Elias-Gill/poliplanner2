package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
	"github.com/elias-gill/poliplanner2/internal/db/store"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrEmailTaken         = errors.New("email already taken")
	ErrShortPassword      = errors.New("password too short")
	ErrPasswordsMismatch  = errors.New("passwords do not match")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type UserService struct {
	userStorer store.UserStorer
}

func NewUserService(userStorer store.UserStorer) *UserService {
	return &UserService{userStorer: userStorer}
}

func (s *UserService) AuthenticateUser(ctx context.Context, login string, rawPassword string) (*model.User, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	user, err := s.userStorer.GetByUsername(ctx, login)
	if err != nil {
		user, err = s.userStorer.GetByEmail(ctx, login)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rawPassword)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, username, email, rawPassword, confirmPassword string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))

	if len(username) < 3 {
		return ValidationError{Field: "username", Message: "must be at least 3 characters"}
	}

	if !isAlphanumeric(username) {
		return ValidationError{Field: "username", Message: "only letters, numbers, -, _ allowed"}
	}

	if !isValidEmail(email) {
		return ValidationError{Field: "email", Message: "invalid email format"}
	}

	if len(rawPassword) < 6 {
		return ValidationError{Field: "password", Message: "must be at least 6 characters"}
	}

	if rawPassword != confirmPassword {
		return ValidationError{Field: "confirm_password", Message: "passwords do not match"}
	}

	// Check uniqueness
	_, err := s.userStorer.GetByUsername(ctx, username)
	if err == nil {
		return ErrUsernameTaken
	}
	_, err = s.userStorer.GetByEmail(ctx, email)
	if err == nil {
		return ErrEmailTaken
	}

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
	email = strings.ToLower(strings.TrimSpace(email))

	if !isValidEmail(email) {
		return "", ValidationError{Field: "email", Message: "invalid email format"}
	}

	user, err := s.userStorer.GetByEmail(ctx, email)
	if err != nil {
		// No leak existence
		return "", nil
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
		return "", err
	}

	return token, nil
}

func (s *UserService) CommitPasswordRecovery(ctx context.Context, token, newPassword, confirmPassword string) error {
	if newPassword != confirmPassword {
		return ValidationError{Field: "confirm_password", Message: "passwords do not match"}
	}

	if len(newPassword) < 6 {
		return ValidationError{Field: "password", Message: "must be at least 6 characters"}
	}

	user, err := s.userStorer.GetByRecoveryToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	if user.RecoveryTokenUsed || user.RecoveryTokenExpiration.Before(time.Now()) {
		return ErrInvalidToken
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
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isAlphanumeric(str string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	return re.MatchString(str)
}
