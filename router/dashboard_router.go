package router

import (
	"github.com/go-chi/chi/v5"
	// "html/template"
	"net/http"
)

func NewDashboardRouter() func(r chi.Router) {
	// layouts := template.Must(template.ParseGlob("web/templates/layout/base_layout.html"))

	// NOTE: made like this so the main layout template is parsed only one time on startup
	return func(r chi.Router) {
		// Dashboard layout
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// TODO: continue
			w.Write([]byte("<h1>Hola, bienvenido a poliplanner</h1>"))
		})
	}
}
