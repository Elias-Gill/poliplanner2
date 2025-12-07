package router

import (
	"net/http"
	"strings"

	// "github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewExcelRouter(clave string) func(r chi.Router) {
	// layouts := web.CleanLayout

	// FIX: hacer mas robusto
	// private final String expectedKey = System.getenv("UPDATE_KEY");
	//
	return func(r chi.Router) {
		// Este endpoint al recibir una request, tratara de scrapear la web de la
		// universidad en busca de nuevos horarios. Require del header: "Authorization: Bearer <key>"
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !isValidRequest(header, clave) {
				http.Error(w, "Invalid credentials", http.StatusForbidden)
			}

		})
	}
}

// FIX: revisar
func isValidRequest(authHeader string, expectedKey string) bool {
	if len(authHeader) == 0 {
		return false
	}

	if strings.Trim(authHeader, "\n \t") != ("Bearer " + expectedKey) {
		return false
	}

	return true
}
