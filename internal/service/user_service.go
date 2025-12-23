package service

import (
	"context"
	"crypto/rand"
	"database/sql"
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
	// Normalize username
	username = strings.ToLower(username)

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
	// Normalize
	username = strings.ToLower(username)
	email = strings.ToLower(email)

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

// Starts a new password recovery process, creating a new recovery token for the given user
func (s *UserService) StartPasswordRecovery(
	ctx context.Context,
	email string,
) (string, error) {
	user, err := s.userStorer.GetByEmail(ctx, s.db, strings.ToLower(email))
	if err != nil {
		return "", CannotFindUserError
	}

	token := newRecoveryToken()
	expiration := time.Now().Add(time.Minute * 15)

	user.RecoveryTokenExpiration = sql.NullTime{Valid: true, Time: expiration}
	user.RecoveryTokenHash = sql.NullString{Valid: true, String: token}
	user.RecoveryTokenUsed = false

	err = s.userStorer.Update(ctx, s.db, user)
	if err != nil {
		return "", fmt.Errorf("Cannot update user in database: %w", err)
	}

	return token, nil
}

func (s *UserService) CommitPasswordRecovery(ctx context.Context, token string, newPassword string) error {
	user, err := s.userStorer.GetByRecoveryToken(ctx, s.db, token)
	if err != nil {
		return fmt.Errorf("Cannot update user in database: %w", err)
	}

	if user.RecoveryTokenUsed ||
		!user.RecoveryTokenHash.Valid ||
		!user.RecoveryTokenExpiration.Valid ||
		user.RecoveryTokenExpiration.Time.Before(time.Now()) {
		return fmt.Errorf("Token already expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		panic(fmt.Sprintf("Cannot encrypt passwords: %+v", err))
	}

	user.Password = string(hashedPassword)

	// Invalidate used tokens
	user.RecoveryTokenUsed = true
	user.RecoveryTokenHash = sql.NullString{Valid: false}
	user.RecoveryTokenExpiration = sql.NullTime{Valid: false}

	err = s.userStorer.Update(ctx, s.db, user)
	if err != nil {
		return fmt.Errorf("Error commiting password recovery: %w", err)
	}

	return nil
}

func newRecoveryToken() string {
	b := make([]byte, 32) // 256-bit random recovery token
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("Cannot generate secure recovery Token %+v", err))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
