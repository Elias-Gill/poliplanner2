package tools

import (
	"github.com/elias-gill/poliplanner2/internal/http/render"
	"github.com/go-chi/chi/v5"

	"net/http"

	"github.com/elias-gill/poliplanner2/logger"
)

type Handler struct {
	tmpl *render.TemplateManager
}

func NewHandler(tmpl *render.TemplateManager) *Handler {
	return &Handler{
		tmpl: tmpl,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.index)
	r.Get("/calculator", h.calculator)
	r.Get("/interactive_graph", h.interactiveGraph)

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/index.html", nil); err != nil {
		logger.Error("Cannot render tools index template", "error", err)
	}
}

func (h *Handler) calculator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/calculator.html", nil); err != nil {
		logger.Error("Cannot render calculator template", "error", err)
	}
}

func (h *Handler) interactiveGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.tmpl.RenderPage(w, "tools/interactive_graph.html", nil); err != nil {
		logger.Error("Cannot render interactive_graph template", "error", err)
	}
}
