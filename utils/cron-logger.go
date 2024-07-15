package utils

import (
	"github.com/charmbracelet/log"
	"github.com/discord-plays/bigben/logger"
	"github.com/robfig/cron/v3"
)

type cronLogger struct {
	logger *log.Logger
}

func (c *cronLogger) Info(msg string, keysAndValues ...interface{}) {}

func (c *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, "error", err)
	c.logger.Error(msg, keysAndValues)
}

var _ cron.Logger = (*cronLogger)(nil)

func WithCronLogger() cron.Option {
	return cron.WithLogger(&cronLogger{logger.Logger})
}
