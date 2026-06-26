package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("warning: .env file not found")
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	log, err := logger.New(env)
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer log.Sync()

	app, err := setupDependencies(log)
	if err != nil {
		log.Fatal("Failed to setup dependencies", zap.Error(err))
	}
	// defer app.DB.Close()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: app.Router,
	}

	log.Info("starting server",
		zap.String("port", "8080"),
		zap.String("environment", env),
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}

}
