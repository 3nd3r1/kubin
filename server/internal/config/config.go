package config

import (
	"sync"

	"github.com/3nd3r1/kubin/server/internal/log"
	"github.com/kelseyhightower/envconfig"
)

var (
	appConfig *AppConfig
	once      sync.Once
)

type AppConfig struct {
	Server struct {
		Port         int    `envconfig:"SERVER_PORT" default:"8080"`
		ReadTimeout  int    `envconfig:"SERVER_READ_TIMEOUT" default:"10"`
		WriteTimeout int    `envconfig:"SERVER_WRITE_TIMEOUT" default:"10"`
		IdleTimeout  int    `envconfig:"SERVER_IDLE_TIMEOUT" default:"10"`
		Env          string `envconfig:"SERVER_ENV" default:"development"`
	} `split_words:"true"`
}

func init() {
	once.Do(func() {
		appConfig = &AppConfig{}
		readEnv(appConfig)
	})
}

func readEnv(cfg *AppConfig) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.WithError(err).Error("Failed parsing envVars, using default values")
	}
}

func Get() *AppConfig {
	return appConfig
}
