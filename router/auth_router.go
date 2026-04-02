package router

import (
	"errors"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/app/email"
	userApp "github.com/elias-gill/poliplanner2/internal/app/user"
	"github.com/elias-gill/poliplanner2/internal/auth"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	userDomain "github.com/elias-gill/poliplanner2/internal/domain/user"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

func NewAuthRouter(userService *userApp.UserService, emailService *email.EmailService) func(r chi.Router) {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages")

	// templates paths
	template500Path := path.Join(baseDir, "500.html")
	templateBadFormPath := path.Join(baseDir, "bad_form.html")
	templateLoginPath := path.Join(baseDir, "auth", "login.html")
	templateSignupPath := path.Join(baseDir, "auth", "signup.html")
	templateRecoveryPath := path.Join(baseDir, "auth", "password-recovery.html")
	templateRecoveryCommitPath := path.Join(baseDir, "auth", "password-recovery-commit.html")

	// parse templates
	pages500Template := parseTemplateWithBaseLayout(template500Path)
	pagesBadFormTemplate := parseTemplateWithBaseLayout(templateBadFormPath)
	loginTemplate := parseTemplateWithBaseLayout(templateLoginPath)
	signupTemplate := parseTemplateWithBaseLayout(templateSignupPath)
	recoveryTemplate := parseTemplateWithBaseLayout(templateRecoveryPath)
	recoveryCommitTemplate := parseTemplateWithBaseLayout(templateRecoveryCommitPath)

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			customRedirect(w, r, "/dashboard")
		})

		r.Get("/500", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			pages500Template.Execute(w, nil)
		})

		r.Get("/bad_form", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			pagesBadFormTemplate.Execute(w, nil)
		})

		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			data := map[string]any{
				"Redirect": r.URL.Query().Get("redirect"),
				"Error":    "",
				"Username": "",
			}
			loginTemplate.Execute(w, data)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}

			username := strings.TrimSpace(r.FormValue("username"))
			password := r.FormValue("password")
			redirect := r.URL.Query().Get("redirect")

			u, err := userService.AuthenticateUser(r.Context(), username, password)
			if err != nil {
				data := map[string]any{
					"Redirect": redirect,
					"Username": username,
					"Error":    "Usuario o contraseña incorrectos",
				}
				loginTemplate.Execute(w, data)
				return
			}

			// Login exitoso
			sessionID := auth.CreateSession(u.ID)

			http.SetCookie(w, &http.Cookie{
				Name:     auth.SessionIdCookie,
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   config.Get().Security.SecureHTTP,
				SameSite: http.SameSiteLaxMode, // Mejor que NoneMode para login normal
				Expires:  time.Now().In(timezone.ParaguayTZ).Add(30 * time.Minute),
			})

			if redirect == "" {
				redirect = "/dashboard"
			}

			http.Redirect(w, r, redirect, http.StatusSeeOther)
		})

		r.Get("/signup", func(w http.ResponseWriter, r *http.Request) {
			data := map[string]any{
				"Username": "",
				"Email":    "",
				"Error":    "",
				"Success":  "",
			}
			signupTemplate.Execute(w, data)
		})

		r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				data := map[string]any{
					"Error": "Error al procesar el formulario",
				}
				signupTemplate.Execute(w, data)
				return
			}

			username := strings.TrimSpace(r.FormValue("username"))
			email := strings.TrimSpace(r.FormValue("email"))
			password := r.FormValue("password")
			confirm := r.FormValue("confirm_password")

			err := userService.CreateUser(r.Context(), username, email, password, confirm)

			if err != nil {
				var msg string

				switch e := err.(type) {
				case userDomain.ValidationError:
					msg = e.Message
				case error:
					switch {
					case errors.Is(err, userDomain.ErrUsernameTaken):
						msg = "Este nombre de usuario ya está en uso"
					case errors.Is(err, userDomain.ErrEmailTaken):
						msg = "Este correo electrónico ya está registrado"
					default:
						msg = "Ocurrió un error inesperado"
					}
				}

				data := map[string]any{
					"Username": username,
					"Email":    email,
					"Error":    msg,
				}
				signupTemplate.Execute(w, data)
				return
			}

			data := map[string]any{
				"Success": "¡Cuenta creada correctamente! Ya puedes iniciar sesión.",
			}
			signupTemplate.Execute(w, data)
		})

		r.Get("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			data := map[string]any{
				"Error":   "",
				"Success": "",
				"Email":   "",
			}
			err := recoveryTemplate.Execute(w, data)
			if err != nil {
				logger.Error("Cannot load recovery template", "error", err)
				customRedirect(w, r, "/500")
			}
		})

		r.Post("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				data := map[string]any{
					"Error": "Error al procesar el formulario",
					"Email": r.Form.Get("email"),
				}
				recoveryTemplate.Execute(w, data)
				return
			}

			email := strings.TrimSpace(r.Form.Get("email"))

			// Create a recovery token
			token, err := userService.StartPasswordRecovery(r.Context(), email)

			if err != nil {
				var msg string
				if ve, ok := err.(userDomain.ValidationError); ok && ve.Field == "email" {
					msg = ve.Message
				} else {
					msg = "Ocurrió un error al procesar tu solicitud. Inténtalo nuevamente."
				}

				data := map[string]any{
					"Error": msg,
					"Email": email,
				}
				recoveryTemplate.Execute(w, data)
				return
			}

			// User not exists (do not leak existence)
			if token == "" {
				data := map[string]any{
					"Success": "Si el correo existe, te hemos enviado un enlace de recuperación.",
					"Email":   email,
				}
				recoveryTemplate.Execute(w, data)
				return
			}

			// Send email
			if err := emailService.SendRecoveryEmail(email, token); err != nil {
				logger.Error("Failed to send recovery email", "error", err)
				data := map[string]any{
					"Error": "Hubo un problema al enviar el correo. Por favor intenta nuevamente más tarde.",
					"Email": email,
				}
				recoveryTemplate.Execute(w, data)
				return
			}

			data := map[string]any{
				"Success": "Si el correo existe, te hemos enviado un enlace de recuperación.",
				"Email":   email,
			}
			recoveryTemplate.Execute(w, data)
		})

		r.Get("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if token == "" {
				customRedirect(w, r, "/500")
				return
			}

			data := map[string]any{
				"Token":   token,
				"Error":   "",
				"Success": "",
			}
			recoveryCommitTemplate.Execute(w, data)
		})

		r.Post("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if token == "" {
				data := map[string]any{
					"Token": token,
					"Error": "Token inválido",
				}
				recoveryCommitTemplate.Execute(w, data)
				return
			}

			if err := r.ParseForm(); err != nil {
				data := map[string]any{
					"Token": token,
					"Error": "Error al procesar el formulario",
				}
				recoveryCommitTemplate.Execute(w, data)
				return
			}

			password := r.Form.Get("password")
			confirm := r.Form.Get("confirm_password")

			err := userService.CommitPasswordRecovery(r.Context(), token, password, confirm)
			if err != nil {
				var msg string
				if ve, ok := err.(userDomain.ValidationError); ok {
					msg = ve.Message
				} else if errors.Is(err, userDomain.ErrInvalidToken) {
					msg = "El enlace es inválido o ha expirado"
				} else {
					msg = "No se pudo actualizar la contraseña"
				}

				data := map[string]any{
					"Token": token,
					"Error": msg,
				}
				recoveryCommitTemplate.Execute(w, data)
				return
			}

			data := map[string]any{
				"Token":   token,
				"Success": "¡Contraseña actualizada correctamente! Ya puedes iniciar sesión.",
			}
			recoveryCommitTemplate.Execute(w, data)
		})
	}
}
