package middleware

import (
	"net/http"
	"net/url"
	"strings"

	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/cookie"
	authModel "github.com/elias-gill/poliplanner2/internal/model/auth"
	authSrv "github.com/elias-gill/poliplanner2/internal/service/auth"
	"github.com/elias-gill/poliplanner2/logger"
)

var ProtectedRoutes = []string{
	"/dashboard",
	"/schedule",
	"/user",
}

// SessionMiddleware verifies session authentication for protected routes.
func NewSessionMiddleware(authManager *authSrv.SessionService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !isProtectedRoute(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			loginPage := buildLoginRedirect(r)

			// If session cookie is not present, redirect to the login page
			cookie, err := r.Cookie(cookie.SessionIDCookie)
			if err != nil {
				logger.Debug("session middleware redirect", "cause", "cookie not present")
				utils.Redirect(w, r, loginPage)
				return
			}

			// If present, then authenticate the session
			session, err := authManager.ValidateSession(
				r.Context(),
				authModel.SessionID(cookie.Value),
			)

			// If session is not authenticated, then redirect to login page
			if err != nil {
				logger.Debug(
					"session middleware redirect",
					"cause", "invalid session",
					"error", err,
				)
				utils.Redirect(w, r, loginPage)
				return
			}

			next.ServeHTTP(w, utils.InjectUserID(r, session.User))
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
