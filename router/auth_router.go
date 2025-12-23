package router

import (
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
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
			execTemplateWithLayout(w, "web/templates/pages/login.html", layouts, data)
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

			// Generate a new session
			sessionID := auth.CreateSession(user.ID)

			// Set the session cookie
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
			execTemplateWithLayout(w, "web/templates/pages/signup.html", layouts, nil)
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

			err := service.CreateUser(r.Context(), username, email, rawPassword)
			if err != nil {
				w.Write([]byte(newErrorFragment(err.Error())))
				return
			}

			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusOK)
		})

		r.Get("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			execTemplateWithLayout(w, "web/templates/pages/auth/password-recovery.html", layouts, nil)
		})

		r.Post("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(newErrorFragment("Error al parsear el form")))
				return
			}

			email := strings.TrimSpace(r.Form.Get("email"))
			if email == "" {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(newErrorFragment("El email no puede estar vacío")))
				return
			}

			if !isValidEmail(email) {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(newErrorFragment("Email inválido")))
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*300)
			defer cancel()

			token, err := userService.StartPasswordRecovery(ctx, email)
			if err != nil {
				if errors.Is(err, service.CannotFindUserError) {
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte(newSuccessFragment("Se ha enviado un link de recuperación al correo proporcionado.")))
					return
				} else {
					customRedirect(w, r, "/500")
					return
				}
			}

			err = emailService.SendRecoveryEmail(email, token)
			if err != nil {
				customRedirect(w, r, "/500")
				logger.Error("Error sending recovery email", "error", err)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(newSuccessFragment("Se ha enviado un link de recuperación al correo proporcionado.")))
		})

		r.Get("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if strings.TrimSpace(token) == "" {
				customRedirect(w, r, "/500")
				return
			}

			data := map[string]any{
				"Token": token,
			}

			execTemplateWithLayout(w, "web/templates/pages/auth/password-recovery-commit.html", layouts, data)
		})

		r.Post("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if strings.TrimSpace(token) == "" {
				w.Write([]byte(newErrorFragment("Token inválido")))
				return
			}

			if err := r.ParseForm(); err != nil {
				w.Write([]byte(newErrorFragment("Error al procesar el formulario")))
				return
			}

			password := r.Form.Get("password")
			confirm := r.Form.Get("confirm_password")

			if password == "" || confirm == "" {
				w.Write([]byte(newErrorFragment("La contraseña no puede estar vacía")))
				return
			}

			if password != confirm {
				w.Write([]byte(newErrorFragment("Las contraseñas no coinciden")))
				return
			}

			if len(password) < 6 {
				w.Write([]byte(newErrorFragment("La contraseña debe tener al menos 6 caracteres")))
				return
			}

			err := userService.CommitPasswordRecovery(r.Context(), token, password)
			if err != nil {
				w.Write([]byte(newErrorFragment("El enlace es inválido o ya expiró")))
				return
			}

			w.Write([]byte(`
				<section class="success">
				<span>Contraseña actualizada correctamente</span>
				</section>`))
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

func newSuccessFragment(msg string) string {
	return `
	<section role="alert" class="success">
	<span>` + msg + `</span>
	<button type="button" style="background-color: var(--color-success);" onclick="this.parentElement.remove()" aria-label="Cerrar alerta">×</button>
	</section>
	`
}
