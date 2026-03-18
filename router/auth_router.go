package router

import (
	"errors"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/app/email"
	"github.com/elias-gill/poliplanner2/internal/app/user"
	"github.com/elias-gill/poliplanner2/internal/auth"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

func NewAuthRouter(userService *user.UserService, emailService *email.EmailService) func(r chi.Router) {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages")

	// templates paths
	template500Path := path.Join(baseDir, "500.html")
	templateLoginPath := path.Join(baseDir, "auth", "login.html")
	templateSignupPath := path.Join(baseDir, "auth", "signup.html")
	templateRecoveryPath := path.Join(baseDir, "auth", "password-recovery.html")
	templateRecoveryCommitPath := path.Join(baseDir, "auth", "password-recovery-commit.html")

	// parse templates
	pages500Template := parseTemplateWithBaseLayout(template500Path)
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

		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			data := map[string]any{"Redirect": r.URL.Query().Get("redirect")}
			loginTemplate.Execute(w, data)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			u, err := userService.AuthenticateUser(r.Context(), username, password)
			if err != nil {
				executeFragment(w, r, "messages/error_message", "Invalid username or password")
				return
			}

			sessionID := auth.CreateSession(u.ID)

			// set session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   config.Get().Security.SecureHTTP,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().In(timezone.ParaguayTZ).Add(30 * time.Minute),
			})

			redirectTo := r.URL.Query().Get("redirect")
			if redirectTo == "" {
				redirectTo = "/dashboard"
			}

			customRedirect(w, r, redirectTo)
		})

		r.Get("/signup", func(w http.ResponseWriter, r *http.Request) {
			signupTemplate.Execute(w, nil)
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
				case user.ValidationError:
					field, msg = e.Field, e.Message
				case error:
					switch {
					case errors.Is(err, user.ErrUsernameTaken):
						field, msg = "username", "Username already taken"
					case errors.Is(err, user.ErrEmailTaken):
						field, msg = "email", "Email already in use"
					default:
						field, msg = "", "Unexpected error"
					}
				}

				if field != "" {
					executeFragment(w, r, "messages/error_message", field+": "+msg)
				} else {
					executeFragment(w, r, "messages/error_message", "Failed to create account")
				}
				return
			}

			customRedirect(w, r, "/login")
		})

		r.Get("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			recoveryTemplate.Execute(w, nil)
		})

		r.Post("/password-recovery", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				executeFragment(w, r, "messages/error_message", "Failed to process form")
				return
			}

			email := strings.TrimSpace(r.Form.Get("email"))

			token, err := userService.StartPasswordRecovery(r.Context(), email)
			if err != nil {
				var msg string
				if ve, ok := err.(user.ValidationError); ok && ve.Field == "email" {
					msg = ve.Message
				} else {
					msg = "If the email exists, a recovery link has been sent."
				}
				executeFragment(w, r, "messages/success_message", msg)
				return
			}

			// User does not exists
			if token == "" {
				executeFragment(w, r, "messages/success_message",
					"If the email exists, a recovery link has been sent.")
				return
			}

			if err := emailService.SendRecoveryEmail(email, token); err != nil {
				logger.Error("Failed to send recovery email", "error", err)
				customRedirect(w, r, "/500")
				return
			}

			executeFragment(w, r, "messages/success_message",
				"If the email exists, a recovery link has been sent.")
		})

		r.Get("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if token == "" {
				customRedirect(w, r, "/500")
				return
			}
			data := map[string]any{"Token": token}
			recoveryCommitTemplate.Execute(w, data)
		})

		r.Post("/password-recovery/{token}", func(w http.ResponseWriter, r *http.Request) {
			token := chi.URLParam(r, "token")
			if token == "" {
				executeFragment(w, r, "messages/error_message", "Invalid Token")
				return
			}

			if err := r.ParseForm(); err != nil {
				executeFragment(w, r, "messages/error_message", "Failed to parse form")
				return
			}

			password := r.Form.Get("password")
			confirm := r.Form.Get("confirm_password")

			err := userService.CommitPasswordRecovery(r.Context(), token, password, confirm)
			if err != nil {
				var field, msg string
				if ve, ok := err.(user.ValidationError); ok {
					field, msg = ve.Field, ve.Message
				} else if errors.Is(err, user.ErrInvalidToken) {
					msg = "Invalid or expired token"
				} else {
					msg = "Failed to update password"
				}

				if field != "" {
					executeFragment(w, r, "messages/error_message", field+": "+msg)
				} else {
					executeFragment(w, r, "messages/error_message", msg)
				}
				return
			}

			executeFragment(w, r, "messages/success_message", "Password updated successfully")
		})
	}
}
