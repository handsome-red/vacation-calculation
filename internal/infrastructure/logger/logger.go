package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type Config struct {
	Level  string
	Format string
	Output io.Writer
}

type SlogLogger struct {
	l *slog.Logger
}

func NewLogger(config Config) *SlogLogger {

	var level slog.Level

	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	opts.Level = level

	var handler slog.Handler

	out := config.Output
	if out == nil {
		out = os.Stdout
	}

	if config.Format == "json" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return &SlogLogger{l: slog.New(handler)}
}

func (s *SlogLogger) Debug(ctx context.Context, msg string, attrs ...any) {
	s.l.DebugContext(ctx, msg, attrs...)
}

func (s *SlogLogger) Info(ctx context.Context, msg string, attrs ...any) {
	s.l.InfoContext(ctx, msg, attrs...)
}

func (s *SlogLogger) Warn(ctx context.Context, msg string, attrs ...any) {
	s.l.WarnContext(ctx, msg, attrs...)
}

func (s *SlogLogger) Error(ctx context.Context, msg string, attrs ...any) {
	s.l.ErrorContext(ctx, msg, attrs...)
}
