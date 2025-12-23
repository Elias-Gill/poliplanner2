package router

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/auth"
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

		r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
			auth.DeleteSession(r)
			customRedirect(w, r, "/")
		})
	}
}
