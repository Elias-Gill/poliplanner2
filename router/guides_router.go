package router

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

// REFACTOR: deberia de ver la opcion de listar todas la guias directamente y parsearlas
// directo. Y de alguna manera tambien generar un robots.txt automatico
func NewGuidesRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		// Render login
		r.Get("/calculo_notas", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			execTemplateWithLayout(w, "web/templates/pages/guides/calculo_notas.html", layouts, nil)
		})

		r.Get("/about", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			execTemplateWithLayout(w, "web/templates/pages/guides/about.html", layouts, nil)
		})

		r.Get("/manual_del_bicho", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			execTemplateWithLayout(w, "web/templates/pages/guides/manual_del_bicho.html", layouts, nil)
		})

		r.Get("/news", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			execTemplateWithLayout(w, "web/templates/pages/guides/news.html", layouts, nil)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			execTemplateWithLayout(w, "web/templates/pages/guides/index.html", layouts, nil)
		})
	}
}
