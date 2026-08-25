package config

import "os"

type config struct {
	Port        string
	DatabaseUrl string
}

func Load() *config {
	return &config{
		Port:        getEnv("PORT", "8080"),
		DatabaseUrl: getEnv("DATABASE_URL", "postgres://username:password@localhost:5432/auth_service?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
