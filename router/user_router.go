package router

import (
	"net/http"
	"path"
	"time"

	"github.com/elias-gill/poliplanner2/internal/app/user"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/go-chi/chi/v5"
)

// REFACTOR: si es que me molesta puedo cambiar el nombre y locacion de los endpoints,
// pero de momento me parece aceptable esta porqueria
func NewUserRouter(userService *user.UserService) func(r chi.Router) {
	p := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "user", "index.html")
	tpl := parseTemplateWithBaseLayout(p)

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tpl.Execute(w, nil)
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
