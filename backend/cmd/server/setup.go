package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/auth"
	"bitbucket.org/HyperEther/document-management-system/internal/config"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/db"
	"go.uber.org/zap"
)

type Application struct {
	DB     *sql.DB
	Router http.Handler
}

func setupDependencies(cfg *config.AppConfig, log *zap.Logger) (*Application, error) {
	r := NewRouter("production")
	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("Successfully connected to the database!")

	authRepo := auth.NewRepository(database)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)
	r.RegisterRoutes(authHandler)

	return &Application{
		DB:     database,
		Router: r.Build(),
	}, nil
}
