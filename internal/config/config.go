package config

import (
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/infrastructure/logger"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr              string        `envconfig:"ADDR" default:":8080"`
	ReadTimeout       time.Duration `envconfig:"READ_TIMEOUT" default:"5s"`
	WriteTimeout      time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
	IdleTimeout       time.Duration `envconfig:"IDLE_TIMEOUT" default:"30s"`
	ReadHeaderTimeout time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"2s"`
	ShutdownTimeout   time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s"`
	DatabasePath      string        `envconfig:"DB_PATH" default:"./data/vacations.db"`
	DatabaseMaxConn   int           `envconfig:"DB_MAX_CONN" default:"10"`
	TemplatesPath     string        `envconfig:"TEMPLATES_PATH" default:"web/templates"`
	StaticPath        string        `envconfig:"STATIC_PATH" default:"web/static"`
	LogLevel          string        `envconfig:"LOG_LEVEL" default:"info"`
	LogFormat         string        `envconfig:"LOG_FORMAT" default:"text"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("loading config: %w", err)
	}

	return cfg, nil
}

func (c Config) LoggerConfig() logger.Config {
	return logger.Config{
		Level:  c.LogLevel,
		Format: c.LogFormat,
	}
}
