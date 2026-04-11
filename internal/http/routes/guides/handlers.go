package guides

import (
	"html/template"
	"net/http"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
	utils "github.com/elias-gill/poliplanner2/internal/http"
)

type GuidesHandlers struct {
	calcGuideTemplate   *template.Template
	aboutTemplate       *template.Template
	manualBichoTemplate *template.Template
	newsTemplate        *template.Template
	indexTemplate       *template.Template
}

func newGuidesHandlers() *GuidesHandlers {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "guides")

	return &GuidesHandlers{
		calcGuideTemplate:   utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "calculo_notas.html")),
		aboutTemplate:       utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "about.html")),
		manualBichoTemplate: utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "manual_del_bicho.html")),
		newsTemplate:        utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "news.html")),
		indexTemplate:       utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "index.html")),
	}
}

// ==================== Handlers ====================

func (h *GuidesHandlers) CalculoNotas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.calcGuideTemplate.Execute(w, nil)
}

func (h *GuidesHandlers) About(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.aboutTemplate.Execute(w, nil)
}

func (h *GuidesHandlers) ManualDelBicho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.manualBichoTemplate.Execute(w, nil)
}

func (h *GuidesHandlers) News(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.newsTemplate.Execute(w, nil)
}

func (h *GuidesHandlers) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.indexTemplate.Execute(w, nil)
}
