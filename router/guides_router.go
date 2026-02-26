package router

import (
	"net/http"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/go-chi/chi/v5"
)

// REFACTOR: deberia de hacer algun metodo para cargar guias automaticamente
func NewGuidesRouter() func(r chi.Router) {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "guides")

	// pre-compile templates
	calcGuideTemplate := parseTemplateWithBaseLayout(path.Join(baseDir, "calculo_notas.html"))
	aboutTemplate := parseTemplateWithBaseLayout(path.Join(baseDir, "about.html"))
	manualBichoTemplate := parseTemplateWithBaseLayout(path.Join(baseDir, "manual_del_bicho.html"))
	newsTemplate := parseTemplateWithBaseLayout(path.Join(baseDir, "news.html"))
	indexTemplate := parseTemplateWithBaseLayout(path.Join(baseDir, "index.html"))

	return func(r chi.Router) {
		r.Get("/calculo_notas", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			calcGuideTemplate.Execute(w, nil)
		})

		r.Get("/about", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			aboutTemplate.Execute(w, nil)
		})

		r.Get("/manual_del_bicho", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			manualBichoTemplate.Execute(w, nil)
		})

		r.Get("/news", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			newsTemplate.Execute(w, nil)
		})

		// Guides index
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			indexTemplate.Execute(w, nil)
		})
	}
}
