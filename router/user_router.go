package router

import (
	"net/http"
	"path"

	"github.com/elias-gill/poliplanner2/internal/app/auth"
	"github.com/elias-gill/poliplanner2/internal/app/user"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/go-chi/chi/v5"
)

// REFACTOR: si es que me molesta puedo cambiar el nombre y locacion de los endpoints,
// pero de momento me parece aceptable esta porqueria
func NewUserRouter(userService *user.UserService, authService *auth.AuthManager) func(r chi.Router) {
	p := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "user", "index.html")
	tpl := parseTemplateWithBaseLayout(p)

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tpl.Execute(w, nil)
		})

		r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			// TODO: Invalidate cookie
			// FIX: logout with method

			customRedirect(w, r, "/")
		})
	}
}
