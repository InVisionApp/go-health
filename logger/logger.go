package logger

import (
	"fmt"
	"log"
)

//go:generate counterfeiter -o fakes/ilogger.go . ILogger

// ILogger interface allows you to use a custom logger. Since the `logger` pkg does
// not expose an interface for a logger (and there is no "accepted" interface),
// we roll our own and supplement it with some helpers/shims for common logging
// libraries such as `logrus`. See [DOC.md(DOCS.md#Logging).
type ILogger interface {
	Debug(msg string, args map[string]interface{})
	Info(msg string, args map[string]interface{})
	Warn(msg string, args map[string]interface{})
	Error(msg string, args map[string]interface{})
}

type basic struct{}

// NewBasic creates a simple logger that is used internally by the health pkg
// when the user has not supplied their own logger.
func NewBasic() *basic {
	return &basic{}
}

func (m *basic) Debug(msg string, args map[string]interface{}) {
	log.Printf("[DEBUG] %s [%s]\n", msg, pretty(args))
}

func (m *basic) Info(msg string, args map[string]interface{}) {
	log.Printf("[INFO] %s [%s]\n", msg, pretty(args))
}

func (m *basic) Warn(msg string, args map[string]interface{}) {
	log.Printf("[WARN] %s [%s]\n", msg, pretty(args))
}

func (m *basic) Error(msg string, args map[string]interface{}) {
	log.Printf("[ERROR] %s [%s]\n", msg, pretty(args))
}

func pretty(m map[string]interface{}) string {
	s := ""
	for k, v := range m {
		s += fmt.Sprintf("%s=%v ", k, v)
	}

	return s[:len(s)-1]
}

type noop struct{}

// NewNoop creates a noop logger that can be used to silence all logging from this library.
func NewNoop() *noop {
	return &noop{}
}

func (m *noop) Debug(msg string, args map[string]interface{}) {}
func (m *noop) Info(msg string, args map[string]interface{})  {}
func (m *noop) Warn(msg string, args map[string]interface{})  {}
func (m *noop) Error(msg string, args map[string]interface{}) {}
