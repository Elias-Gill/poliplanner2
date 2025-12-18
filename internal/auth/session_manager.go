package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/elias-gill/poliplanner2/internal/logger"
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
		panic(fmt.Sprintf("Cannot generate secure session IDs: %+v", err))
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

// SessionMiddleware is an HTTP middleware that verifies whether the incoming request
// is associated with an authenticated user.
//
// Authentication is based on a session cookie that stores a session identifier.
// When a request reaches this middleware, the session ID is extracted from the cookie
// and validated against the sessions table.
//
// If the session is valid, the user ID associated with that session is injected into
// the request context under the key "userID", making it available to downstream handlers.
func SessionMiddleware(next http.Handler) http.Handler {
	protected := []string{
		"/dashboard",
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

		loginPage := "/login?redirect=" + url.QueryEscape(r.URL.RequestURI())

		// Validate session. If invalid, redirects to the login page.
		cookie, err := r.Cookie("session_id")
		if err != nil {
			logger.Debug("Session middleware redirection", "cause", "cookie not present")
			customRedirect(w, r, loginPage)
			return
		}

		session, ok := getSession(cookie.Value)
		if !ok {
			logger.Debug("Session middleware redirection", "cause", "session expired or invalid token")
			customRedirect(w, r, loginPage)
			return
		}

		logger.Debug("User already authenticated")
		ctx := context.WithValue(r.Context(), "userID", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func customRedirect(w http.ResponseWriter, r *http.Request, target string) {
	isHtmx := r.Header.Get("HX-Request") == "true"
	if isHtmx {
		w.Header().Add("HX-redirect", target)
	} else {
		http.Redirect(w, r, target, http.StatusFound)
	}
}
