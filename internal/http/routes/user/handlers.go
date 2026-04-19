package user

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/app/auth"
	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/middleware"
)

type userHandlers struct {
	auth *auth.AuthManager
}

func newUserHandlers(authManager *auth.AuthManager) *userHandlers {
	return &userHandlers{
		auth: authManager,
	}
}

// ==================== Handlers ====================

func (h *userHandlers) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	userID := middleware.MustExtractUserID(r)

	cookie, err := r.Cookie(middleware.SessionIdCookie)
	if err == nil {
		h.auth.Logout(r.Context(), userID, auth.SessionID(cookie.Value))
	}

	// Always invalidate the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionIdCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	utils.CustomRedirect(w, r, "/login")
}
