package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/i18n"
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
		AppName:       i18n.GetTranslator().Get(language, "app.name"),
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

// ExportEmailData contains data for export email template
type ExportEmailData struct {
	FileName string
	AppName  string
}

// SendEmailWithAttachment sends an email with a file attachment
func (s *Service) SendEmailWithAttachment(toEmail, language string, attachment []byte, attachmentName string) error {
	s.logger.Debug("Sending email with attachment", "to", toEmail, "attachment", attachmentName, "language", language)

	data := ExportEmailData{
		FileName: attachmentName,
		AppName:  i18n.GetTranslator().Get(language, "app.name"),
	}

	// Select template based on language
	var tmpl *template.Template
	var err error

	switch language {
	case "uk_UA":
		tmpl, err = template.New("export").Parse(exportTemplateUK)
	case "ru_UA":
		tmpl, err = template.New("export").Parse(exportTemplateRU)
	case "bg_BG":
		tmpl, err = template.New("export").Parse(exportTemplateBG)
	default:
		tmpl, err = template.New("export").Parse(exportTemplateEN)
	}

	if err != nil {
		s.logger.Error("Failed to parse export email template", "error", err)
		return err
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		s.logger.Error("Failed to execute export email template", "error", err)
		return err
	}

	// Build MIME multipart message
	boundary := "==KKAL_EXPORT_BOUNDARY=="
	subject := s.getExportSubject(language)

	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", s.config.SMTPFrom))
	message.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	message.WriteString("\r\n")

	// HTML part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody.String())
	message.WriteString("\r\n")

	// Attachment part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n")
	message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachmentName))
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString("\r\n")

	// Base64 encode attachment with line breaks every 76 characters
	encoded := base64.StdEncoding.EncodeToString(attachment)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		message.WriteString(encoded[i:end])
		message.WriteString("\r\n")
	}

	// End boundary
	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Send email via SMTP
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	smtpAddr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	err = smtp.SendMail(smtpAddr, auth, s.config.SMTPFrom, []string{toEmail}, message.Bytes())
	if err != nil {
		s.logger.Error("Failed to send email with attachment", "error", err, "to", toEmail)
		return err
	}

	s.logger.Info("Export email sent", "to", toEmail, "attachment", attachmentName)
	return nil
}
