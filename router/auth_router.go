package router

import (
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/auth"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewAuthRouter(service *service.UserService) func(r chi.Router) {
	layouts := web.CleanLayout

	return func(r chi.Router) {
		// Redirect to the dashboard.
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		})

		// --- Handle LOGIN ---
		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			// Set redirection parameter if a redirection path is present
			data := map[string]any{
				"Redirect": r.URL.Query().Get("redirect"),
			}

			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/login.html"))
			tpl.Execute(w, data)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			// Authentication
			username := r.FormValue("username")
			password := r.FormValue("password")
			user, err := service.AuthenticateUser(r.Context(), username, password)
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(newErrorFragment("Usuario o contraseña incorrectos")))
				return
			}

			// Create a session cookie
			sessionID := auth.CreateSession(user.ID)
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   config.Get().Security.SecureHTTP,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().Add(30 * time.Minute),
			})

			// Redirect if a redirection path is present
			redirectTo := r.URL.Query().Get("redirect")
			if len(redirectTo) == 0 {
				redirectTo = "/dashboard"
			}

			logger.Debug("Redirecting from login", "path", redirectTo)

			w.Header().Set("HX-Redirect", redirectTo)
		})

		// --- Handle SIGNUP ---
		r.Get("/signup", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("web/templates/pages/signup.html"))
			tpl.Execute(w, nil)
		})

		r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			email := r.FormValue("email")
			rawPassword := r.FormValue("password")

			if err := validateUsername(username); err != nil {
				w.Write([]byte(newErrorFragment(err.Error())))
				return
			}

			if !isValidEmail(email) {
				w.Write([]byte(newErrorFragment("Invalid email provided")))
				return
			}

			if len(rawPassword) < 6 {
				w.Write([]byte(newErrorFragment("Password must have at least 6 characters")))
				return
			}

			err := service.CreateUser(r.Context(), strings.ToLower(username), email, rawPassword)
			if err != nil {
				w.Write([]byte(newErrorFragment(err.Error())))
				return
			}

			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusOK)
		})
	}
}

func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("El nombre de usuario debe tener al menos 3 caracteres")
	}
	if !isAlphanumeric(username) {
		return fmt.Errorf("El nombre de usuario solo puede contener letras, numeros y _")
	}
	return nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isAlphanumeric(str string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	return re.MatchString(str)
}

func newErrorFragment(msg string) string {
	return `
	<section role="alert" class="error">
		<span>` + msg + `</span>
		<button type="button" style="background-color: var(--color-error);" onclick="this.parentElement.remove()" aria-label="Cerrar alerta">×</button>
	</section>
	`
}
