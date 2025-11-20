package service

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
)

var sessions = map[string]Session{}

var mu sync.RWMutex

type Session struct {
	UserID int64
	// TODO: other fields
}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GetSession(session_id string) (*Session, bool) {
	mu.RLock()
	s, ok := sessions[session_id]
	mu.RUnlock()

	if !ok {
		return nil, false
	}

	return &s, true
}

func CreateSession(userID int64) string {
	id := newSessionID()

	mu.Lock()
	sessions[id] = Session{
		UserID: userID,
	}
	mu.Unlock()

	return id
}

func SessionMidleware(next http.Handler) http.Handler {
	// FIX: completar el manager de sesiones
	return next
}
