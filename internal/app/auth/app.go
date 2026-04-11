package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"github.com/elias-gill/poliplanner2/logger"
)

type AuthManager struct {
	userStorer    user.UserRepository
	sessionStorer SessionRepository
}

func New(userStore user.UserRepository, sessionStore SessionRepository) *AuthManager {
	return &AuthManager{
		userStorer:    userStore,
		sessionStorer: sessionStore,
	}
}

// ================================
// =         Public API           =
// ================================

var (
	ErrSessionExpired      = errors.New("session has expired")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionSaveFailed   = errors.New("failed to save session")
	ErrSessionDeleteFailed = errors.New("failed to delete session")
	ErrSessionStoreFailure = errors.New("session store failure")
	ErrUserStoreFailure    = errors.New("user store failure")
)

// Login authenticates a user and creates a new session.
func (a *AuthManager) Login(ctx context.Context, login string, rawPassword string) (*Session, error) {
	login = normalizeLogin(login)

	u, err := a.getUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if err := u.AuthenticatePassword(rawPassword); err != nil {
		return nil, user.ErrInvalidCredentials
	}

	s := NewSession(u.ID)

	if err := a.sessionStorer.Save(ctx, s); err != nil {
		logger.Error("Session store failure", "error", err)
		return nil, errors.Join(ErrSessionSaveFailed, err)
	}

	return s, nil
}

// ValidateSession retrieves and validates an existing session.
func (a *AuthManager) ValidateSession(ctx context.Context, token SessionID) (*Session, error) {
	s, err := a.sessionStorer.Get(ctx, token)
	if err != nil {
		return nil, errors.Join(ErrSessionStoreFailure, err)
	}

	if s == nil {
		return nil, ErrSessionNotFound
	}

	if s.HasExpired() {
		if err := a.sessionStorer.Delete(ctx, s.ID); err != nil {
			return nil, errors.Join(ErrSessionExpired, ErrSessionDeleteFailed, err)
		}

		return nil, ErrSessionExpired
	}

	s.ExtendIfNeeded()

	if err := a.sessionStorer.Save(ctx, s); err != nil {
		return nil, errors.Join(ErrSessionSaveFailed, err)
	}

	return s, nil
}

// Logout invalidates a session by removing it from the store.
func (a *AuthManager) Logout(ctx context.Context, token SessionID) error {
	if err := a.sessionStorer.Delete(ctx, token); err != nil {
		return errors.Join(ErrSessionDeleteFailed, err)
	}
	return nil
}

// ================================
// =        Private Helpers       =
// ================================

// normalizeLogin standardizes login input (username or email).
func normalizeLogin(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

// getUserByLogin attempts to retrieve a user by username first, then by email.
// Separates infrastructure failures from invalid credentials.
func (a *AuthManager) getUserByLogin(ctx context.Context, login string) (*user.User, error) {
	u, err := a.userStorer.GetByUsername(ctx, login)
	if err == nil {
		return u, nil
	}

	if !errors.Is(err, user.ErrUserNotFound) {
		logger.Error("User store failure", "error", err)
		return nil, errors.Join(ErrUserStoreFailure, err)
	}

	u, err = a.userStorer.GetByEmail(ctx, login)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, user.ErrInvalidCredentials
		}
		return nil, errors.Join(ErrUserStoreFailure, err)
	}

	return u, nil
}
