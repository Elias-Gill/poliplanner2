package news

import (
	"net/http"

	render "github.com/elias-gill/poliplanner2/internal/render/html"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"

	utils "github.com/elias-gill/poliplanner2/internal/http"
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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.Redirect(w, r, "/")
	})

	r.Get("/limpieza_secciones_fantasma", h.limpieza)

	return r
}

func (h *Handler) limpieza(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.RenderPage(w, "news/limpieza_secciones_fantasma.html", nil); err != nil {
		logger.Error("Cannot render news template", "error", err)
	}
}
