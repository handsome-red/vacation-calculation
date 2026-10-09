package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/di"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/logger"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/router"
)

func Run(ctx context.Context, cfg config.Config) error {
	// Создаем логгер с настройками из конфига
	slogLogger := logger.NewLogger(cfg.LoggerConfig())

	// Поднимаем до интерфейс, чтобы application и domain не знали о реализации
	var log ports.Logger = slogLogger

	log.Info(ctx, "application starting")

	container, err := di.NewContainer(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("container: %w", err)
	}
	defer container.Close()

	r := routes.NewRouter(container, log)

	srv := http.Server{
		Addr:              cfg.Addr,
		Handler:           r.Build(),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info(ctx, "http server listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info(ctx, "shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	log.Info(ctx, "application stopped")

	return nil
}
