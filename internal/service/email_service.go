package service

import (
	"context"

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

	return client
}

func (e *EmailService) SendRecoveryEmail(to string) {
	if !e.enabled {
		logger.Debug("Email support is disabled")
		return
	}

	email := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Email: "poliplanner.web@gmail.com",
			Name:  "Poliplanner",
		},
		To: []brevo.SendSmtpEmailTo{
			{Email: to},
		},
		Subject:     "Hello world",
		HtmlContent: "<p>hello world</p>",
	}

	ctx := context.Background()
	_, _, err := e.client.TransactionalEmailsApi.SendTransacEmail(ctx, email)
	if err != nil {
		logger.Error("Cannot send email", "error", err)
	}
}
