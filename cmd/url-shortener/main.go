package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Tbits007/url-shortener/internal/config"
	"github.com/Tbits007/url-shortener/internal/lib/logger/sl"
	"github.com/Tbits007/url-shortener/internal/storage/postgres"
)

const (
    envLocal = "local"
    envDev   = "dev"
    envProd  = "prod"
)

func main() {
	// Config
	cfg := config.MustLoad()

	// Logger
	log := setupLogger(cfg.Env)
    log = log.With(slog.String("env", cfg.Env))

    log.Info("initializing server", slog.String("address", cfg.HTTPServer.Address))
    log.Debug("logger debug mode enabled")

	// Storage
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
        cfg.Postgres.Password,
        cfg.Postgres.Host,
        cfg.Postgres.Port,
        cfg.Postgres.DBName,
	)
	storage, err := postgres.New(connStr)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	_ = storage
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)		
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)			
	}

	return log 
}