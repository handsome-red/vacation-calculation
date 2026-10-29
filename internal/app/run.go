package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/bootstrap_admin"
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

	if err := bootstrapAdmin(ctx, cfg, container, log); err != nil {
		return err
	}

	startVacationReminders(ctx, cfg, container, log)

	r, err := routes.NewRouter(container, log, nil)
	if err != nil {
		return fmt.Errorf("router: %w", err)
	}

	handler, err := r.Build()
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	srv := http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
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

// startVacationReminders запускает фоновую рассылку уведомлений о начале отпуска:
// сразу при старте и затем каждые REMINDER_INTERVAL, пока не отменён ctx.
func startVacationReminders(ctx context.Context, cfg config.Config, container *di.Container, log ports.Logger) {
	if container.SendRemindersUseCase == nil {
		log.Info(ctx, "vacation reminders disabled: SMTP_HOST, SMTP_USERNAME or SMTP_PASSWORD is not set")
		return
	}

	run := func() {
		res, err := container.SendRemindersUseCase.Handle(ctx, time.Now())
		if err != nil {
			log.Error(ctx, "vacation reminders failed", "error", err)
			return
		}
		if res.Sent > 0 || res.Failed > 0 {
			log.Info(ctx, "vacation reminders", "sent", res.Sent, "failed", res.Failed)
		}
	}

	log.Info(ctx, "vacation reminders enabled", "smtp_host", cfg.SMTPHost, "interval", cfg.ReminderInterval.String())
	go func() {
		run()
		ticker := time.NewTicker(cfg.ReminderInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}

// bootstrapAdmin создаёт первого администратора из ADMIN_*, если их ещё нет.
func bootstrapAdmin(ctx context.Context, cfg config.Config, container *di.Container, log ports.Logger) error {
	if cfg.AdminEmail == "" {
		return nil
	}

	hiredAt := cfg.AdminHiredAt
	if hiredAt == "" {
		hiredAt = time.Now().Format("2006-01-02")
	}

	res, err := container.BootstrapAdminUseCase.Handle(ctx, bootstrap_admin.Command{
		Email:          cfg.AdminEmail,
		Password:       cfg.AdminPassword,
		LastName:       cfg.AdminLastName,
		FirstName:      cfg.AdminFirstName,
		MiddleName:     cfg.AdminMiddleName,
		BirthDate:      cfg.AdminBirthDate,
		Position:       cfg.AdminPosition,
		HiredAt:        hiredAt,
		DistrictCode:   cfg.AdminDistrictCode,
		DepartmentCode: cfg.AdminDepartmentCode,
	})
	if err != nil {
		return fmt.Errorf("bootstrap admin: %w", err)
	}

	switch {
	case res.Created:
		log.Info(ctx, "admin created from ADMIN_* environment", "email", cfg.AdminEmail)
	case res.Promoted:
		log.Info(ctx, "existing user promoted to admin from ADMIN_EMAIL", "email", cfg.AdminEmail)
	default:
		log.Info(ctx, "admin already exists, ADMIN_* ignored")
	}
	return nil
}
