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
	database, err := db.NewPostgresConnection(cfg.DB, log)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	authHandler := auth.NewHandler()
	r.RegisterRoutes(authHandler)

	return &Application{
		DB:     database,
		Router: r.Build(),
	}, nil
}
