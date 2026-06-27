package config

import (
	"os"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/db"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/token"
	"bitbucket.org/HyperEther/document-management-system/internal/storage"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Env     string
	Port    string
	ApiURL  string
	AppRoot string
	DB      db.Config
	Minio   storage.MinioConfig
	Auth    token.Config
}

func Load() *AppConfig {
	godotenv.Load() // best-effort, intentionally ignored
	env := getEnv("APP_ENV", "development")
	return &AppConfig{
		Env:     env,
		Port:    getEnv("PORT", "8080"),
		AppRoot: getEnv("APP_ROOT", ""),
		ApiURL:  getEnv("API_URL", "http://localhost:8080/api/v1"),
		Auth: token.Config{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		},
		Minio: storage.MinioConfig{
			Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:  getEnv("MINIO_ACCESS_KEY", "oblak_admin"),
			SecretKey:  getEnv("MINIO_SECRET_KEY", "oblak_super_secret_password"),
			BucketName: "oblak-artifacts",
			UseSSL:     getEnv("MINIO_USE_SSL", "") == "true",
		},
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
