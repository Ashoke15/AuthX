package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/Ashoke15/AuthX/internal/auth"
)

//go:embed templates/verification_email.html templates/password_reset_email.html
var templatesFs embed.FS

// mailer.go interface
type Mailer interface {
	SendVerificationEmail(toEmail, code string) error
	SendPasswordResetEmail(toEmail, code string) error
}

type SMTPMailer struct {
	host                string
	port                string
	from                string
	appName             string
	verifytmpl          *template.Template
	passwordResetTemple *template.Template
}

func NewSmtpMailer(host, port, from, appName string) (*SMTPMailer, error) {
	verifytmpl, err := template.ParseFS(templatesFs, "templates/verification_email.html")
	if err != nil {
		return nil, fmt.Errorf("parse email template: %w", err)
	}

	resetTemple, err := template.ParseFS(templatesFs, "templates/password_reset_email.html")
	if err != nil {
		return nil, fmt.Errorf("parse password reset templates: %w", err)
	}

	return &SMTPMailer{host: host, port: port, from: from, appName: appName, verifytmpl: verifytmpl, passwordResetTemple: resetTemple}, nil
}

type EmailData struct {
	AppName       string
	Email         string
	Code          string
	ExpiryMinutes int
}

func (m *SMTPMailer) SendVerificationEmail(toEmail, code string) error {
	return m.send(m.verifytmpl, toEmail, "verify your email", EmailData{
		AppName: m.appName, Email: toEmail, Code: code,
		ExpiryMinutes: int(auth.OTPTTL.Minutes()),
	})
}

func (m *SMTPMailer) SendPasswordResetEmail(toEmail, code string) error {
	return m.send(m.passwordResetTemple, toEmail, "Reset your password", EmailData{
		AppName: m.appName, Email: toEmail, Code: code,
		ExpiryMinutes: int(auth.OTPTTL.Minutes()),
	})
}

func (m *SMTPMailer) send(tmpl *template.Template, toEmail, subject string, data EmailData) error {
	var body bytes.Buffer

	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("render email template: %w", err)
	}

	header := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		m.from, toEmail, subject,
	)

	msg := []byte(header + body.String())
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	return smtp.SendMail(addr, nil, m.from, []string{toEmail}, msg)
}

type ConsoleMailer struct{}

func NewConsolemailer() *ConsoleMailer {
	return &ConsoleMailer{}
}

// ConsoleMailer
func (m *ConsoleMailer) SendVerificationEmail(toEmail, code string) error {
	println("=== VERIFICATION EMAIL ===")
	println("To:", toEmail)
	println("Code:", code)
	println("===========================")
	return nil
}

// ConsoleMailer
func (m *ConsoleMailer) SendPasswordResetEmail(toEmail, code string) error {
	println("=== PASSWORD RESET EMAIL ===")
	println("To:", toEmail)
	println("Code:", code)
	println("=============================")
	return nil
}
