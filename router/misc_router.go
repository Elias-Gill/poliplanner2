package router

import (
	"html/template"
	"net/http"

	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewMiscRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/misc/index.html"))
			tpl.Execute(w, nil)
		})

		r.Get("/calculator", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/misc/calculator.html"))
			tpl.Execute(w, nil)
		})
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tpl := template.Must(template.Must(web.BaseLayout.Clone()).ParseFiles("web/templates/pages/404.html"))
	tpl.Execute(w, nil)
}
