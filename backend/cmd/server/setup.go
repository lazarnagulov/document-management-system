package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/auth"
	"bitbucket.org/HyperEther/document-management-system/internal/config"
	"bitbucket.org/HyperEther/document-management-system/internal/document"
	"bitbucket.org/HyperEther/document-management-system/internal/extractor"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/chunker"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/db"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/middleware"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/queue"
	"bitbucket.org/HyperEther/document-management-system/internal/processing"
	"bitbucket.org/HyperEther/document-management-system/internal/storage"
	"go.uber.org/zap"
)

type Application struct {
	DB     *sql.DB
	Router http.Handler
}

func setupDependencies(ctx context.Context, cfg *config.AppConfig, log *zap.Logger) (*Application, error) {
	if cfg.Auth.AccessSecret == "" || cfg.Auth.RefreshSecret == "" {
		return nil, errors.New("JWT secrets must be set")
	}

	r := NewRouter("production")

	minio, err := storage.NewMinioStorage(cfg.Minio)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio storage: %w", err)
	}

	log.Info("Successfully connected to MinIO",
		zap.String("endpoint", cfg.Minio.Endpoint),
		zap.String("bucket", cfg.Minio.BucketName),
	)

	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("Successfully connected to the database!")

	q := queue.NewMemoryQueue(100)

	authMw := middleware.Authenticate(cfg.Auth)

	authRepo := auth.NewRepository(database)
	authService := auth.NewService(authRepo, cfg.Auth)
	authHandler := auth.NewHandler(authService, authMw)

	docRepo := document.NewRepository(database)
	docService := document.NewService(docRepo, minio, q)
	docHandler := document.NewHandler(docService, authMw)

	r.RegisterRoutes(authHandler, docHandler)

	chunker := chunker.Default()
	extractorFactory := extractor.NewFactory(cfg.Extractor)
	processor := processing.NewProcessor(docRepo, minio, extractorFactory, chunker, log)
	worker := processing.NewWorker(q, processor)
	worker.Start(ctx)

	return &Application{
		DB:     database,
		Router: r.Build(),
	}, nil
}
