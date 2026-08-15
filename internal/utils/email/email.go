package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"github.com/1kulture/1kulture-backend/internal/config"
	appLogger "github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type EmailService struct {
	config *config.EmailConfig
}

type EmailData struct {
	To       string
	Subject  string
	Template string
	Data     interface{}
}

func NewEmailService(cfg *config.EmailConfig) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

func (s *EmailService) SendEmail(data EmailData) error {
	// Load and parse template
	templatePath := filepath.Join(s.config.TemplatesPath, data.Template+".html")

	// Check if template exists
	if _, err := os.Stat(templatePath); err != nil {
		return fmt.Errorf("email template not found: %w", err)
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data.Data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// In development, just log the email
	if os.Getenv("ENVIRONMENT") == "development" {
		appLogger.WithFields(map[string]interface{}{
			"to":       data.To,
			"subject":  data.Subject,
			"template": data.Template,
		}).Info("Development mode: Email would be sent")

		// Save to file for testing
		devEmailPath := filepath.Join("logs", "emails")
		if err := os.MkdirAll(devEmailPath, 0755); err == nil {
			emailFile := filepath.Join(devEmailPath, fmt.Sprintf("%s_%s.html", time.Now().Format("20060102_150405"), data.Template))
			if err := os.WriteFile(emailFile, body.Bytes(), 0644); err == nil {
				appLogger.Info("Development email saved to: ", emailFile)
			}
		}
		return nil
	}

	// Production: Send via SMTP
	auth := smtp.PlainAuth(
		"",
		s.config.SMTPUsername,
		s.config.SMTPPassword,
		s.config.SMTPHost,
	)

	msg := buildMessage(s.config.FromName, s.config.FromAddress, data.To, data.Subject, body.String())
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	if err := smtp.SendMail(addr, auth, s.config.FromAddress, []string{data.To}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	appLogger.WithFields(map[string]interface{}{
		"to":       data.To,
		"subject":  data.Subject,
		"template": data.Template,
	}).Info("Email sent successfully")

	return nil
}

func (s *EmailService) SendVerificationEmail(to, code string) error {
	return s.SendEmail(EmailData{
		To:       to,
		Subject:  "Verify Your Email - 1Kulture",
		Template: "verification",
		Data: map[string]interface{}{
			"Code":    code,
			"Email":   to,
			"AppName": "1Kulture",
		},
	})
}

func (s *EmailService) SendPasswordResetEmail(to, resetLink string) error {
	return s.SendEmail(EmailData{
		To:       to,
		Subject:  "Reset Your Password - 1Kulture",
		Template: "password_reset",
		Data: map[string]interface{}{
			"ResetLink": resetLink,
			"Email":     to,
			"AppName":   "1Kulture",
		},
	})
}

func buildMessage(fromName, fromAddress, to, subject, body string) string {
	msg := fmt.Sprintf("From: %s <%s>\r\n", fromName, fromAddress)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-version: 1.0;\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += "\r\n"
	msg += body
	return msg
}
