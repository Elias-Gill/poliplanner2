package router

import (
	"net/http"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/go-chi/chi/v5"
)

func NewMiscRouter() func(r chi.Router) {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages")

	// templates paths
	indexPath := path.Join(baseDir, "misc", "index.html")
	calculatorPath := path.Join(baseDir, "misc", "calculator.html")

	indexTemplate := parseTemplateWithBaseLayout(indexPath)
	calculatorTemplate := parseTemplateWithBaseLayout(calculatorPath)

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			indexTemplate.Execute(w, nil)
		})

		r.Get("/calculator", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			calculatorTemplate.Execute(w, nil)
		})
	}
}

// REFACTOR: hacer para que tenga cierta cache como el resto de enpoints, pero no es urgente
// para nada
//
// NOTE: ciertamente el mecanismo simple de cache funciona, pero a la larga puede generar
// demasiados objetos de templates para paginas que no se usan. Hay que tener un criterio a la
// hora de usar para los endpoints mas consultados y/o implementar ciertos mecanismos de lazy
// load si es que parsear estas templates requiere de exceso de ciclos de CPU (cosa que no
// realmente).
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages")
	parseTemplateWithBaseLayout(path.Join(baseDir, "404.html")).Execute(w, nil)
	w.Header().Set("Content-Type", "text/html")
}
