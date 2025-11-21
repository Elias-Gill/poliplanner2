package router

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewGuidesRouter() func(r chi.Router) {
	layouts := template.Must(template.ParseGlob("templates/layout/base_layout.html"))

	// NOTE: made like this so the main layout template is parsed only one time on startup
	return func(r chi.Router) {
		// Render login
		r.Get("/calculo_notas", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/pages/guides/calculo_notas.html"))
			tpl.Execute(w, nil)
		})

		r.Get("/about", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/pages/guides/about.html"))
			tpl.Execute(w, nil)
		})

		r.Get("/manual_del_bicho", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/pages/guides/manual_del_bicho.html"))
			tpl.Execute(w, nil)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/pages/guides/index.html"))
			tpl.Execute(w, nil)
		})
	}
}
