package router

import (
	"net/http"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewExcelRouter() func(r chi.Router) {
	key := config.Get().UPDATE_KEY

	return func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !isValidRequest(header, key) {
				http.Error(w, "invalid credentials", http.StatusForbidden)
				return
			}

			err := service.SearchNewestExcel(r.Context())
			if err != nil {
				http.Error(w, "Cannot parse excel: "+err.Error(), http.StatusForbidden)
				return
			}

			w.WriteHeader(http.StatusOK)
		})
	}
}

func isValidRequest(authHeader, expectedKey string) bool {
	if authHeader == "" {
		return false
	}

	authHeader = strings.TrimSpace(authHeader)
	expected := "Bearer " + expectedKey

	return authHeader == expected
}
