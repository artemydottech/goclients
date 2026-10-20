package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/artemydottech/goclients/internal/app"
	"github.com/artemydottech/goclients/internal/config"
	"github.com/artemydottech/goclients/internal/middleware"
	"github.com/artemydottech/goclients/internal/repository"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info(".env config not found")
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid config", "err", err)
		os.Exit(1)
	}

	slog.SetDefault(slog.New(middleware.NewContextHandler(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel}),
	)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := repository.Open(ctx, cfg.DBPath)
	if err != nil {
		slog.Error("db connection failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := repository.Migrate(ctx, db); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	handler := app.New(db)

	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: middleware.Chain(handler,
			middleware.RequestID,
			middleware.AccessLog(slog.Default()),
			middleware.Recover,
			middleware.SecureHeaders,
			middleware.BodyLimit(cfg.MaxBodyBytes),
		),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown failed", "err", err)
	}
	slog.Info("server stopped")
}
