package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
)

const (
	sessionExtension = 5 * time.Minute
	refreshThreshold = 5 * time.Minute // si queda menos que esto, se extiende
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
	return s.Expiration.Before(now)
}

func NewSession(userID user.UserID) *Session {
	now := time.Now().In(timezone.ParaguayTZ)
	return &Session{
		ID:         generateSessionID(),
		User:       userID,
		CreatedAt:  now,
		Expiration: now.Add(time.Minute * 30), // 30 minutes session duration
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
