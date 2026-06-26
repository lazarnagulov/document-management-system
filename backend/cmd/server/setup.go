package main

import (
	"database/sql"
	"net/http"

	"go.uber.org/zap"
)

type Application struct {
	DB     *sql.DB
	Router http.Handler
}

func setupDependencies(log *zap.Logger) (*Application, error) {
	r := NewRouter("production")

	return &Application{
		DB:     nil,
		Router: r.Build(),
	}, nil
}
