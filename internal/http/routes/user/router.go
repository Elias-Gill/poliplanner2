package user

import (
	render "github.com/elias-gill/poliplanner2/internal/render/html"
	"github.com/elias-gill/poliplanner2/internal/service/auth"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"

	"net/http"

	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/cookie"
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

	r.Get("/", h.index)

	r.Post("/logout", h.logout)

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "user/index.html", nil); err != nil {
		logger.Error("Cannot render user_index template", "error", err)
	}
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// NOTE: user session is garanted by the session middleware to exist for
	// protected endpoints.
	session := utils.MustExtractUserSession(r)
	h.auth.Logout(r.Context(), session.ID)

	// Clear the session cookie to invalidate the session client side anyways
	cookie.ClearSessionCookie(w)

	utils.Redirect(w, r, "/login")
}
