package router

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMiscRouter() func(r chi.Router) {
	layouts := template.Must(template.ParseGlob("web/templates/layout/base_layout.html"))

	// NOTE: made like this so the main layout template is parsed only one time on startup
	return func(r chi.Router) {
		r.Get("/calculator", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/calculator.html"))
			tpl.Execute(w, nil)
		})
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	layouts := template.Must(template.ParseGlob("web/templates/layout/base_layout.html"))
	w.Header().Set("Content-Type", "text/html")
	tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/404.html"))
	tpl.Execute(w, nil)
}
