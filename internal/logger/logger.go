package logger

import (
	"io"
	"log/slog"
	"os"
)

type Config struct {
	Level  string
	Format string
}

func New(cfg Config) *slog.Logger {
	return NewWithOutput(cfg, os.Stdin)
}

func NewWithOutput(cfg Config, output io.Writer) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	// Выбираем формат
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}
