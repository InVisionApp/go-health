package health

import (
	"github.com/sirupsen/logrus"
)

type shim struct {
	logger *logrus.Entry
}

func NewLoggerLogrus(logger *logrus.Entry) ILogger {
	if logger == nil {
		logger = logrus.WithField("pkg", "health")
	}

	return &shim{logger: logger}
}

func (l *shim) Debug(msg string, args map[string]interface{}) {
	l.logger.WithFields(args).Warn(msg)
}

func (l *shim) Info(msg string, args map[string]interface{}) {
	l.logger.WithFields(args).Info(msg)
}

func (l *shim) Warn(msg string, args map[string]interface{}) {
	l.logger.WithFields(args).Warn(msg)
}

func (l *shim) Error(msg string, args map[string]interface{}) {
	l.logger.WithFields(args).Error(msg)
}
