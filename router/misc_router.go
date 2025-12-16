package router

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewMiscRouter() func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			execTemplateWithLayout(w, "web/templates/pages/misc/index.html", layouts, nil)
		})

		r.Get("/calculator", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			execTemplateWithLayout(w, "web/templates/pages/misc/calculator.html", layouts, nil)
		})
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	execTemplateWithLayout(w, "web/templates/pages/404.html", web.BaseLayout, nil)
}
