package tools

import (
	"github.com/go-chi/chi/v5"
)

func NewToolsRouter() func(r chi.Router) {
	handlers := newToolsHandlers()

	return func(r chi.Router) {
		r.Get("/", handlers.Index)
		r.Get("/calculator", handlers.Calculator)
		r.Get("/interactive_graph", handlers.InteractiveGraph)
	}
}
