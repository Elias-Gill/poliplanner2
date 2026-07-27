package auth

import (
	"errors"
	"net/http"
	"strings"

	utils "github.com/elias-gill/poliplanner2/internal/http"
	"github.com/elias-gill/poliplanner2/internal/http/cookie"
	"github.com/elias-gill/poliplanner2/internal/http/render"
	userModel "github.com/elias-gill/poliplanner2/internal/model/user"
	"github.com/elias-gill/poliplanner2/internal/service/auth"
	"github.com/elias-gill/poliplanner2/internal/service/email"
	userService "github.com/elias-gill/poliplanner2/internal/service/user"
	"github.com/elias-gill/poliplanner2/logger"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	tmpl         *render.TemplateManager
	userService  *userService.UserService
	authService  *auth.SessionService
	emailService *email.EmailSender
}

func NewHandler(
	tmpl *render.TemplateManager,
	userService *userService.UserService,
	authService *auth.SessionService,
	emailService *email.EmailSender,
) *Handler {
	return &Handler{
		tmpl:         tmpl,
		userService:  userService,
		authService:  authService,
		emailService: emailService,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.Redirect(w, r, "/dashboard")
	})

	r.Get("/500", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if err := h.tmpl.RenderPage(w, "500.html", nil); err != nil {
			logger.Error("Cannot render 500 template", "error", err)
		}
	})

	r.Get("/bad_form", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if err := h.tmpl.RenderPage(w, "bad_form.html", nil); err != nil {
			logger.Error("Cannot render bad_form template", "error", err)
		}
	})

	r.Get("/login", h.loginPage)
	r.Post("/login", h.login)

	// REFACTOR: plantearme si no deberian de estar en user por ejemplo
	r.Get("/signup", h.signupPage)
	r.Post("/signup", h.signup)

	r.Get("/password-recovery", h.passwordRecoveryPage)
	r.Post("/password-recovery", h.passwordRecovery)

	r.Get("/password-recovery/{token}", h.passwordRecoveryCommitPage)
	r.Post("/password-recovery/{token}", h.passwordRecoveryCommit)

	return r
}

// ======================================
// =         Handlers HTTP              =
// ======================================

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Redirect": r.URL.Query().Get("redirect"),
		"Error":    "",
		"Username": "",
	}

	if err := h.tmpl.RenderPage(w, "auth/login.html", data); err != nil {
		logger.Error("Cannot render login template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
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
		h.tmpl.RenderPage(w, "auth/login.html", data)
		return
	}

	// Successful login
	cookie.SetSessionCookie(w, session.ID)
	utils.Redirect(w, r, redirect)
}

func (h *Handler) signupPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Username": "",
		"Email":    "",
		"Error":    "",
		"Success":  "",
	}
	h.tmpl.RenderPage(w, "auth/signup.html", data)
}

func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		data := map[string]any{
			"Error": "Error al procesar el formulario",
		}
		h.tmpl.RenderPage(w, "auth/signup.html", data)
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
		case userModel.ValidationError:
			msg = e.Message
		case error:
			switch {
			case errors.Is(err, userModel.ErrUsernameTaken):
				msg = "Este nombre de usuario ya está en uso"
			case errors.Is(err, userModel.ErrEmailTaken):
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
		h.tmpl.RenderPage(w, "auth/signup.html", data)
		return
	}

	data := map[string]any{
		"Success": "¡Cuenta creada correctamente! Ya puedes iniciar sesión.",
	}
	h.tmpl.RenderPage(w, "auth/signup.html", data)
}

func (h *Handler) passwordRecoveryPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Error":   "",
		"Success": "",
		"Email":   "",
	}
	err := h.tmpl.RenderPage(w, "auth/password-recovery.html", data)
	if err != nil {
		logger.Error("Cannot load recovery template", "error", err)
		utils.Redirect(w, r, "/500")
	}
}

func (h *Handler) passwordRecovery(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		data := map[string]any{
			"Error": "Error al procesar el formulario",
			"Email": r.Form.Get("email"),
		}
		h.tmpl.RenderPage(w, "auth/password-recovery.html", data)
		return
	}

	email := strings.TrimSpace(r.Form.Get("email"))
	token, err := h.userService.StartPasswordRecovery(r.Context(), email)
	if err != nil {
		var msg string
		if ve, ok := err.(userModel.ValidationError); ok && ve.Field == "email" {
			msg = ve.Message
		} else {
			msg = "Ocurrió un error al procesar tu solicitud. Inténtalo nuevamente."
		}

		data := map[string]any{
			"Error": msg,
			"Email": email,
		}
		h.tmpl.RenderPage(w, "auth/password-recovery.html", data)
		return
	}

	if token == "" {
		data := map[string]any{
			"Success": "Si el correo existe, te hemos enviado un enlace de recuperación.",
			"Email":   email,
		}
		h.tmpl.RenderPage(w, "aut/password-recovery.html", data)
		return
	}

	if err := h.emailService.SendRecoveryEmail(email, token); err != nil {
		logger.Error("Failed to send recovery email", "error", err)
		data := map[string]any{
			"Error": "Hubo un problema al enviar el correo. Por favor intenta nuevamente más tarde.",
			"Email": email,
		}
		h.tmpl.RenderPage(w, "auth/password-recovery.html", data)
		return
	}

	data := map[string]any{
		"Success": "Si el correo existe, te hemos enviado un enlace de recuperación.",
		"Email":   email,
	}
	h.tmpl.RenderPage(w, "auth/password-recovery.html", data)
}

func (h *Handler) passwordRecoveryCommitPage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		utils.Redirect(w, r, "/500")
		return
	}

	data := map[string]any{
		"Token":   token,
		"Error":   "",
		"Success": "",
	}
	h.tmpl.RenderPage(w, "auth/password-recovery-commit.html", data)
}

func (h *Handler) passwordRecoveryCommit(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		data := map[string]any{
			"Token": token,
			"Error": "Token inválido",
		}
		h.tmpl.RenderPage(w, "auth/password-recovery-commit.html", data)
		return
	}

	if err := r.ParseForm(); err != nil {
		data := map[string]any{
			"Token": token,
			"Error": "Error al procesar el formulario",
		}
		h.tmpl.RenderPage(w, "auth/password-recovery-commit.html", data)
		return
	}

	password := r.Form.Get("password")
	confirm := r.Form.Get("confirm_password")

	err := h.userService.CommitPasswordRecovery(r.Context(), token, password, confirm)
	if err != nil {
		var msg string
		if ve, ok := err.(userModel.ValidationError); ok {
			msg = ve.Message
		} else if errors.Is(err, userModel.ErrInvalidToken) {
			msg = "El enlace es inválido o ha expirado"
		} else {
			msg = "No se pudo actualizar la contraseña"
		}

		data := map[string]any{
			"Token": token,
			"Error": msg,
		}
		h.tmpl.RenderPage(w, "auth/password-recovery-commit.html", data)
		return
	}

	data := map[string]any{
		"Token":   token,
		"Success": "¡Contraseña actualizada correctamente! Ya puedes iniciar sesión.",
	}
	h.tmpl.RenderPage(w, "auth/password-recovery-commit.html", data)
}
