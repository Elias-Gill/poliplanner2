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

	/*
		* Funciona exactamente igual que el endpoint "/sync", pero esta pensado para
		* automatizacion con web scrapping.
		*
		* Este endpoint al recibir una request, tratara de scrapear la web de la
		* universidad en busca de nuevos horarios.
		*
		* Require del header: "Authorization: Bearer <key>"
			@PostMapping("/sync/ci")
			public ResponseEntity<?> automaticExcelSync(@RequestHeader("Authorization") String authHeader) {
				logger.warn(">>> POST '/sync/ci' alcanzado desde CI/CD");
				try {
				if (!tokenValidator.isValid(authHeader)) {
					return ResponseEntity.status(HttpStatus.FORBIDDEN).body("Credenciales invalidas");
				}

				Boolean hasNewVersion = service.autonomousExcelSync();
				if (hasNewVersion) {
					return ResponseEntity.status(HttpStatus.OK)
					.body("Version de excel actualizada a la nueva version disponible");
				}

				return ResponseEntity.status(HttpStatus.OK)
				.body("Excel ya se encuentra en su ultima version");
			} catch (Exception e) {
			logger.error("Error al sincronizar Excel", e);
			return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
			.body("No se pudo sincronizar: " + e.getMessage());
		}
	*/
}

func isValidRequest(authHeader string, expectedKey string) bool {
	if len(authHeader) == 0 {
		return false
	}

	if strings.Trim(authHeader, "\n \t") != ("Bearer " + expectedKey) {
		return false
	}

	return true
}
