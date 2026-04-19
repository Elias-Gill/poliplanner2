package user

import (
	"github.com/elias-gill/poliplanner2/internal/app/auth"
	"github.com/go-chi/chi/v5"
)

func NewUserRouter(auth *auth.AuthManager) func(r chi.Router) {
	handlers := newUserHandlers(auth)

	return func(r chi.Router) {
		r.Post("/logout", handlers.logout)
	}
}
