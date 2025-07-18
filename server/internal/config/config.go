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
