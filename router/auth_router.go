package router

import (
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
	"regexp"
	"time"

	"github.com/elias-gill/poliplanner2/service"
	"github.com/go-chi/chi/v5"
)

func NewAuthRouter() func(r chi.Router) {
	layouts := template.Must(template.ParseGlob("templates/layout/*.html"))

	// NOTE: made like this so the main layout template is parsed only one time on startup
	return func(r chi.Router) {
		// Render login
		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/pages/login.html"))
			tpl.Execute(w, nil)
		})

		// Handle login POST
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			user, err := service.AuthenticateUser(username, password)
			if err != nil {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(constructError("Usuario o contraseña incorrectos")))
				return
			}

			sessionID := service.CreateSession(user.UserID)

			cookie := &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().Add(30 * time.Minute),
			}
			http.SetCookie(w, cookie)

			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusOK)
		})

		// Handle signup POST
		r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			email := r.FormValue("email")
			rawPassword := r.FormValue("password")

			if err := validateUsername(username); err != nil {
				w.Write([]byte(constructError(err.Error())))
				return
			}

			if !isValidEmail(email) {
				w.Write([]byte(constructError("Invalid email provided")))
				return
			}

			if len(rawPassword) < 6 {
				w.Write([]byte(constructError("Password must have at least 6 characters")))
				return
			}

			err := service.CreateUser(username, email, rawPassword)
			if err != nil {
				w.Write([]byte(constructError(err.Error())))
				return
			}

			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusOK)
		})

		// Render signup
		r.Get("/signup", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			tpl := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/pages/signup.html"))
			tpl.Execute(w, nil)
		})
	}
}

func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("Username must have at least 3 characters")
	}
	if !isAlphanumeric(username) {
		return fmt.Errorf("Username must only contain letters, numbers and _")
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

func constructError(msg string) string {
	return `
	<section role="alert" class="error">
		<span>` + msg + `</span>
		<button type="button" style="background-color: var(--color-error);" onclick="this.parentElement.remove()" aria-label="Cerrar alerta">×</button>
	</section>
	`
}
