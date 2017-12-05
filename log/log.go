package log

import (
	"log"
	"fmt"
)

//go:generate counterfeiter -o fakes/ilogger.go . ILogger

// ILogger interface allows you to use a custom logger. Since the `log` pkg does
// not expose an interface for a logger (and there is no "accepted" interface),
// we roll our own and supplement it with some helpers/shims for common logging
// libraries such as `logrus`. See [DOC.md(DOCS.md#Logging).
type ILogger interface {
	Debug(msg string, args map[string]interface{})
	Info(msg string, args map[string]interface{})
	Warn(msg string, args map[string]interface{})
	Error(msg string, args map[string]interface{})
}

type defaultLogger struct{}

// NewDefaultLogger creates a simple logger that is used internally by the health pkg
// when the user has not supplied their own logger.
func NewDefaultLogger() *defaultLogger {
	return &defaultLogger{}
}

func (m *defaultLogger) Debug(msg string, args map[string]interface{}) {
	log.Printf("[DEBUG] %s [%s]\n", msg, pretty(args))
}

func (m *defaultLogger) Info(msg string, args map[string]interface{})  {
	log.Printf("[INFO] %s [%s]\n", msg, pretty(args))
}

func (m *defaultLogger) Warn(msg string, args map[string]interface{})  {
	log.Printf("[WARN] %s [%s]\n", msg, pretty(args))
}

func (m *defaultLogger) Error(msg string, args map[string]interface{}) {
	log.Printf("[ERROR] %s [%s]\n", msg, pretty(args))
}

func pretty(m map[string]interface{}) string {
	s := ""
	for k, v := range m {
		s += fmt.Sprintf("%s=%v ", k, v)
	}

	return s[:len(s)-1]
}

type mockLogger struct{}

// NewMockLogger creates a noop logger that can be used to silence all logging from this library.
func NewMockLogger() *mockLogger {
	return &mockLogger{}
}

func (m *mockLogger) Debug(msg string, args map[string]interface{}) {}
func (m *mockLogger) Info(msg string, args map[string]interface{})  {}
func (m *mockLogger) Warn(msg string, args map[string]interface{})  {}
func (m *mockLogger) Error(msg string, args map[string]interface{}) {}
