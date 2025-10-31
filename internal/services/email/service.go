package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"

	"ypeskov/kkal-tracker/internal/config"
)

// Service handles email sending operations
type Service struct {
	config *config.Config
	logger *slog.Logger
}

// New creates a new email service
func New(cfg *config.Config, logger *slog.Logger) *Service {
	return &Service{
		config: cfg,
		logger: logger.With("service", "email"),
	}
}

// ActivationEmailData contains data for activation email template
type ActivationEmailData struct {
	Email         string
	ActivationURL string
	AppName       string
}

// SendActivationEmail sends an activation email to the user
func (s *Service) SendActivationEmail(toEmail, token, language string) error {
	s.logger.Debug("Sending activation email", "to", toEmail, "language", language)

	activationURL := fmt.Sprintf("%s/activate/%s", s.config.AppURL, token)

	data := ActivationEmailData{
		Email:         toEmail,
		ActivationURL: activationURL,
		AppName:       "Kkal Tracker",
	}

	// Select template based on language
	var tmpl *template.Template
	var err error

	switch language {
	case "uk_UA":
		tmpl, err = template.New("activation").Parse(activationTemplateUK)
	case "ru_UA":
		tmpl, err = template.New("activation").Parse(activationTemplateRU)
	default:
		tmpl, err = template.New("activation").Parse(activationTemplateEN)
	}

	if err != nil {
		s.logger.Error("Failed to parse email template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		s.logger.Error("Failed to execute email template", "error", err)
		return err
	}

	// Prepare email headers
	subject := s.getSubject(language)
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		s.config.SMTPFrom, toEmail, subject)

	message := []byte(headers + body.String())

	// Send email via SMTP
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	smtpAddr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	err = smtp.SendMail(smtpAddr, auth, s.config.SMTPFrom, []string{toEmail}, message)
	if err != nil {
		s.logger.Error("Failed to send email", "error", err, "to", toEmail)
		return err
	}

	s.logger.Info("Activation email sent", "to", toEmail)
	return nil
}
