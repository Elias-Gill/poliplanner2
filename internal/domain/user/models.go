package user

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"golang.org/x/crypto/bcrypt"
)

// ============================
// =          Errors          =
// ============================

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

// ============================
// =          Models          =
// ============================

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

// NewUser creates a new User instance after validating the username, email, and password.
// Returns a ValidationError if any field is invalid. Also hashes the password before returning the User.
func NewUser(username, email, password, confirm string) (*User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))

	if len(username) < 3 {
		return nil, ValidationError{"username", "must be at least 3 characters"}
	}

	if !isAlphanumeric(username) {
		return nil, ValidationError{"username", "only letters, numbers, -, _ allowed"}
	}

	if err := ValidateEmailInput(email); err != nil {
		return nil, err
	}

	if err := ValidatePasswordInput(password, confirm); err != nil {
		return nil, err
	}

	u := &User{
		Username: username,
		Email:    email,
	}
	u.SetPassword(password)

	return u, nil
}

// =============================
// =          Methods          =
// =============================

// SetPassword hashes the provided password using bcrypt and stores it in the User struct.
// Returns an error if the hashing fails.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	u.Password = string(hash)

	return nil
}

// GetPassword returns the hashed password stored in the User struct.
// Typically used internally for authentication or persistence.
func (u *User) GetPassword() string {
	return u.Password
}

// AuthenticatePassword compares a raw password with the stored hashed password.
// Returns ErrInvalidCredentials if the password does not match.
func (u *User) AuthenticatePassword(rawPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(rawPassword)); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}

// SetupRecovery generates a new recovery token for the user, sets its expiration time,
// and marks it as unused. Returns the generated token string.
func (u *User) SetupRecovery() string {
	token := newRecoveryToken()
	expiration := time.Now().In(timezone.ParaguayTZ).Add(15 * time.Minute)

	u.RecoveryTokenExpiration = &expiration
	u.RecoveryTokenHash = &token
	u.RecoveryTokenUsed = false

	return token
}

// ConfirmRecovery validates the recovery token state and expiration, sets a new password,
// and marks the token as used. Returns ErrInvalidToken if the token is expired or already used.
func (u *User) ConfirmRecovery(newPassword string) error {
	// Check expiration and usage
	if u.RecoveryTokenUsed || u.RecoveryTokenExpiration.Before(time.Now().In(timezone.ParaguayTZ)) {
		return ErrInvalidToken
	}

	err := u.SetPassword(newPassword)
	if err != nil {
		return err
	}

	u.RecoveryTokenUsed = true
	u.RecoveryTokenHash = nil
	u.RecoveryTokenExpiration = nil

	return nil
}

// ================================
// =          Public Fns          =
// ================================

// ValidatePasswordInput checks that the password meets minimum length requirements
// and matches the confirmation. Returns a ValidationError on failure.
func ValidatePasswordInput(password string, confirm string) error {
	if len(password) < 6 {
		return ValidationError{"password", "must be at least 6 characters"}
	}

	if password != confirm {
		return ValidationError{"confirm_password", "passwords do not match"}
	}

	return nil
}

// ValidateEmailInput checks that the provided email has a valid format.
// Returns a ValidationError on invalid email.
func ValidateEmailInput(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ValidationError{"email", "invalid email format"}
	}

	return nil
}

// ===========================
// =          Utils          =
// ===========================

func isAlphanumeric(str string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	return re.MatchString(str)
}

func newRecoveryToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
