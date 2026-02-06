package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/auth"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/internal/service"
	"github.com/elias-gill/poliplanner2/web"
	"github.com/go-chi/chi/v5"
)

func NewAuthRouter(userService *service.UserService, emailService *service.EmailService) func(r chi.Router) {
	layouts := web.BaseLayout

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		})

		r.Get("/500", func(w http.ResponseWriter, r *http.Request) {
			execTemplateWithLayout(w, "web/templates/pages/500.html", layouts, nil)
		})

		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			data := map[string]any{"Redirect": r.URL.Query().Get("redirect")}
			execTemplateWithLayout(w, "web/templates/pages/auth/login.html", layouts, data)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			user, err := userService.AuthenticateUser(r.Context(), username, password)
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(newErrorFragment("Invalid username or password")))
				return
			}

			sessionID := auth.CreateSession(user.ID)

			// set session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   config.Get().Security.SecureHTTP,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().Add(30 * time.Minute),
			})

			redirectTo := r.URL.Query().Get("redirect")
			if redirectTo == "" {
				redirectTo = "/dashboard"
			}

			w.Header().Set("HX-Redirect", redirectTo)
		})

		r.Get("/signup", func(w http.ResponseWriter, r *http.Request) {
			execTemplateWithLayout(w, "web/templates/pages/auth/signup.html", layouts, nil)
		})

		r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")
			confirm := r.FormValue("confirm_password")

			err := userService.CreateUser(r.Context(), username, email, password, confirm)
			if err != nil {
				var field, msg string

				switch e := err.(type) {
				case service.ValidationError:
					field, msg = e.Field, e.Message
				case error:
					switch {
					case errors.Is(err, service.ErrUsernameTaken):
						field, msg = "username", "Username already taken"
					case errors.Is(err, service.ErrEmailTaken):
						field, msg = "email", "Email already in use"
					default:
						field, msg = "", "Unexpected error"
					}
				}

				if field != "" {
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte(newFieldErrorFragment(field, msg)))
				} else {
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte(newErrorFragment("Failed to create account")))
				}
				return
			}

			w.Header().Set("HX-Redirect", "/login")
		})

		r.Get("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			execTemplateWithLayout(w, "web/templates/pages/auth/password-recovery.html", layouts, nil)
		})

		r.Post("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				w.Write([]byte(newErrorFragment("Failed to process form")))
				return
			}

			email := strings.TrimSpace(r.Form.Get("email"))

			token, err := userService.StartPasswordRecovery(r.Context(), email)
			if err != nil {
				var msg string
				if ve, ok := err.(service.ValidationError); ok && ve.Field == "email" {
					msg = ve.Message
				} else {
					msg = "If the email exists, a recovery link has been sent."
				}
				w.Write([]byte(newSuccessFragment(msg)))
				return
			}

			if token == "" {
				w.Write([]byte(newSuccessFragment("If the email exists, a recovery link has been sent.")))
				return
			}

			if err := emailService.SendRecoveryEmail(email, token); err != nil {
				logger.Error("Failed to send recovery email", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			w.Write([]byte(newSuccessFragment("If the email exists, a recovery link has been sent.")))
		})

		r.Get("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if token == "" {
				customRedirect(w, r, "/500")
				return
			}
			data := map[string]any{"Token": token}
			execTemplateWithLayout(w, "web/templates/pages/auth/password-recovery-commit.html", layouts, data)
		})

		r.Post("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if token == "" {
				w.Write([]byte(newErrorFragment("Invalid token")))
				return
			}

			if err := r.ParseForm(); err != nil {
				w.Write([]byte(newErrorFragment("Failed to process form")))
				return
			}

			password := r.Form.Get("password")
			confirm := r.Form.Get("confirm_password")

			err := userService.CommitPasswordRecovery(r.Context(), token, password, confirm)
			if err != nil {
				var field, msg string
				if ve, ok := err.(service.ValidationError); ok {
					field, msg = ve.Field, ve.Message
				} else if errors.Is(err, service.ErrInvalidToken) {
					msg = "Invalid or expired token"
				} else {
					msg = "Failed to update password"
				}

				if field != "" {
					w.Write([]byte(newFieldErrorFragment(field, msg)))
				} else {
					w.Write([]byte(newErrorFragment(msg)))
				}
				return
			}

			w.Write([]byte(newSuccessFragment("Password updated successfully")))
		})
	}
}

// Helpers (actualizados para soportar errores por campo)
func newErrorFragment(msg string) string {
	return `
	<section role="alert" class="error">
		<span>` + msg + `</span>
		<button type="button" onclick="this.parentElement.remove()" aria-label="Close">×</button>
	</section>
	`
}

func newSuccessFragment(msg string) string {
	return `
	<section role="alert" class="success">
		<span>` + msg + `</span>
		<button type="button" onclick="this.parentElement.remove()" aria-label="Close">×</button>
	</section>
	`
}

func newFieldErrorFragment(field, msg string) string {
	return fmt.Sprintf(`
		<div class="field-error" data-field="%s">
			<span>%s</span>
		</div>
	`, field, msg)
}
