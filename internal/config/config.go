package config

import "os"

type config struct {
	Port        string
	DatabaseUrl string
	JWTSecret   string
}

func Load() *config {
	return &config{
		Port:        getEnv("PORT", "8080"),
		DatabaseUrl: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/auth_service?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
