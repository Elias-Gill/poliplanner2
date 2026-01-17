package router

import (
	"net/http"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewUserRouter(userService *service.UserService) func(r chi.Router) {
	layout := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			execTemplateWithLayout(w, "web/templates/pages/user/index.html", layout, nil)
		})

		// FIX: This just invalidates the cookie, proper JWT token invalidation and blacklist
		// is not implemented.
		r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   config.Get().Security.SecureHTTP,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Unix(0, 0), // Set expiration to a past date
			})

			customRedirect(w, r, "/")
		})
	}
}
