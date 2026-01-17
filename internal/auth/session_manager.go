package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// ================================
// =         Public API           =
// ================================

type Session struct {
	UserID int64
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

// CreateSession generates a JWT token containing the userID and expiration time
func CreateSession(userID int64) string {
	expirationTime := time.Now().Add(30 * time.Minute)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Get().Security.UpdateKey))
	if err != nil {
		// Panic or handle error properly if token generation fails
		panic("Error generating JWT: " + err.Error())
	}

	return tokenString
}

// getSession validates the JWT token and returns the session containing session info
func getSession(tokenString string) (*Session, bool) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Verify the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenMalformed
		}
		return []byte(config.Get().Security.UpdateKey), nil
	})

	if err != nil || !token.Valid {
		logger.Debug("Invalid or expired JWT", "error", err)
		return nil, false
	}

	return &Session{UserID: claims.UserID}, true
}

func customRedirect(w http.ResponseWriter, r *http.Request, target string) {
	isHtmx := r.Header.Get("HX-Request") == "true"
	if isHtmx {
		w.Header().Add("HX-redirect", target)
	} else {
		http.Redirect(w, r, target, http.StatusFound)
	}
}
