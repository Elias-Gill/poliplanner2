package service

import (
	"context"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/logger"
	brevo "github.com/getbrevo/brevo-go/lib"
)

type EmailService struct {
	enabled bool
	apiKey  string
	client  *brevo.APIClient
}

func NewEmailService(apiKey string) *EmailService {
	client := &EmailService{enabled: false}

	if apiKey == "" {
		logger.Warn("Email api key is not set, Email Service will be disabled")
		return client
	}

	// Configure Brevo client
	client.enabled = true
	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	client.client = brevo.NewAPIClient(cfg)

	logger.Info("Email service configured")

	return client
}

func (e *EmailService) SendRecoveryEmail(to string, recToken string) error {
	if !e.enabled {
		logger.Debug("Email support is disabled")
		return fmt.Errorf("Email service is disabled")
	}

	email := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Email: "poliplanner.web@gmail.com",
			Name:  "Poliplanner",
		},
		To: []brevo.SendSmtpEmailTo{
			{Email: to},
		},
		Subject: "Poliplanner - Recuperacion de contraseña",
		HtmlContent: `
			<div style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; color: #333;">
			  <h2 style="color: #222;">Recuperación de contraseña</h2>

			  <img
				src="https://poliplanner2.fly.dev/static/favicon.png"
				alt="Logo PoliPlanner"
				style="display: block; margin: 16px auto 24px auto; width: 160px; height: auto;"
			  />

			  <p>Recibimos una solicitud para recuperar tu contraseña. Si no solicitaste este cambio, podés ignorar este correo.</p>

			  <a
			    href="https://poliplanner2.fly.dev/password-recovery/` + recToken + `"
			    style="
			  	display: block;
			  	width: 100%;
			  	box-sizing: border-box;
			  	background-color: #2563eb;
			  	color: #ffffff;
			  	padding: 14px 20px;
			  	text-decoration: none;
			  	border-radius: 6px;
			  	text-align: center;
			  	font-weight: 600;
			    ">
			    Restablecer contraseña
			  </a>


			  <div style=" width: 100%; height: 1px; background-color: #e5e7eb; margin: 24px 0; "> </div>

			  <p style="font-size: 12px; color: #666;">
				Este enlace expirará en 15 minutos.
			  </p>
			</div>`,
	}

	ctx := context.Background()
	_, _, err := e.client.TransactionalEmailsApi.SendTransacEmail(ctx, email)
	if err != nil {
		logger.Error("Cannot send email", "error", err)
	}

	return err
}
