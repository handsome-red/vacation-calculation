package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/handsome-red/vacation-calculation/internal/app"
	"github.com/handsome-red/vacation-calculation/internal/config"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Создаем контекст, который отменится по SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Запускаем приложение
	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("application failed: %v", err)
	}
}
