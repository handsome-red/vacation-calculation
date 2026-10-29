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

	// Почта администратора для уведомлений сотрудникам о начале отпуска.
	// Пустой SMTP_HOST, SMTP_USERNAME или SMTP_PASSWORD отключает рассылку.
	SMTPHost     string `envconfig:"SMTP_HOST"`
	SMTPPort     int    `envconfig:"SMTP_PORT" default:"465"`
	SMTPUsername string `envconfig:"SMTP_USERNAME"`
	SMTPPassword string `envconfig:"SMTP_PASSWORD"`
	SMTPFrom     string `envconfig:"SMTP_FROM"` // по умолчанию SMTP_USERNAME
	SMTPTLS      string `envconfig:"SMTP_TLS"`  // ssl | starttls | none; по умолчанию по порту
	SMTPFromName string `envconfig:"SMTP_FROM_NAME" default:"Калькулятор отпусков"`
	// Адреса отдела кадров для копии уведомления, через запятую; по умолчанию —
	// ящик администратора (SMTP_FROM или SMTP_USERNAME).
	HREmail string `envconfig:"HR_EMAIL"`
	// Как часто проверять, кому пора отправить уведомление.
	ReminderInterval time.Duration `envconfig:"REMINDER_INTERVAL" default:"1h"`

	// Первый администратор: создаётся при старте, если в БД нет ни одного.
	// Пустой ADMIN_EMAIL отключает создание.
	AdminEmail          string `envconfig:"ADMIN_EMAIL"`
	AdminPassword       string `envconfig:"ADMIN_PASSWORD"`
	AdminLastName       string `envconfig:"ADMIN_LAST_NAME" default:"Администратор"`
	AdminFirstName      string `envconfig:"ADMIN_FIRST_NAME" default:"Системный"`
	AdminMiddleName     string `envconfig:"ADMIN_MIDDLE_NAME"`
	AdminBirthDate      string `envconfig:"ADMIN_BIRTH_DATE" default:"1970-01-01"`
	AdminPosition       string `envconfig:"ADMIN_POSITION" default:"software_engineer"`
	AdminHiredAt        string `envconfig:"ADMIN_HIRED_AT"` // по умолчанию — дата первого запуска
	AdminDistrictCode   string `envconfig:"ADMIN_DISTRICT" default:"kazan"`
	AdminDepartmentCode string `envconfig:"ADMIN_DEPARTMENT" default:"epidemiological_supervision"`
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
