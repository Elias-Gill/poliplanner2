package user

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
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

type UserID int64

type User struct {
	ID UserID

	Username string
	Email    string
	Password string

	RecoveryTokenHash       *string
	RecoveryTokenExpiration *time.Time
	RecoveryTokenUsed       bool
}

func NewUser(username, email, password, confirm string) (*User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))

	if len(username) < 3 {
		return nil, ValidationError{"username", "must be at least 3 characters"}
	}

	if !isAlphanumeric(username) {
		return nil, ValidationError{"username", "only letters, numbers, -, _ allowed"}
	}

	if !isValidEmail(email) {
		return nil, ValidationError{"email", "invalid email format"}
	}

	if len(password) < 6 {
		return nil, ValidationError{"password", "must be at least 6 characters"}
	}

	if password != confirm {
		return nil, ValidationError{"confirm_password", "passwords do not match"}
	}

	return &User{
		Username: username,
		Email:    email,
	}, nil
}

func (u *User) SetPasswordHash(hash string) {
	u.Password = hash
}

func isAlphanumeric(str string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	return re.MatchString(str)
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
