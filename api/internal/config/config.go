package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Port            string
	Environment     string
	AllowedOrigins  []string
	DefaultLanguage string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		AllowedOrigins:  strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		DefaultLanguage: getEnv("DEFAULT_LANGUAGE", "pt-BR"),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		slog.Error("required environment variable not set", slog.String("key", key))
		os.Exit(1)
	}
	return val
}
