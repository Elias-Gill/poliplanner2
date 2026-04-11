package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/app/auth"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/logger"
)

const (
	UserIDKey = "userID"

	SessionIdCookie = "session_id"
)

var ProtectedRoutes = []string{
	"/dashboard",
	"/schedule",
	"/user",
}

// SessionMiddleware verifies session authentication for protected routes.
func NewSessionMiddleware(authManager *auth.AuthManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !isProtectedRoute(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			loginPage := buildLoginRedirect(r)

			// If session cookie is not present, redirect to the login page
			cookie, err := r.Cookie(SessionIdCookie)
			if err != nil {
				logger.Debug("session middleware redirect", "cause", "cookie not present")
				utils.CustomRedirect(w, r, loginPage)
				return
			}

			// If present, then authenticate the session
			session, err := authManager.ValidateSession(
				r.Context(),
				auth.SessionID(cookie.Value),
			)

			// If session is not authenticated, then redirect to login page
			if err != nil {
				logger.Debug(
					"session middleware redirect",
					"cause", "invalid session",
					"error", err,
				)
				utils.CustomRedirect(w, r, loginPage)
				return
			}

			next.ServeHTTP(w, injectUserIntoContext(r, session.User))
		})
	}
}

// isProtectedRoute checks if a route requires authentication.
func isProtectedRoute(path string) bool {
	for _, p := range ProtectedRoutes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// buildLoginRedirect builds the login redirect URL preserving the requested path.
func buildLoginRedirect(r *http.Request) string {
	return "/login?redirect=" + url.QueryEscape(r.URL.RequestURI())
}

// injectUserIntoContext adds the authenticated user to the request context.
func injectUserIntoContext(r *http.Request, userID user.UserID) *http.Request {
	ctx := context.WithValue(r.Context(), "userID", userID)
	return r.WithContext(ctx)
}

// This function should never fail or panic if the session middleware is functioning correctly.
// If a protected endpoint is reached without a userID set in the request context,
// the application is in an invalid state and something unexpected has occurred.
//
// If this is the case, then probably the endpoint has not been added to the "protected
// endpoints" array list in the middleware configuration.
func MustExtractUserID(r *http.Request) user.UserID {
	switch id := r.Context().Value(UserIDKey).(type) {
	case user.UserID:
		return id
	default:
		panic("The user ID is not set in the request of a protected endpoint, something wrong must have happened with the session manager middleware")
	}
}
