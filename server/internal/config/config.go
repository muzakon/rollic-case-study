package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Env string

const (
	EnvDevelopment Env = "development"
	EnvProduction  Env = "production"
	EnvTesting     Env = "testing"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Name string `env:"APP_NAME" env-default:"leaderboard-api"`
	Env  Env    `env:"APP_ENV" env-default:"development"`
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT" env-default:"8080"`
}

type DatabaseConfig struct {
	Host            string `env:"DB_HOST" env-default:"localhost"`
	Port            string `env:"DB_PORT" env-default:"5432"`
	User            string `env:"DB_USER" env-required:"true"`
	Password        string `env:"DB_PASSWORD" env-required:"true"`
	Name            string `env:"DB_NAME" env-required:"true"`
	SSLMode         string `env:"DB_SSL_MODE" env-default:"disable"`
	MaxOpenConns    int    `env:"DB_MAX_OPEN_CONNS" env-default:"100"`
	MaxIdleConns    int    `env:"DB_MAX_IDLE_CONNS" env-default:"20"`
	ConnMaxLifetime int    `env:"DB_CONN_MAX_LIFETIME_MINUTES" env-default:"15"`
}

func LoadConfig() *Config {
	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			log.Fatalf("Config error: %s", err)
		}
	}

	return &cfg
}
