package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/Ashoke15/AuthX/internal/auth"
)

//go:embed templates/verification_email.html
var templatesFs embed.FS

type SMTPMailer struct {
	host    string
	port    string
	from    string
	appName string
	tmpl    *template.Template
}

func NewSmtpMailer(host, port, from, appName string) (*SMTPMailer, error) {
	tmpl, err := template.ParseFS(templatesFs, "templates/verification_email.html")
	if err != nil {
		return nil, fmt.Errorf("parse email template: %w", err)
	}

	return &SMTPMailer{host: host, port: port, from: from, appName: appName, tmpl: tmpl}, nil
}

type verificationEmailData struct {
	AppName       string
	Email         string
	Code          string
	ExpiryMinutes int
}

func (m *SMTPMailer) SendVerificationEmail(toEmail, code string) error {
	var body bytes.Buffer

	data := verificationEmailData{
		AppName:       m.appName,
		Email:         toEmail,
		Code:          code,
		ExpiryMinutes: int(auth.OTPTTL.Minutes()),
	}

	if err := m.tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("render email template: %w", err)
	}

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Verify your email\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		m.from, toEmail,
	)

	msg := []byte(headers + body.String())

	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	return smtp.SendMail(addr, nil, m.from, []string{toEmail}, msg)
}
