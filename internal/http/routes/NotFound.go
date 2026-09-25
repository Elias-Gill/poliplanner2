package routes

import (
	"net/http"

	render "github.com/elias-gill/poliplanner2/internal/render/html"
	"github.com/elias-gill/poliplanner2/logger"
)

func NotFound(tmplMan *render.TemplateManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Path
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		if err := tmplMan.RenderPage(w, "404.html", page); err != nil {
			logger.Error("Cannot render 404 template", "error", err)
		}
	}
}
