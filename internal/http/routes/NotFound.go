package routes

import (
	"github.com/elias-gill/poliplanner2/internal/http/render"
	"net/http"
)

func NotFound(tmplMan *render.TemplateManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Path
		w.Header().Set("Content-Type", "text/html")
		tmplMan.RenderPage(w, "404.html", page)
	}
}
