package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/auth"
	"bitbucket.org/HyperEther/document-management-system/internal/config"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/db"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/middleware"
	"bitbucket.org/HyperEther/document-management-system/internal/storage"
	"go.uber.org/zap"
)

type Application struct {
	DB     *sql.DB
	Router http.Handler
}

func setupDependencies(cfg *config.AppConfig, log *zap.Logger) (*Application, error) {
	if cfg.Auth.AccessSecret == "" || cfg.Auth.RefreshSecret == "" {
		return nil, errors.New("JWT secrets must be set")
	}

	r := NewRouter("production")

	_, err := storage.NewMinioStorage(cfg.Minio, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio storage: %w", err)
	}

	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("Successfully connected to the database!")

	authMw := middleware.Authenticate(cfg.Auth)

	authRepo := auth.NewRepository(database)
	authService := auth.NewService(authRepo, cfg.Auth)
	authHandler := auth.NewHandler(authService, authMw)
	r.RegisterRoutes(authHandler)

	return &Application{
		DB:     database,
		Router: r.Build(),
	}, nil
}
