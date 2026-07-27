package guides

import (
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/http/render"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
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

	r.Get("/calculo_notas", h.calculoNotas)
	r.Get("/about", h.about)
	r.Get("/manual_del_bicho", h.manualDelBicho)
	r.Get("/news", h.news)
	r.Get("/", h.index)

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) calculoNotas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "guides/calculo_notas.html", nil); err != nil {
		logger.Error("Cannot render calculo_notas template", "error", err)
	}
}

func (h *Handler) about(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "guides/about.html", nil); err != nil {
		logger.Error("Cannot render about template", "error", err)
	}
}

func (h *Handler) manualDelBicho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "guides/manual_del_bicho.html", nil); err != nil {
		logger.Error("Cannot render manual_del_bicho template", "error", err)
	}
}

func (h *Handler) news(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "guides/news.html", nil); err != nil {
		logger.Error("Cannot render news template", "error", err)
	}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "guides/index.html", nil); err != nil {
		logger.Error("Cannot render guides index template", "error", err)
	}
}
