package log

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

var (
	logConfig *LogConfig
	once      sync.Once
	logger    *zap.SugaredLogger
)

type LogConfig struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`
}

func init() {
	once.Do(func() {
		logConfig = &LogConfig{}
		envconfig.Process("", logConfig)
	})

	var l *zap.Logger

	switch logConfig.Level {
	case "debug":
		l, _ = zap.NewDevelopment()
	default:
		l, _ = zap.NewProduction()
	}

	logger = l.Sugar()
}

// With adds structured fields and returns a new logger instance
func With(fields ...any) *zap.SugaredLogger {
	return logger.With(fields...)
}

func WithError(err error) *zap.SugaredLogger {
	return logger.With("error", err)
}

// Debug logs at Debug level
func Debug(msg string) {
	logger.Debug(msg)
}

// Info logs at Info level
func Info(msg string) {
	logger.Info(msg)
}

// Error logs at Error level
func Error(msg string) {
	logger.Error(msg)
}

// Fatal logs at Fatal level and exits
func Fatal(msg string) {
	logger.Fatal(msg)
}
