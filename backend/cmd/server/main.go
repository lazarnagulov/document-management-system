// @title           Document Management System API
// @version         1.0
// @description     DMS REST API
// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey  AccessTokenCookie
// @in                          cookie
// @name                        access_token
// @description                 Short-lived JWT access token. Set automatically by /auth/login and /auth/refresh.

// @securityDefinitions.apikey  RefreshTokenCookie
// @in                          cookie
// @name                        refresh_token
// @description                 Long-lived JWT refresh token. HttpOnly, scoped to /api/v1/auth/refresh only.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/config"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log, err := logger.New(cfg.Env, cfg.AppRoot)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := setupDependencies(ctx, cfg, log)
	if err != nil {
		log.Fatal("failed to setup dependencies", zap.Error(err))
	}
	defer app.DB.Close()

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: app.Router,
	}

	log.Info("starting server",
		zap.String("port", cfg.Port),
		zap.String("environment", cfg.Env),
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	log.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}

}
