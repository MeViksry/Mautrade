package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MeViksry/Mautrade/backend/internal/config"
	"github.com/MeViksry/Mautrade/backend/internal/httpapi"
	"github.com/MeViksry/Mautrade/backend/internal/platform/postgres"
	"github.com/MeViksry/Mautrade/backend/internal/platform/queue"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Connect(rootCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	if db != nil {
		defer db.Close()
	}

	natsClient, err := queue.Connect(rootCtx, cfg.NATSURL)
	if err != nil {
		logger.Warn("nats unavailable; api starts without queue publisher", "error", err)
	} else {
		defer natsClient.Close()
	}

	api, err := httpapi.NewServer(cfg, db, natsClient, logger)
	if err != nil {
		logger.Error("create api", "error", err)
		os.Exit(1)
	}
	stopExecutionResults, err := api.StartExecutionResultConsumer(rootCtx)
	if err != nil {
		logger.Warn("execution result consumer unavailable", "error", err)
	} else {
		defer stopExecutionResults()
	}

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: api.Handler(),
	}

	go func() {
		logger.Info("mautrade api listening", "addr", cfg.HTTPAddr, "env", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	<-rootCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("mautrade api stopped")
}
