package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vacation-calculation/internal/app"
	"vacation-calculation/internal/config"
	"vacation-calculation/internal/logger"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LoggerConfig())
	slog.SetDefault(log)

	log.Info("Starting application",
		"addr", cfg.Addr,
		"db_path", cfg.DatabasePath,
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
		"shutdown_timeout", cfg.ShutdownTimeout,
		"timestamp", time.Now().UTC(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Info("Received signal", "signal", sig)
		cancel()
	}()

	if err := app.Run(ctx, cfg, log); err != nil {
		log.Error("Application terminated with error",
			"error", err,
			"timestamp", time.Now(),
		)
		os.Exit(1)
	}

	log.Info("Application finished successfully")
}
