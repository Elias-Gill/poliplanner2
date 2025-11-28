package router

import (
	"html/template"
	"net/http"

	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewGuidesRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		// Render login
		r.Get("/calculo_notas", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/guides/calculo_notas.html"))
			tpl.Execute(w, nil)
		})

		r.Get("/about", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/guides/about.html"))
			tpl.Execute(w, nil)
		})

		r.Get("/manual_del_bicho", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/guides/manual_del_bicho.html"))
			tpl.Execute(w, nil)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/guides/index.html"))
			tpl.Execute(w, nil)
		})
	}
}
