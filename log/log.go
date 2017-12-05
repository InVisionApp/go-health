package log

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

type mockLogger struct{}

// NewMockLogger creates a noop logger that is used internally by the health pkg
// when the user has not supplied their own logger.
func NewMockLogger() *mockLogger {
	return &mockLogger{}
}

func (m *mockLogger) Debug(msg string, args map[string]interface{}) {}
func (m *mockLogger) Info(msg string, args map[string]interface{})  {}
func (m *mockLogger) Warn(msg string, args map[string]interface{})  {}
func (m *mockLogger) Error(msg string, args map[string]interface{}) {}
