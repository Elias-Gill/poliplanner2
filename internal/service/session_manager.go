package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var (
	mu       sync.RWMutex
	sessions = make(map[string]*Session)
)

type Session struct {
	UserID int64
}

func newSessionID() string {
	b := make([]byte, 32) // 256-bit random session ID
	_, err := rand.Read(b)
	if err != nil {
		panic("Cannot generate secure session IDs")
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func getSession(sessionID string) (*Session, bool) {
	mu.RLock()
	s, ok := sessions[sessionID]
	mu.RUnlock()
	return s, ok
}

func CreateSession(userID int64) string {
	id := newSessionID()

	mu.Lock()
	sessions[id] = &Session{UserID: userID}
	mu.Unlock()

	return id
}

// HTTP middleware setting the user id on the request context
func SessionMiddleware(next http.Handler) http.Handler {
	protected := []string{
		"/schedule",
		"/user",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routePath := r.URL.Path

		// Determine if path is protected
		needsAuth := false
		for _, p := range protected {
			if strings.HasPrefix(routePath, p) {
				needsAuth = true
				break
			}
		}

		if !needsAuth {
			next.ServeHTTP(w, r)
			return
		}

		// Validate session. If invalid, redirects to the login page.
		cookie, err := r.Cookie("session_id")
		if err != nil {
			target := url.QueryEscape(r.URL.RequestURI())
			http.Redirect(w, r, "/login?redirect="+target, http.StatusTemporaryRedirect)
			return
		}
		session, ok := getSession(cookie.Value)
		if !ok {
			target := url.QueryEscape(r.URL.RequestURI())
			http.Redirect(w, r, "/login?redirect="+target, http.StatusTemporaryRedirect)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
