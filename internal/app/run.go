package app

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/handler"
	"github.com/handsome-red/vacation-calculation/internal/repository"
	"github.com/handsome-red/vacation-calculation/internal/service"
)

func Run(ctx context.Context, cfg config.Config, log *slog.Logger) error {

	log.Info("Loading templates", "path", cfg.TemplatesPath)
	templates, err := loadTemplates(cfg.TemplatesPath)
	if err != nil {
		return fmt.Errorf("loading templates: %w", err)
	}

	log.Info("Connecting to database", "path", cfg.DatabasePath)
	repo, err := repository.NewSQLiteRepository(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer repo.Close()

	serv := service.NewVacationService(repo)
	hand := handler.NewHandler(serv, templates)

	mux := NewRouter(hand, log)

	srv := http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Info("Starting HTTP server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server failed: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("Shutting down server gracefully")

		shutdownCtx, timeoutCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer timeoutCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("Server shutdown failed", "error", err)
			return fmt.Errorf("server shuting failed: %w", err)
		}
	}

	log.Info("Server shutdown complete successfully")

	return nil
}

// app/exe/main.go
func loadTemplates(path string) (*template.Template, error) {
	// Просто загружаем все файлы из одной папки
	return template.ParseGlob(path + "/*.html")
}
