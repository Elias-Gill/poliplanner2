package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/model/user"
)

const (
	refreshThreshold = 6 * time.Hour
	sessionExtension = 30 * time.Hour

	maxSessionDuration = 30 * 24 * time.Hour // 30 days max session duration
)

type SessionID string

type Session struct {
	ID         SessionID
	User       user.UserID
	CreatedAt  time.Time
	Expiration time.Time
	LastUse    time.Time
}

// REFACTOR: en algun momento tengo que cambiar el tema de las horas para que sea mas limpio.
// Usar todo en UTC y luego hacer las transformaciones cuando se necesario o algo asi.

// ExtendIfNeeded extends the session expiration only if the remaining time is below the
// defined threshold, preventing the session from being indefinitely prolonged by continuous
// activity.
func (s *Session) ExtendIfNeeded() {
	now := time.Now().In(timezone.ParaguayTZ)

	s.LastUse = now

	timeLeft := s.Expiration.Sub(now)
	if timeLeft < refreshThreshold {
		s.Expiration = now.Add(sessionExtension)
	}
}

func (s *Session) HasExpired() bool {
	now := time.Now().In(timezone.ParaguayTZ)

	sessionLimitReached := now.Sub(s.CreatedAt) > maxSessionDuration

	return s.Expiration.Before(now) || sessionLimitReached
}

func NewSession(userID user.UserID) *Session {
	now := time.Now().In(timezone.ParaguayTZ)
	return &Session{
		ID:         generateSessionID(),
		User:       userID,
		CreatedAt:  now,
		Expiration: now.Add(sessionExtension),
		LastUse:    now,
	}
}

func generateSessionID() SessionID {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return SessionID(base64.RawURLEncoding.EncodeToString(b))
}
