package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL              string
	Port                     string
	AppEnv                   string
	JWTSecret                string
	FrontendOrigin           string
	InitialAdminName         string
	InitialAdminEmail        string
	InitialAdminPassword     string
	InitialReceptionName     string
	InitialReceptionEmail    string
	InitialReceptionPassword string
}

func Load() Config {
	return Config{
		DatabaseURL:              getEnv("DATABASE_URL", "postgres://hotel_laranjeira:hotel_laranjeira_dev@localhost:5432/hotel_laranjeira?sslmode=disable"),
		Port:                     getEnv("PORT", "8080"),
		AppEnv:                   getEnv("APP_ENV", "development"),
		JWTSecret:                getEnv("JWT_SECRET", "hotel-laranjeira-development-secret-change-me"),
		FrontendOrigin:           getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		InitialAdminName:         os.Getenv("INITIAL_ADMIN_NAME"),
		InitialAdminEmail:        os.Getenv("INITIAL_ADMIN_EMAIL"),
		InitialAdminPassword:     os.Getenv("INITIAL_ADMIN_PASSWORD"),
		InitialReceptionName:     os.Getenv("INITIAL_RECEPTION_NAME"),
		InitialReceptionEmail:    os.Getenv("INITIAL_RECEPTION_EMAIL"),
		InitialReceptionPassword: os.Getenv("INITIAL_RECEPTION_PASSWORD"),
	}
}

func (c Config) Validate() error {
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	if env != "dev" && env != "development" && env != "production" && env != "staging" && env != "test" {
		return fmt.Errorf("APP_ENV inválido")
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT inválida")
	}
	if len(c.JWTSecret) < 32 || (!c.IsDevelopment() && c.JWTSecret == "hotel-laranjeira-development-secret-change-me") {
		return fmt.Errorf("JWT_SECRET deve conter pelo menos 32 bytes e ser exclusivo fora de DEV")
	}
	origin, err := url.Parse(c.FrontendOrigin)
	if err != nil || origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || origin.Path != "" || (origin.Scheme != "http" && origin.Scheme != "https") {
		return fmt.Errorf("FRONTEND_ORIGIN deve ser uma origem HTTP sem caminho")
	}
	if !c.IsDevelopment() && origin.Scheme != "https" {
		return fmt.Errorf("FRONTEND_ORIGIN deve usar HTTPS fora de DEV")
	}
	return nil
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
