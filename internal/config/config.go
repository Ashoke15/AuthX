package config

import "os"

type config struct {
	Port        string
	DatabaseUrl string
	JWTSecret   string
	SMTPHost    string
	SMTPPort    string
	SMTPFrom    string
	AppName     string
}

func Load() *config {
	return &config{
		Port:        getEnv("PORT", "8080"),
		DatabaseUrl: getEnv("DATABASE_URL", "postgres://postgres:0000@localhost:5432/auth_service?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		SMTPHost:    getEnv("SMTP_host", "localhost"),
		SMTPPort:    getEnv("SMTP_port", "1025"),
		SMTPFrom:    getEnv("SMTP_from", "no-reply@authsevice.dev"),
		AppName:     getEnv("App_Name", "AuthX"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
