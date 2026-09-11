package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	_ "modernc.org/sqlite"
)

type Config struct {
	Path string
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		c.Path,
	)
}

func NewDB(ctx context.Context, cfg Config, log ports.Logger) (*sql.DB, error) {
	log.Info(ctx, "connecting to database", "path", cfg.Path)

	db, err := sql.Open("sqlite", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}

	log.Info(ctx, "connected to database", "path", cfg.Path)

	return db, nil

}

// func ()
