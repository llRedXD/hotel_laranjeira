package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL          string
	Port                 string
	AppEnv               string
	JWTSecret            string
	FrontendOrigin       string
	InitialAdminName     string
	InitialAdminEmail    string
	InitialAdminPassword string
	InitialReceptionName     string
	InitialReceptionEmail    string
	InitialReceptionPassword string
}

func Load() Config {
	return Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://hotel_laranjeira:hotel_laranjeira_dev@localhost:5432/hotel_laranjeira?sslmode=disable"),
		Port:                 getEnv("PORT", "8080"),
		AppEnv:               getEnv("APP_ENV", "development"),
		JWTSecret:            getEnv("JWT_SECRET", "hotel-laranjeira-development-secret-change-me"),
		FrontendOrigin:       getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		InitialAdminName:     os.Getenv("INITIAL_ADMIN_NAME"),
		InitialAdminEmail:    os.Getenv("INITIAL_ADMIN_EMAIL"),
		InitialAdminPassword: os.Getenv("INITIAL_ADMIN_PASSWORD"),
		InitialReceptionName:     os.Getenv("INITIAL_RECEPTION_NAME"),
		InitialReceptionEmail:    os.Getenv("INITIAL_RECEPTION_EMAIL"),
		InitialReceptionPassword: os.Getenv("INITIAL_RECEPTION_PASSWORD"),
	}
}

func (c Config) IsDevelopment() bool {
	environment := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return environment == "development" || environment == "dev"
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
