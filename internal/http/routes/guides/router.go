package guides

import (
	"github.com/go-chi/chi/v5"
)

func NewGuidesRouter() func(r chi.Router) {
	handlers := newGuidesHandlers()

	return func(r chi.Router) {
		r.Get("/calculo_notas", handlers.CalculoNotas)
		r.Get("/about", handlers.About)
		r.Get("/manual_del_bicho", handlers.ManualDelBicho)
		r.Get("/news", handlers.News)
		r.Get("/", handlers.Index)
	}
}
