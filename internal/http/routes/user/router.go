package user

import (
	render "github.com/elias-gill/poliplanner2/internal/render/html"
	"github.com/elias-gill/poliplanner2/internal/service/auth"
	"github.com/go-chi/chi/v5"

	"net/http"

	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/cookie"
	authModel "github.com/elias-gill/poliplanner2/internal/model/auth"
)

type Handler struct {
	tmpl *render.TemplateManager
	auth *auth.SessionService
}

func NewHandler(
	tmpl *render.TemplateManager,
	authManager *auth.SessionService,
) *Handler {
	return &Handler{
		tmpl: tmpl,
		auth: authManager,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/logout", h.logout)

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	userID := utils.MustExtractUserID(r)

	sessionCookie, err := cookie.GetSessionCookie(r)
	if err == nil {
		sessionToken := authModel.SessionID(sessionCookie)

		// Logout server side
		h.auth.Logout(
			r.Context(),
			userID,
			sessionToken,
		)
	}

	// Clear the session cookie to invalidate the session client side
	cookie.ClearSessionCookie(w)

	utils.Redirect(w, r, "/login")
}
