package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

type AuthManager struct {
	userStorer    user.UserStorer
	sessionStorer SessionStorer
}

func NewAuthManager(userStore user.UserStorer, sessionStore SessionStorer) *AuthManager {
	return &AuthManager{
		userStorer:    userStore,
		sessionStorer: sessionStore,
	}
}

// ================================
// =         Public API           =
// ================================

// Errores personalizados
var (
	ErrSessionExpired    = errors.New("session has expired")
	ErrSessionSaveFailed = errors.New("failed to save session")
	ErrSessionNotFound   = errors.New("session not found")
)

func (a *AuthManager) AuthenticateUser(ctx context.Context, login string, rawPassword string) (*Session, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	u, err := a.userStorer.GetByUsername(ctx, login)
	if err != nil {
		u, err = a.userStorer.GetByEmail(ctx, login)
		if err != nil {
			return nil, user.ErrInvalidCredentials
		}
	}

	if err := u.AuthenticatePassword(rawPassword); err != nil {
		return nil, user.ErrInvalidCredentials
	}

	s := NewSession(u.ID)

	if err := a.sessionStorer.Save(ctx, s); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSessionSaveFailed, err)
	}

	return s, nil
}

func (a *AuthManager) AuthenticateSession(ctx context.Context, token SessionID) (*Session, error) {
	s, err := a.sessionStorer.Get(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSessionNotFound, err)
	}

	if s.HasExpired() {
		// FIX: error handling
		// Invalidate session
		_ = a.sessionStorer.Delete(ctx, s.ID)
		return nil, ErrSessionExpired
	}

	s.ExtendIfNeeded()

	if err := a.sessionStorer.Save(ctx, s); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSessionSaveFailed, err)
	}

	return s, nil
}
