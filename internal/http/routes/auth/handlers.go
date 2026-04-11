package auth

import (
	"errors"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/elias-gill/poliplanner2/internal/app/auth"
	"github.com/elias-gill/poliplanner2/internal/app/email"
	userApp "github.com/elias-gill/poliplanner2/internal/app/user"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	userDomain "github.com/elias-gill/poliplanner2/internal/domain/user"
	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/middleware"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

var (
	baseDir = path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages")

	// templates paths
	template500Path            = path.Join(baseDir, "500.html")
	templateBadFormPath        = path.Join(baseDir, "bad_form.html")
	templateLoginPath          = path.Join(baseDir, "auth", "login.html")
	templateSignupPath         = path.Join(baseDir, "auth", "signup.html")
	templateRecoveryPath       = path.Join(baseDir, "auth", "password-recovery.html")
	templateRecoveryCommitPath = path.Join(baseDir, "auth", "password-recovery-commit.html")

	// parsed templates
	pages500Template       = utils.ParseTemplateWithBaseLayout(template500Path)
	pagesBadFormTemplate   = utils.ParseTemplateWithBaseLayout(templateBadFormPath)
	loginTemplate          = utils.ParseTemplateWithBaseLayout(templateLoginPath)
	signupTemplate         = utils.ParseTemplateWithBaseLayout(templateSignupPath)
	recoveryTemplate       = utils.ParseTemplateWithBaseLayout(templateRecoveryPath)
	recoveryCommitTemplate = utils.ParseTemplateWithBaseLayout(templateRecoveryCommitPath)
)

type AuthHandlers struct {
	userService  *userApp.User
	authService  *auth.AuthManager
	emailService *email.EmailSender
}

func newAuthHandlers(userService *userApp.User, authService *auth.AuthManager, emailService *email.EmailSender) *AuthHandlers {
	return &AuthHandlers{
		userService:  userService,
		authService:  authService,
		emailService: emailService,
	}
}

// ==================== Handlers ====================

func (h *AuthHandlers) RedirectToDashboard(w http.ResponseWriter, r *http.Request) {
	utils.CustomRedirect(w, r, "/dashboard")
}

func (h *AuthHandlers) Serve500Page(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	pages500Template.Execute(w, nil)
}

func (h *AuthHandlers) ServeBadFormPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	pagesBadFormTemplate.Execute(w, nil)
}

func (h *AuthHandlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Redirect": r.URL.Query().Get("redirect"),
		"Error":    "",
		"Username": "",
	}
	loginTemplate.Execute(w, data)
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/dashboard"
	}

	session, err := h.authService.Login(r.Context(), username, password)
	if err != nil {
		data := map[string]any{
			"Redirect": redirect,
			"Username": username,
			"Error":    "Usuario o contraseña incorrectos",
		}
		loginTemplate.Execute(w, data)
		return
	}

	// Successful login
	setSessionCookie(w, session.ID)
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h *AuthHandlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Username": "",
		"Email":    "",
		"Error":    "",
		"Success":  "",
	}
	signupTemplate.Execute(w, data)
}

func (h *AuthHandlers) Signup(w http.ResponseWriter, r *http.Request) {
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

	err := h.userService.CreateUser(r.Context(), username, email, password, confirm)

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
}

func (h *AuthHandlers) PasswordRecoveryPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Error":   "",
		"Success": "",
		"Email":   "",
	}
	err := recoveryTemplate.Execute(w, data)
	if err != nil {
		logger.Error("Cannot load recovery template", "error", err)
		utils.CustomRedirect(w, r, "/500")
	}
}

func (h *AuthHandlers) PasswordRecovery(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		data := map[string]any{
			"Error": "Error al procesar el formulario",
			"Email": r.Form.Get("email"),
		}
		recoveryTemplate.Execute(w, data)
		return
	}

	email := strings.TrimSpace(r.Form.Get("email"))

	token, err := h.userService.StartPasswordRecovery(r.Context(), email)

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

	if token == "" {
		data := map[string]any{
			"Success": "Si el correo existe, te hemos enviado un enlace de recuperación.",
			"Email":   email,
		}
		recoveryTemplate.Execute(w, data)
		return
	}

	if err := h.emailService.SendRecoveryEmail(email, token); err != nil {
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
}

func (h *AuthHandlers) PasswordRecoveryCommitPage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		utils.CustomRedirect(w, r, "/500")
		return
	}

	data := map[string]any{
		"Token":   token,
		"Error":   "",
		"Success": "",
	}
	recoveryCommitTemplate.Execute(w, data)
}

func (h *AuthHandlers) PasswordRecoveryCommit(w http.ResponseWriter, r *http.Request) {
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

	err := h.userService.CommitPasswordRecovery(r.Context(), token, password, confirm)
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
}

func setSessionCookie(w http.ResponseWriter, token auth.SessionID) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionIdCookie,
		Value:    string(token),
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Get().Security.SecureHTTP,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().In(timezone.ParaguayTZ).Add(30 * time.Minute),
	})
}
