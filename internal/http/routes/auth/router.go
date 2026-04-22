package auth

import (
	"github.com/elias-gill/poliplanner2/internal/app/auth"
	"github.com/elias-gill/poliplanner2/internal/app/email"
	userApp "github.com/elias-gill/poliplanner2/internal/app/user"
	"github.com/go-chi/chi/v5"
)

func NewAuthRouter(userService *userApp.User, authService *auth.AuthManager, emailService *email.EmailSender) func(r chi.Router) {
	handlers := newAuthHandlers(userService, authService, emailService)

	return func(r chi.Router) {
		r.Get("/", handlers.RedirectToDashboard)
		r.Get("/500", handlers.Serve500Page)
		r.Get("/bad_form", handlers.ServeBadFormPage)

		r.Get("/login", handlers.LoginPage)
		r.Post("/login", handlers.Login)

		r.Get("/signup", handlers.SignupPage)
		r.Post("/signup", handlers.Signup)

		r.Get("/password-recovery", handlers.PasswordRecoveryPage)
		r.Post("/password-recovery", handlers.PasswordRecovery)

		r.Get("/password-recovery/{token}", handlers.PasswordRecoveryCommitPage)
		r.Post("/password-recovery/{token}", handlers.PasswordRecoveryCommit)
	}
}
