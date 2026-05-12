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

	backend "pulse-check-backend/internal"
	utils "pulse-check-backend/internal/utils"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := backend.Config{
		Addr:            utils.EnvString("HTTP_ADDR", ":8080"),
		ServiceName:     utils.EnvString("OTEL_SERVICE_NAME", "pulse-check-backend"),
		ShutdownTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tracerProvider, err := utils.InitTracerProvider(ctx, cfg.ServiceName)
	if err != nil {
		logger.Error("init tracer provider failed", slog.Any("error", err))
		os.Exit(1)
	}

	app := backend.NewApp(logger)
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server started", slog.String("addr", cfg.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			logger.Error("http server failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	app.SetReady(false)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	var shutdownErr error

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("error", err))
		shutdownErr = errors.Join(shutdownErr, err)
	}

	if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
		logger.Error("tracer provider shutdown failed", slog.Any("error", err))
		shutdownErr = errors.Join(shutdownErr, err)
	}

	if err := <-errCh; err != nil {
		logger.Error("http server stopped with error", slog.Any("error", err))
		shutdownErr = errors.Join(shutdownErr, err)
	}

	if shutdownErr != nil {
		os.Exit(1)
	}

	logger.Info("http server stopped")
}