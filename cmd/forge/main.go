package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yashlunawat/forge/internal/config"
	"github.com/yashlunawat/forge/internal/database"
	"github.com/yashlunawat/forge/internal/server"
	"github.com/yashlunawat/forge/internal/store"
	"github.com/yashlunawat/forge/internal/store/memory"
	"github.com/yashlunawat/forge/internal/store/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.Environment)
	st, closeStore, err := loadStore(context.Background(), logger, cfg)
	if err != nil {
		logger.Error("load store", "error", err)
		os.Exit(1)
	}
	defer closeStore()

	app, err := server.New(cfg, logger, st)
	if err != nil {
		logger.Error("initialize server", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              cfg.Address,
		Handler:           app.Router(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("forge listening", "addr", cfg.Address, "env", cfg.Environment)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func newLogger(environment string) *slog.Logger {
	level := slog.LevelInfo
	if environment == "development" {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}

func loadStore(ctx context.Context, logger *slog.Logger, cfg config.Config) (store.Store, func(), error) {
	if cfg.DatabaseURL == "" {
		logger.Info("using in-memory store")
		return memory.NewStore(), func() {}, nil
	}

	db, err := database.OpenPostgres(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	if err := database.Migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	logger.Info("using postgres store")
	return postgres.NewStore(db), func() {
		if err := db.Close(); err != nil {
			logger.Error("close database", "error", err)
		}
	}, nil
}
