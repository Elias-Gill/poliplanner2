package tools

import (
	"html/template"
	"net/http"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
	utils "github.com/elias-gill/poliplanner2/internal/http"
)

type ToolsHandlers struct {
	indexTemplate            *template.Template
	calculatorTemplate       *template.Template
	interactiveGraphTemplate *template.Template
}

func newToolsHandlers() *ToolsHandlers {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "tools")

	return &ToolsHandlers{
		indexTemplate:            utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "index.html")),
		calculatorTemplate:       utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "calculator.html")),
		interactiveGraphTemplate: utils.ParseTemplateWithBaseLayout(path.Join(baseDir, "interactive_graph.html")),
	}
}

// ==================== Handlers ====================

func (h *ToolsHandlers) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	h.indexTemplate.Execute(w, nil)
}

func (h *ToolsHandlers) Calculator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	h.calculatorTemplate.Execute(w, nil)
}

func (h *ToolsHandlers) InteractiveGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	h.interactiveGraphTemplate.Execute(w, nil)
}
