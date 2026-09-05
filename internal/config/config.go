package config

import "os"

type config struct {
	Port        string
	DatabaseUrl string
	JWTSecret   string

	SMTPHost string
	SMTPPort string
	SMTPFrom string
	AppName  string

	SMSProvider      string
	MSG91AuthKey     string
	MSG91FlowId      string
	MSG91SenderId    string
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
}

func Load() *config {
	return &config{
		Port:        getEnv("PORT", "8080"),
		DatabaseUrl: getEnv("DATABASE_URL", "postgres://postgres:0000@localhost:5432/auth_service?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),

		SMTPHost: getEnv("SMTP_host", "localhost"),
		SMTPPort: getEnv("SMTP_port", "1025"),
		SMTPFrom: getEnv("SMTP_from", "no-reply@authsevice.dev"),
		AppName:  getEnv("App_Name", "AuthX"),

		SMSProvider:      getEnv("SMS_PROVIDER", "console"),
		MSG91AuthKey:     getEnv("MSG91_AUTH_KEY", ""),
		MSG91FlowId:      getEnv("MSG91_FLOW_ID", ""),
		MSG91SenderId:    getEnv("MSG_SENDER_ID", ""),
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
