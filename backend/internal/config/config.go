package config

import (
	"os"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/db"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Env    string
	Port   string
	ApiURL string
	DB     db.Config
}

func Load() *AppConfig {
	godotenv.Load() // best-effort, intentionally ignored
	env := getEnv("APP_ENV", "development")
	return &AppConfig{
		Env:    env,
		Port:   getEnv("PORT", "8080"),
		ApiURL: getEnv("API_URL", "http://localhost:8080/api/v1"),
		DB: db.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "securepass2026!"),
			DBName:   getEnv("DB_NAME", "db"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
