// Package config reads the server configuration from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port         int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"10s"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout  time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
	Env          string        `envconfig:"SERVER_ENV" default:"development"`
	LogLevel     string        `envconfig:"LOG_LEVEL" default:"info"`
	// DataDir holds the SQLite database and the blob directory.
	DataDir string `envconfig:"SERVER_DATA_DIR" default:"data"`
	// PublicURL is the base for links handed back to the CLI, like
	// https://kubin.example.com. Empty means taken from the request.
	PublicURL      string `envconfig:"SERVER_PUBLIC_URL"`
	MaxUploadBytes int64  `envconfig:"SERVER_MAX_UPLOAD_BYTES" default:"67108864"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}

func (c Config) Production() bool {
	return c.Env == "production"
}
