package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const DotEnvFilename = ".env"

type HTTPServerConfig struct {
	ListenAddr        string        `envconfig:"API_LISTEN_ADDR" default:":8080"`
	IdleTimeout       time.Duration `envconfig:"IDLE_TIMEOUT" default:"60s"`
	ReadHeaderTimeout time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"10s"`
}

type DatabaseConfig struct {
	DSN string `envconfig:"DB_POSTGRES_DSN" default:"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"`
}

type Config struct {
	ServiceName              string `envconfig:"SERVICE_NAME"`
	ReleaseID                string `envconfig:"RELEASE_ID"`
	LogLevel                 string `envconfig:"LOG_LEVEL" default:"debug"`
	MaxLogLen                int    `envconfig:"MAX_LOG_LEN" default:"2000"`
	RequestsReviewExpireDays int    `envconfig:"REQUESTS_REVIEW_EXPIRE_DAYS" default:"14"`
	HTTP                     HTTPServerConfig
	Database                 DatabaseConfig
}

func NewConfigFromEnv() (*Config, error) {
	cfg := &Config{}
	_ = godotenv.Load(DotEnvFilename)
	if err := envconfig.Process("", cfg); err != nil {
		return nil, errors.Wrap(err, "failed to process env vars")
	}
	return cfg, nil
}
