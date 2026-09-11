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

// func Run(ctx context.Context, cfg config.Config, log *slog.Logger) error {

// 	log.Info("Loading templates", "path", cfg.TemplatesPath)
// 	templates, err := loadTemplates(cfg.TemplatesPath)
// 	if err != nil {
// 		return fmt.Errorf("loading templates: %w", err)
// 	}

// 	log.Info("Connecting to database", "path", cfg.DatabasePath)
// 	repo, err := repository.NewSQLiteRepository(cfg.DatabasePath)
// 	if err != nil {
// 		return fmt.Errorf("connecting to database: %w", err)
// 	}
// 	defer repo.Close()

// 	serv := service.NewVacationService(repo)
// 	hand := handler.NewHandler(serv, templates)

// 	mux := NewRouter(hand, log)

// 	srv := http.Server{
// 		Addr:              cfg.Addr,
// 		Handler:           mux,
// 		ReadTimeout:       5 * time.Second,
// 		WriteTimeout:      10 * time.Second,
// 		IdleTimeout:       30 * time.Second,
// 		ReadHeaderTimeout: 2 * time.Second,
// 	}

// 	errCh := make(chan error, 1)

// 	go func() {
// 		log.Info("Starting HTTP server", "address", srv.Addr)
// 		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
// 			errCh <- fmt.Errorf("server failed: %w", err)
// 		}
// 		close(errCh)
// 	}()

// 	select {
// 	case err := <-errCh:
// 		return err
// 	case <-ctx.Done():
// 		log.Info("Shutting down server gracefully")

// 		shutdownCtx, timeoutCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
// 		defer timeoutCancel()

// 		if err := srv.Shutdown(shutdownCtx); err != nil {
// 			log.Error("Server shutdown failed", "error", err)
// 			return fmt.Errorf("server shuting failed: %w", err)
// 		}
// 	}

// 	log.Info("Server shutdown complete successfully")

// 	return nil
// }
